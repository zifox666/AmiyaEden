package service

import (
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type SkillPlanService struct {
	planRepo *repository.SkillPlanRepository
	sdeRepo  *repository.SdeRepository
	charRepo *repository.EveCharacterRepository
}

func NewSkillPlanService() *SkillPlanService {
	return &SkillPlanService{
		planRepo: repository.NewSkillPlanRepository(),
		sdeRepo:  repository.NewSdeRepository(),
		charRepo: repository.NewEveCharacterRepository(),
	}
}

// ── 请求/响应 DTO ──

type CreateSkillPlanRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	SkillText   string `json:"skill_text" binding:"required"`
}

type UpdateSkillPlanRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	SkillText   string `json:"skill_text" binding:"required"`
}

type SkillPlanItemDTO struct {
	SkillTypeID   int    `json:"skill_type_id"`
	SkillName     string `json:"skill_name"`
	RequiredLevel int    `json:"required_level"`
}

type SkillPlanDTO struct {
	ID          uint               `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	CreatedBy   uint               `json:"created_by"`
	CreatedAt   string             `json:"created_at"`
	UpdatedAt   string             `json:"updated_at"`
	Items       []SkillPlanItemDTO `json:"items"`
}

// 技能检查结果

type SkillCheckCharacterResult struct {
	UserID        uint               `json:"user_id"`
	UserName      string             `json:"user_name"`
	CharacterID   int64              `json:"character_id"`
	CharacterName string             `json:"character_name"`
	Satisfied     int                `json:"satisfied"`
	Total         int                `json:"total"`
	Status        string             `json:"status"` // "satisfied" | "unsatisfied"
	MissingSkills []MissingSkillItem `json:"missing_skills"`
}

type MissingSkillItem struct {
	SkillTypeID   int    `json:"skill_type_id"`
	SkillName     string `json:"skill_name"`
	RequiredLevel int    `json:"required_level"`
	CurrentLevel  int    `json:"current_level"`
}

type SkillCheckSummary struct {
	PlanName         string                      `json:"plan_name"`
	TotalCharacters  int                         `json:"total_characters"`
	SatisfiedCount   int                         `json:"satisfied_count"`
	UnsatisfiedCount int                         `json:"unsatisfied_count"`
	SatisfiedRate    float64                     `json:"satisfied_rate"`
	Characters       []SkillCheckCharacterResult `json:"characters"`
}

// ── 解析剪贴板文本 ──

// 解析格式：
// <localized hint="English Name">中文名</localized> Level
// 或纯文本：技能名 Level
var localizedRe = regexp.MustCompile(`<localized\s+hint="([^"]+)">[^<]*</localized>\s+(\d+)`)
var plainSkillRe = regexp.MustCompile(`^(.+?)\s+(\d+)$`)

type parsedSkill struct {
	EnglishName string
	Level       int
}

func parseSkillText(text string) ([]parsedSkill, error) {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	var skills []parsedSkill
	seen := make(map[string]int) // name -> max level

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var name string
		var level int

		// 尝试 localized 格式
		if matches := localizedRe.FindStringSubmatch(line); len(matches) == 3 {
			name = strings.TrimSpace(matches[1])
			level, _ = strconv.Atoi(matches[2])
		} else if matches := plainSkillRe.FindStringSubmatch(line); len(matches) == 3 {
			// 纯文本格式
			name = strings.TrimSpace(matches[1])
			level, _ = strconv.Atoi(matches[2])
		} else {
			continue
		}

		if level < 1 || level > 5 || name == "" {
			continue
		}

		// 同名技能取最高等级
		if existing, ok := seen[name]; !ok || level > existing {
			seen[name] = level
		}
	}

	for name, level := range seen {
		skills = append(skills, parsedSkill{EnglishName: name, Level: level})
	}

	if len(skills) == 0 {
		return nil, errors.New("no valid skills found in text")
	}
	return skills, nil
}

// ── CRUD ──

func (s *SkillPlanService) Create(userID uint, req *CreateSkillPlanRequest) (*SkillPlanDTO, error) {
	parsed, err := parseSkillText(req.SkillText)
	if err != nil {
		return nil, err
	}

	// 查找英文名对应的 typeID
	names := make([]string, 0, len(parsed))
	for _, p := range parsed {
		names = append(names, p.EnglishName)
	}
	nameToID, err := s.sdeRepo.GetTypeIDsByNames(names)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup skill types: %w", err)
	}

	items := make([]model.SkillPlanItem, 0, len(parsed))
	for _, p := range parsed {
		typeID, ok := nameToID[p.EnglishName]
		if !ok {
			return nil, fmt.Errorf("unknown skill: %s", p.EnglishName)
		}
		items = append(items, model.SkillPlanItem{
			SkillTypeID:   int(typeID),
			RequiredLevel: p.Level,
		})
	}

	plan := &model.SkillPlan{
		Name:        req.Name,
		Description: req.Description,
		CreatedBy:   userID,
	}

	if err := s.planRepo.Create(plan, items); err != nil {
		return nil, err
	}

	return s.toDTO(plan, items)
}

