package service

import (
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"amiya-eden/pkg/xlsx"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

const ATScoreSourceURL = "https://docs.google.com/spreadsheets/d/1AVYlWlvuMKnA3yuqqDCcAkia8pvhpb9OBcM29WFw5rM"

var ATDuplicatePenalties = map[string]int{"FR": 0, "DE": 1, "CR": 2, "BC": 3, "BS": 4, "IN": 2, "LF": 1}

type ATScoreService struct {
	repo    *repository.ATScoreRepository
	sdeRepo *repository.SdeRepository
}

func NewATScoreService() *ATScoreService {
	return &ATScoreService{repo: repository.NewATScoreRepository(), sdeRepo: repository.NewSdeRepository()}
}

type ATScoreRuleDetail struct {
	Rule      model.ATScoreRuleVersion        `json:"rule"`
	Ships     []model.ATScoreShip             `json:"ships"`
	Banned    []model.ATScoreBannedShip       `json:"banned"`
	Penalties []model.ATScoreDuplicatePenalty `json:"penalties"`
}

type atScoreImportShip struct {
	Name    string
	Points  int
	Size    string
	Penalty int
}

func (s *ATScoreService) Import(userID uint, data []byte, sourceURL string) (*ATScoreRuleDetail, error) {
	if len(data) == 0 {
		return nil, errors.New("XLSX 文件为空")
	}
	if len(data) > 20*1024*1024 {
		return nil, errors.New("XLSX 文件不能超过 20 MiB")
	}
	book, err := xlsx.Open(data)
	if err != nil {
		return nil, fmt.Errorf("读取 XLSX 失败: %w", err)
	}
	if len(book.Sheets) > 20 {
		return nil, errors.New("XLSX 工作表不能超过 20 个")
	}
	version, parsed, bannedNames, err := parseATScoreWorkbook(book)
	if err != nil {
		return nil, err
	}
	if len(parsed) == 0 {
		return nil, errors.New("积分表没有有效舰船")
	}

	allNames := make([]string, 0, len(parsed)+len(bannedNames))
	for _, ship := range parsed {
		allNames = append(allNames, ship.Name)
	}
	allNames = append(allNames, bannedNames...)
	typeIDs, err := s.sdeRepo.GetTypeIDsByNames(allNames)
	if err != nil {
		return nil, err
	}
	unknown := []string{}
	for _, name := range allNames {
		if typeIDs[name] == 0 {
			unknown = append(unknown, name)
		}
	}
	if len(unknown) > 0 {
		sort.Strings(unknown)
		return nil, fmt.Errorf("SDE 无法解析以下舰船: %s", strings.Join(unknown, ", "))
	}
	ships := make([]model.ATScoreShip, len(parsed))
	for i, ship := range parsed {
		ships[i] = model.ATScoreShip{TypeID: typeIDs[ship.Name], TypeName: ship.Name, BasePoints: ship.Points, SizeCode: ship.Size}
	}
	banned := make([]model.ATScoreBannedShip, len(bannedNames))
	for i, name := range bannedNames {
		banned[i] = model.ATScoreBannedShip{TypeID: typeIDs[name], TypeName: name}
	}
	penaltyBySize := make(map[string]int)
	for _, ship := range parsed {
		penaltyBySize[ship.Size] = ship.Penalty
	}
	penalties := make([]model.ATScoreDuplicatePenalty, 0, len(penaltyBySize))
	for size, value := range penaltyBySize {
		penalties = append(penalties, model.ATScoreDuplicatePenalty{SizeCode: size, Penalty: value})
	}
	sort.Slice(penalties, func(i, j int) bool { return penalties[i].SizeCode < penalties[j].SizeCode })
	hash := sha256.Sum256(data)
	if sourceURL == "" {
		sourceURL = ATScoreSourceURL
	}
	rule := &model.ATScoreRuleVersion{Version: version, SourceURL: sourceURL, FileHash: hex.EncodeToString(hash[:]), Status: model.ATScoreRuleDraft, CreatedBy: userID}
	if err := s.repo.CreateRule(rule, ships, banned, penalties); err != nil {
		return nil, err
	}
	return &ATScoreRuleDetail{Rule: *rule, Ships: ships, Banned: banned, Penalties: penalties}, nil
}

func parseATScoreWorkbook(book *xlsx.Workbook) (string, []atScoreImportShip, []string, error) {
	if book.Sheets["New Static Values (Do Not Edit)"] != nil {
		return parseATXXIIV14Workbook(book)
	}
	if book.Sheets["Ships"] != nil && book.Sheets["Version History"] != nil {
		return parseLegacyATScoreWorkbook(book)
	}
	return "", nil, nil, errors.New("不支持的积分表：需要 ATXXII V1.4 的 New Static Values (Do Not Edit) 工作表")
}

func parseATXXIIV14Workbook(book *xlsx.Workbook) (string, []atScoreImportShip, []string, error) {
	version := ""
	for row := 1; row <= 200; row++ {
		candidate := strings.TrimSpace(book.Cell("How to Edit This Sheet", fmt.Sprintf("A%d", row)))
		if strings.HasPrefix(strings.ToUpper(candidate), "V") {
			version = candidate
		}
	}
	if version == "" {
		return "", nil, nil, errors.New("ATXXII 积分表版本无效")
	}

	parsed := make([]atScoreImportShip, 0, 300)
	seen := map[string]bool{}
	emptyRows := 0
	for row := 4; row <= 1000; row++ {
		name := strings.TrimSpace(book.Cell("New Static Values (Do Not Edit)", fmt.Sprintf("F%d", row)))
		pointsText := book.Cell("New Static Values (Do Not Edit)", fmt.Sprintf("H%d", row))
		hullType := book.Cell("New Static Values (Do Not Edit)", fmt.Sprintf("I%d", row))
		inflation := book.Cell("New Static Values (Do Not Edit)", fmt.Sprintf("J%d", row))
		if strings.Contains(name+pointsText+hullType+inflation, "#") {
			return "", nil, nil, fmt.Errorf("New Static Values 第 %d 行包含公式错误", row)
		}
		if name == "" {
			emptyRows++
			if emptyRows > 50 {
				break
			}
			continue
		}
		emptyRows = 0
		if name == "Empty" || name == "Capsule" {
			continue
		}
		key := strings.ToLower(name)
		if seen[key] {
			return "", nil, nil, fmt.Errorf("舰船 %s 重复", name)
		}
		seen[key] = true
		points, err := parseATPoints(name, pointsText)
		if err != nil {
			return "", nil, nil, err
		}
		sizeCode, ok := atXXIIHullSizeCode(hullType)
		if !ok {
			return "", nil, nil, fmt.Errorf("舰船 %s 的 Hull Type 无效: %s", name, hullType)
		}
		inflationValue, err := parseATPoints(name+" Inflation Value", inflation)
		if err != nil {
			return "", nil, nil, err
		}
		if expected, exists := ATDuplicatePenalties[sizeCode]; !exists || inflationValue != expected {
			sizeCode = fmt.Sprintf("%s%d", sizeCode, inflationValue)
		}
		parsed = append(parsed, atScoreImportShip{Name: name, Points: points, Size: sizeCode, Penalty: inflationValue})
	}
	return version, parsed, nil, nil
}

func parseLegacyATScoreWorkbook(book *xlsx.Workbook) (string, []atScoreImportShip, []string, error) {
	version := book.Cell("Version History", "B2")
	if version == "" {
		version = book.Cell("Comp Pick", "B1")
	}
	if version == "" || strings.Contains(version, "#") {
		return "", nil, nil, errors.New("XLSX 版本无效")
	}
	parsed := []atScoreImportShip{}
	bannedNames := []string{}
	seenShips, seenBanned := map[string]bool{}, map[string]bool{}
	emptyRows := 0
	for row := 3; row <= 5000; row++ {
		name := book.Cell("Ships", fmt.Sprintf("G%d", row))
		pointsText := book.Cell("Ships", fmt.Sprintf("I%d", row))
		size := strings.ToUpper(book.Cell("Ships", fmt.Sprintf("L%d", row)))
		banned := book.Cell("Ships", fmt.Sprintf("N%d", row))
		if strings.Contains(name+pointsText+size+banned, "#N/A") {
			return "", nil, nil, fmt.Errorf("Ships 第 %d 行包含 #N/A", row)
		}
		if banned != "" {
			key := strings.ToLower(banned)
			if seenBanned[key] {
				return "", nil, nil, fmt.Errorf("禁用船 %s 重复", banned)
			}
			seenBanned[key] = true
			bannedNames = append(bannedNames, banned)
		}
		if name == "" {
			emptyRows++
			if emptyRows > 100 {
				break
			}
			continue
		}
		emptyRows = 0
		if name == "Empty" || name == "Capsule" {
			continue
		}
		key := strings.ToLower(name)
		if seenShips[key] {
			return "", nil, nil, fmt.Errorf("舰船 %s 重复", name)
		}
		seenShips[key] = true
		points, err := parseATPoints(name, pointsText)
		if err != nil {
			return "", nil, nil, err
		}
		penalty := ATDuplicatePenalties[size]
		parsed = append(parsed, atScoreImportShip{Name: name, Points: points, Size: size, Penalty: penalty})
	}
	return version, parsed, bannedNames, nil
}

func parseATPoints(name, value string) (int, error) {
	pointsFloat, err := strconv.ParseFloat(value, 64)
	if err != nil || pointsFloat < 0 || math.Trunc(pointsFloat) != pointsFloat {
		return 0, fmt.Errorf("舰船 %s 的分值无效: %s", name, value)
	}
	return int(pointsFloat), nil
}

func atXXIIHullSizeCode(hullType string) (string, bool) {
	switch strings.TrimSpace(hullType) {
	case "Corvette":
		return "FR", true
	case "Frigate":
		return "FR", true
	case "Destroyer":
		return "DE", true
	case "Cruiser":
		return "CR", true
	case "Battlecruiser":
		return "BC", true
	case "Battleship":
		return "BS", true
	case "Industrial":
		return "IN", true
	case "Logistics Frigate":
		return "LF", true
	case "Logistics":
		return "LG", true
	default:
		return "", false
	}
}

func (s *ATScoreService) Publish(id uint) error                     { return s.repo.Publish(id) }
func (s *ATScoreService) List() ([]model.ATScoreRuleVersion, error) { return s.repo.ListRules() }
func (s *ATScoreService) Get(id uint) (*ATScoreRuleDetail, error) {
	rule, err := s.repo.GetRule(id)
	if err != nil {
		return nil, err
	}
	ships, err := s.repo.GetShips(id)
	if err != nil {
		return nil, err
	}
	banned, err := s.repo.GetBanned(id)
	if err != nil {
		return nil, err
	}
	penalties, err := s.repo.GetPenalties(id)
	if err != nil {
		return nil, err
	}
	return &ATScoreRuleDetail{Rule: *rule, Ships: ships, Banned: banned, Penalties: penalties}, nil
}

func (s *ATScoreService) ValidateRoster(ruleID uint, format string, typeIDs []int64) (*model.ATScoreRosterResult, error) {
	ships, err := s.repo.GetShips(ruleID)
	if err != nil {
		return nil, err
	}
	bannedShips, err := s.repo.GetBanned(ruleID)
	if err != nil {
		return nil, err
	}
	penalties, err := s.repo.GetPenalties(ruleID)
	if err != nil {
		return nil, err
	}
	return CalculateATRoster(format, typeIDs, ships, bannedShips, penalties)
}

// CalculateATRoster is the deterministic, database-free AT scoring core.
func CalculateATRoster(format string, typeIDs []int64, ships []model.ATScoreShip, bannedShips []model.ATScoreBannedShip, penalties []model.ATScoreDuplicatePenalty) (*model.ATScoreRosterResult, error) {
	if format != "7v7" && format != "10v10" {
		return nil, errors.New("赛制仅支持 7v7 或 10v10")
	}
	maxSlots := 7
	if format == "10v10" {
		maxSlots = 10
	}
	result := &model.ATScoreRosterResult{Format: format, MaxPoints: 200, Valid: true, Errors: []string{}}
	if len(typeIDs) > maxSlots {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("%s 最多 %d 个有效槽位", format, maxSlots))
	}
	shipMap := map[int64]model.ATScoreShip{}
	for _, ship := range ships {
		shipMap[ship.TypeID] = ship
	}
	banned := map[int64]bool{}
	for _, ship := range bannedShips {
		banned[ship.TypeID] = true
	}
	penaltyMap := map[string]int{}
	for _, penalty := range penalties {
		penaltyMap[penalty.SizeCode] = penalty.Penalty
	}
	counts := map[int64]int{}
	for _, typeID := range typeIDs {
		counts[typeID]++
	}
	for _, typeID := range typeIDs {
		ship, ok := shipMap[typeID]
		if !ok {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("TypeID %d 不在该规则版本中", typeID))
			continue
		}
		if banned[typeID] {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("%s 为禁用船", ship.TypeName))
		}
		penalty, ok := penaltyMap[ship.SizeCode]
		if !ok {
			penalty = penaltyMap["OTHER"]
		}
		effective := ship.BasePoints + (counts[typeID]-1)*penalty
		result.TotalPoints += effective
		result.Slots = append(result.Slots, model.ATScoreRosterSlot{TypeID: typeID, TypeName: ship.TypeName, BasePoints: ship.BasePoints, DuplicateCount: counts[typeID], EffectivePoints: effective})
	}
	if result.TotalPoints > 200 {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("总分 %d 超过 200", result.TotalPoints))
	}
	return result, nil
}