func (s *SkillPlanService) Update(id uint, req *UpdateSkillPlanRequest) (*SkillPlanDTO, error) {
	plan, err := s.planRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("skill plan not found")
	}

	parsed, err := parseSkillText(req.SkillText)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(parsed))
	for _, p := range parsed {
		names = append(names, p.EnglishName)
	}
	nameToID, err := s.sdeRepo.GetTypeIDsByNames(names)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup skill types: %w", err)
	}

	items := make([]model.SkillPlanItem, 0, len(parsed))
	for _, p := range parsed {
		typeID, ok := nameToID[p.EnglishName]
		if !ok {
			return nil, fmt.Errorf("unknown skill: %s", p.EnglishName)
		}
		items = append(items, model.SkillPlanItem{
			SkillTypeID:   int(typeID),
			RequiredLevel: p.Level,
		})
	}

	plan.Name = req.Name
	plan.Description = req.Description

	if err := s.planRepo.Update(plan, items); err != nil {
		return nil, err
	}

	return s.toDTO(plan, items)
}

func (s *SkillPlanService) Delete(id uint) error {
	return s.planRepo.Delete(id)
}

func (s *SkillPlanService) GetByID(id uint, language string) (*SkillPlanDTO, error) {
	plan, err := s.planRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("skill plan not found")
	}
	items, err := s.planRepo.GetItems(id)
	if err != nil {
		return nil, err
	}
	return s.toDTOWithNames(plan, items, language)
}

func (s *SkillPlanService) List(page, pageSize int) ([]SkillPlanDTO, int64, error) {
	plans, total, err := s.planRepo.List(page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	dtos := make([]SkillPlanDTO, 0, len(plans))
	for _, p := range plans {
		items, _ := s.planRepo.GetItems(p.ID)
		dto, _ := s.toDTO(&p, items)
		if dto != nil {
			dtos = append(dtos, *dto)
		}
	}
	return dtos, total, nil
}

func (s *SkillPlanService) ListAll() ([]SkillPlanDTO, error) {
	plans, err := s.planRepo.ListAll()
	if err != nil {
		return nil, err
	}
	dtos := make([]SkillPlanDTO, 0, len(plans))
	for _, p := range plans {
		dto := SkillPlanDTO{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			CreatedBy:   p.CreatedBy,
			CreatedAt:   p.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   p.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		dtos = append(dtos, dto)
	}
	return dtos, nil
}

// ── 技能检查 ──

func (s *SkillPlanService) CheckAllCharacters(planID uint, language string) (*SkillCheckSummary, error) {
	plan, err := s.planRepo.GetByID(planID)
	if err != nil {
		return nil, errors.New("skill plan not found")
	}

	items, err := s.planRepo.GetItems(planID)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, errors.New("skill plan has no items")
	}

	// 获取所有有效角色
	chars, err := s.charRepo.ListAllWithToken()
	if err != nil {
		return nil, err
	}

	return s.doCheck(plan, items, chars, language)
}

func (s *SkillPlanService) CheckUserCharacters(planID uint, userID uint, language string) (*SkillCheckSummary, error) {
	plan, err := s.planRepo.GetByID(planID)
	if err != nil {
		return nil, errors.New("skill plan not found")
	}

	items, err := s.planRepo.GetItems(planID)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, errors.New("skill plan has no items")
	}

	chars, err := s.charRepo.ListByUserID(userID)
	if err != nil {
		return nil, err
	}

	return s.doCheck(plan, items, chars, language)
}

func (s *SkillPlanService) doCheck(plan *model.SkillPlan, items []model.SkillPlanItem, chars []model.EveCharacter, language string) (*SkillCheckSummary, error) {
	if language == "" {
		language = "zh"
	}

	// 收集所有 char ID
	charIDs := make([]int64, 0, len(chars))
	for _, c := range chars {
		charIDs = append(charIDs, c.CharacterID)
	}

	// 批量获取所有角色技能
	allSkills, err := s.planRepo.GetSkillsByCharacterIDs(charIDs)
	if err != nil {
		return nil, err
	}

	// 构建 charID -> skillID -> level 映射
	charSkillMap := make(map[int64]map[int]int)
	for _, sk := range allSkills {
		if _, ok := charSkillMap[sk.CharacterID]; !ok {
			charSkillMap[sk.CharacterID] = make(map[int]int)
		}
		charSkillMap[sk.CharacterID][sk.SkillID] = sk.TrainedLevel
	}

	// 获取技能名称
	skillIDs := make([]int, 0, len(items))
	for _, item := range items {
		skillIDs = append(skillIDs, item.SkillTypeID)
	}
	published := true
	typeInfos, _ := s.sdeRepo.GetTypes(skillIDs, &published, language)
	nameMap := make(map[int]string)
	for _, ti := range typeInfos {
		nameMap[ti.TypeID] = ti.TypeName
	}

	// 构建 charID -> user 映射
	charToUser := make(map[int64]struct {
		userID   uint
		userName string
	})
	// 简单用角色名作为用户名，因为一个角色对应一个用户
	for _, c := range chars {
		charToUser[c.CharacterID] = struct {
			userID   uint
			userName string
		}{userID: c.UserID, userName: ""}
	}

	// 获取用户名
	userIDs := make([]uint, 0)
	userIDSet := make(map[uint]bool)
	for _, c := range chars {
		if !userIDSet[c.UserID] {
			userIDSet[c.UserID] = true
			userIDs = append(userIDs, c.UserID)
		}
	}
	userRepo := repository.NewUserRepository()
	userNames := make(map[uint]string)
	for _, uid := range userIDs {
		u, err := userRepo.GetByID(uid)
		if err == nil {
			userNames[uid] = u.Nickname
		}
	}

	// 检查每个角色
	var results []SkillCheckCharacterResult
	satisfiedCount := 0

	for _, c := range chars {
		skillMap := charSkillMap[c.CharacterID]
		satisfied := 0
		var missing []MissingSkillItem

		for _, item := range items {
			currentLevel := 0
			if skillMap != nil {
				currentLevel = skillMap[item.SkillTypeID]
			}
			if currentLevel >= item.RequiredLevel {
				satisfied++
			} else {
				missing = append(missing, MissingSkillItem{
					SkillTypeID:   item.SkillTypeID,
					SkillName:     nameMap[item.SkillTypeID],
					RequiredLevel: item.RequiredLevel,
					CurrentLevel:  currentLevel,
				})
			}
		}

		status := "unsatisfied"
		if satisfied == len(items) {
			status = "satisfied"
			satisfiedCount++
		}

		results = append(results, SkillCheckCharacterResult{
			UserID:        c.UserID,
			UserName:      userNames[c.UserID],
			CharacterID:   c.CharacterID,
			CharacterName: c.CharacterName,
			Satisfied:     satisfied,
			Total:         len(items),
			Status:        status,
			MissingSkills: missing,
		})
	}

	totalChars := len(results)
	rate := float64(0)
	if totalChars > 0 {
		rate = float64(satisfiedCount) / float64(totalChars) * 100
	}

	return &SkillCheckSummary{
		PlanName:         plan.Name,
		TotalCharacters:  totalChars,
		SatisfiedCount:   satisfiedCount,
		UnsatisfiedCount: totalChars - satisfiedCount,
		SatisfiedRate:    rate,
		Characters:       results,
	}, nil
}

// ── 内部辅助 ──

func (s *SkillPlanService) toDTO(plan *model.SkillPlan, items []model.SkillPlanItem) (*SkillPlanDTO, error) {
	itemDTOs := make([]SkillPlanItemDTO, 0, len(items))
	for _, item := range items {
		itemDTOs = append(itemDTOs, SkillPlanItemDTO{
			SkillTypeID:   item.SkillTypeID,
			RequiredLevel: item.RequiredLevel,
		})
	}
	return &SkillPlanDTO{
		ID:          plan.ID,
		Name:        plan.Name,
		Description: plan.Description,
		CreatedBy:   plan.CreatedBy,
		CreatedAt:   plan.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   plan.UpdatedAt.Format("2006-01-02 15:04:05"),
		Items:       itemDTOs,
	}, nil
}

func (s *SkillPlanService) toDTOWithNames(plan *model.SkillPlan, items []model.SkillPlanItem, language string) (*SkillPlanDTO, error) {
	if language == "" {
		language = "zh"
	}
	skillIDs := make([]int, 0, len(items))
	for _, item := range items {
		skillIDs = append(skillIDs, item.SkillTypeID)
	}
	published := true
	typeInfos, _ := s.sdeRepo.GetTypes(skillIDs, &published, language)
	nameMap := make(map[int]string)
	for _, ti := range typeInfos {
		nameMap[ti.TypeID] = ti.TypeName
	}

	itemDTOs := make([]SkillPlanItemDTO, 0, len(items))
	for _, item := range items {
		itemDTOs = append(itemDTOs, SkillPlanItemDTO{
			SkillTypeID:   item.SkillTypeID,
			SkillName:     nameMap[item.SkillTypeID],
			RequiredLevel: item.RequiredLevel,
		})
	}

	return &SkillPlanDTO{
		ID:          plan.ID,
		Name:        plan.Name,
		Description: plan.Description,
		CreatedBy:   plan.CreatedBy,
		CreatedAt:   plan.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   plan.UpdatedAt.Format("2006-01-02 15:04:05"),
		Items:       itemDTOs,
	}, nil
}
