package service

import (
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ATMatchService struct {
	repo   *repository.ATMatchRepository
	setSvc *ATSetService
	brRepo *repository.ATBRRepository
	users  *repository.UserRepository
	sde    *repository.SdeRepository
}

func NewATMatchService() *ATMatchService {
	return &ATMatchService{repo: repository.NewATMatchRepository(), setSvc: NewATSetService(), brRepo: repository.NewATBRRepository(), users: repository.NewUserRepository(), sde: repository.NewSdeRepository()}
}

func (s *ATMatchService) SaveSeason(season *model.ATSeason) error {
	if strings.TrimSpace(season.Name) == "" || !season.EndsAt.After(season.StartsAt) {
		return errors.New("赛季名称或时间范围无效")
	}
	return s.repo.SaveSeason(season)
}
func (s *ATMatchService) ListSeasons() ([]model.ATSeason, error) { return s.repo.ListSeasons() }
func (s *ATMatchService) SaveMatchType(item *model.ATMatchType) error {
	item.Name = strings.TrimSpace(item.Name)
	if item.Name == "" || item.ContributionValue < 0 {
		return errors.New("比赛类型或贡献值无效")
	}
	return s.repo.SaveMatchType(item)
}
func (s *ATMatchService) ListMatchTypes() ([]model.ATMatchType, error) {
	return s.repo.ListMatchTypes()
}
func (s *ATMatchService) SetCleanupShips(seasonID uint, typeIDs []int64) error {
	if _, err := s.repo.GetSeason(seasonID); err != nil {
		return err
	}
	seen := map[int64]bool{}
	ships := make([]model.ATCleanupShip, 0, len(typeIDs))
	for _, id := range typeIDs {
		if id <= 0 || seen[id] {
			return errors.New("清场船 TypeID 必须为不重复正整数")
		}
		seen[id] = true
		ships = append(ships, model.ATCleanupShip{TypeID: id})
	}
	if len(typeIDs) > 0 {
		ids := make([]int, len(typeIDs))
		for i, id := range typeIDs {
			ids[i] = int(id)
		}
		types, err := s.sde.GetTypes(ids, nil, "en")
		if err != nil {
			return fmt.Errorf("校验清场船失败: %w", err)
		}
		valid := make(map[int64]bool, len(types))
		for _, item := range types {
			if item.CategoryID == 6 {
				valid[int64(item.TypeID)] = true
			}
		}
		for _, id := range typeIDs {
			if !valid[id] {
				return fmt.Errorf("TypeID %d 不是有效舰船", id)
			}
		}
	}
	return s.repo.SetCleanupShips(seasonID, ships)
}

type ATCleanupShipDTO struct {
	TypeID   int64  `json:"type_id"`
	TypeName string `json:"type_name"`
}

func (s *ATMatchService) ListCleanupShips(seasonID uint, language string) ([]ATCleanupShipDTO, error) {
	if _, err := s.repo.GetSeason(seasonID); err != nil {
		return nil, err
	}
	ships, err := s.repo.GetCleanupShips(seasonID)
	if err != nil || len(ships) == 0 {
		return []ATCleanupShipDTO{}, err
	}
	ids := make([]int, len(ships))
	for i, ship := range ships {
		ids[i] = int(ship.TypeID)
	}
	if language != "zh" && language != "en" {
		language = "en"
	}
	types, err := s.sde.GetTypes(ids, nil, language)
	if err != nil {
		return nil, err
	}
	names := make(map[int64]string, len(types))
	for _, item := range types {
		names[int64(item.TypeID)] = item.TypeName
	}
	result := make([]ATCleanupShipDTO, len(ships))
	for i, ship := range ships {
		result[i] = ATCleanupShipDTO{TypeID: ship.TypeID, TypeName: names[ship.TypeID]}
		if result[i].TypeName == "" {
			result[i].TypeName = fmt.Sprintf("TypeID %d", ship.TypeID)
		}
	}
	return result, nil
}

type ATContributionUserOption struct {
	ID                      uint   `json:"id"`
	Nickname                string `json:"nickname"`
	TournamentCharacterName string `json:"tournament_character_name"`
}

func (s *ATMatchService) SearchContributionUsers(keyword string) ([]ATContributionUserOption, error) {
	users, err := s.repo.SearchContributionUsers(strings.TrimSpace(keyword), 20)
	if err != nil {
		return nil, err
	}
	result := make([]ATContributionUserOption, len(users))
	for i, user := range users {
		result[i] = ATContributionUserOption{ID: user.ID, Nickname: user.Nickname, TournamentCharacterName: user.TournamentCharName}
	}
	return result, nil
}

type ATMatchCreateInput struct {
	SeasonID        uint   `json:"season_id"`
	MatchTypeID     uint   `json:"match_type_id"`
	Format          string `json:"format"`
	Series          string `json:"series"`
	TeamAName       string `json:"team_a_name"`
	TeamBName       string `json:"team_b_name"`
	SetID           uint   `json:"set_id"`
	AllowSelfAssign bool   `json:"allow_self_assign"`
}

func (s *ATMatchService) Create(userID uint, input ATMatchCreateInput) (*model.ATMatch, error) {
	if input.Format != "7v7" && input.Format != "10v10" {
		return nil, errors.New("赛制仅支持 7v7 或 10v10")
	}
	if input.Series != "BO3" && input.Series != "BO5" {
		return nil, errors.New("系列赛仅支持 BO3 或 BO5")
	}
	if strings.TrimSpace(input.TeamAName) == "" || strings.TrimSpace(input.TeamBName) == "" {
		return nil, errors.New("双方队名不能为空")
	}
	if _, err := s.repo.GetSeason(input.SeasonID); err != nil {
		return nil, errors.New("赛季不存在")
	}
	matchType, err := s.repo.GetMatchType(input.MatchTypeID)
	if err != nil || !matchType.Active {
		return nil, errors.New("比赛类型不可用")
	}
	set, err := s.setSvc.Get(input.SetID)
	if err != nil {
		return nil, err
	}
	if set.Set.Format != input.Format {
		return nil, errors.New("Set 赛制与比赛不一致")
	}
	if !set.Score.Valid {
		return nil, errors.New("Set 校验未通过")
	}
	ruleSnapshot, _ := json.Marshal(set.Score)
	cleanup, err := s.repo.GetCleanupShips(input.SeasonID)
	if err != nil {
		return nil, err
	}
	cleanupIDs := make([]int64, len(cleanup))
	for i := range cleanup {
		cleanupIDs[i] = cleanup[i].TypeID
	}
	cleanupSnapshot, _ := json.Marshal(cleanupIDs)
	matchUUID, uuidErr := uuid.NewV7()
	if uuidErr != nil {
		return nil, uuidErr
	}
	match := &model.ATMatch{UUID: matchUUID.String(), SeasonID: input.SeasonID, MatchTypeID: input.MatchTypeID, Format: input.Format, Series: input.Series, TeamAName: strings.TrimSpace(input.TeamAName), TeamBName: strings.TrimSpace(input.TeamBName), SetID: input.SetID, RuleVersionID: set.Set.RuleVersionID, RuleSnapshot: ruleSnapshot, CleanupShipsSnapshot: cleanupSnapshot, ContributionValue: matchType.ContributionValue, FCUserID: userID, AllowSelfAssign: input.AllowSelfAssign, Status: "draft"}
	if err := s.repo.CreateMatch(match); err != nil {
		return nil, err
	}
	return match, nil
}

func (s *ATMatchService) Get(id uint) (*model.ATMatch, []model.ATMatchRound, []model.ATMatchRegistration, error) {
	match, err := s.repo.GetMatch(id)
	if err != nil {
		return nil, nil, nil, err
	}
	if match.UUID == "" {
		generated, _ := uuid.NewV7()
		match.UUID = generated.String()
		if err := s.repo.SetMatchUUID(match.ID, match.UUID); err != nil {
			return nil, nil, nil, err
		}
	}
	rounds, err := s.repo.ListRounds(id)
	if err != nil {
		return nil, nil, nil, err
	}
	if match.Status == "in_progress" && len(rounds) == 0 {
		if _, createErr := s.createAutomatedRound(match, match.FCUserID, nil); createErr != nil {
			return nil, nil, nil, createErr
		}
		rounds, err = s.repo.ListRounds(id)
		if err != nil {
			return nil, nil, nil, err
		}
	}
	registrations, err := s.repo.ListRegistrations(id)
	if err == nil && len(registrations) > 0 {
		ids := make([]uint, len(registrations))
		for i := range registrations {
			ids[i] = registrations[i].UserID
		}
		users, userErr := s.users.ListByIDs(ids)
		if userErr != nil {
			return nil, nil, nil, userErr
		}
		byID := map[uint]model.User{}
		for _, user := range users {
			byID[user.ID] = user
		}
		for i := range registrations {
			registrations[i].UserNickname = byID[registrations[i].UserID].Nickname
			registrations[i].TournamentCharacterName = byID[registrations[i].UserID].TournamentCharName
		}
	}
	return match, rounds, registrations, err
}
func (s *ATMatchService) GetByUUID(id string) (*model.ATMatch, []model.ATMatchRound, []model.ATMatchRegistration, error) {
	match, err := s.repo.GetMatchByUUID(strings.TrimSpace(id))
	if err != nil {
		return nil, nil, nil, err
	}
	return s.Get(match.ID)
}
func (s *ATMatchService) List() ([]model.ATMatch, error) {
	items, err := s.repo.ListMatches()
	if err != nil {
		return nil, err
	}
	for i := range items {
		if items[i].UUID == "" {
			generated, _ := uuid.NewV7()
			items[i].UUID = generated.String()
			if err := s.repo.SetMatchUUID(items[i].ID, items[i].UUID); err != nil {
				return nil, err
			}
		}
	}
	return items, nil
}

var atMatchTransitions = map[string]map[string]bool{"draft": {"recruiting": true, "cancelled": true}, "recruiting": {"ready": true, "cancelled": true}, "ready": {"in_progress": true, "cancelled": true}, "in_progress": {"early_ended": true, "cancelled": true}}

func (s *ATMatchService) Transition(actor, id uint, to, reason string, admin bool) error {
	match, err := s.repo.GetMatch(id)
	if err != nil {
		return err
	}
	if !admin && match.FCUserID != actor {
		return errors.New("只能管理自己创建的比赛")
	}
	if !atMatchTransitions[match.Status][to] {
		return errors.New("不允许的比赛状态转换")
	}
	if (to == "early_ended" || to == "cancelled") && strings.TrimSpace(reason) == "" {
		return errors.New("提前结束或取消必须填写原因")
	}
	updates := map[string]any{}
	if to == "early_ended" || to == "cancelled" {
		updates["end_reason"] = reason
	}
	if err := s.repo.UpdateMatchStatus(id, actor, match.Status, to, reason, updates); err != nil {
		return err
	}
	if to == "in_progress" {
		match.Status = to
		_, err = s.createAutomatedRound(match, actor, nil)
		return err
	}
	if to == "early_ended" {
		return s.Settle(id, actor)
	}
	return nil
}

func (s *ATMatchService) Register(userID, matchID uint) error {
	match, err := s.repo.GetMatch(matchID)
	if err != nil {
		return err
	}
	if match.Status != "recruiting" {
		return errors.New("比赛不在招募中")
	}
	return errors.New("请选择具体阵容位置报名")
}
func (s *ATMatchService) DecideRegistration(actor, matchID, userID uint, status string, admin bool) error {
	match, err := s.repo.GetMatch(matchID)
	if err != nil {
		return err
	}
	if !admin && match.FCUserID != actor {
		return errors.New("无权管理该比赛")
	}
	if status != "confirmed" && status != "rejected" && status != "assigned" {
		return errors.New("报名状态无效")
	}
	registrations, err := s.repo.ListRegistrations(matchID)
	if err != nil {
		return err
	}
	for _, registration := range registrations {
		if registration.UserID == userID {
			registration.Status = status
			return s.repo.UpsertRegistration(&registration)
		}
	}
	return errors.New("报名记录不存在")
}

type ATPositionInput struct {
	UserID    uint     `json:"user_id"`
	SlotOrder int      `json:"slot_order"`
	EWarTags  []string `json:"ewar_tags"`
	BonusTags []string `json:"bonus_tags"`
}

func (s *ATMatchService) SavePosition(actor, matchID uint, input ATPositionInput, admin bool, self bool) error {
	match, err := s.repo.GetMatch(matchID)
	if err != nil {
		return err
	}
	if self {
		input.UserID = actor
		if !match.AllowSelfAssign {
			return errors.New("该比赛不允许成员自行选取位置")
		}
		if match.Status != "recruiting" {
			return errors.New("比赛不在招募中")
		}
	} else if !admin && match.FCUserID != actor {
		return errors.New("无权管理该比赛")
	}
	if input.UserID == 0 || input.SlotOrder <= 0 {
		return errors.New("用户和阵容位置不能为空")
	}
	set, err := s.setSvc.Get(match.SetID)
	if err != nil {
		return err
	}
	var slot *repository.ATSetSlotBundle
	for i := range set.Slots {
		if set.Slots[i].Slot.SlotOrder == input.SlotOrder {
			slot = &set.Slots[i]
			break
		}
	}
	if slot == nil {
		return errors.New("阵容位置不存在")
	}
	ewar, err := normalizeTagList(input.EWarTags)
	if err != nil {
		return err
	}
	bonus, err := normalizeTagList(input.BonusTags)
	if err != nil {
		return err
	}
	if len(ewar) == 0 {
		_ = json.Unmarshal(slot.Slot.EWarTags, &ewar)
	}
	if len(bonus) == 0 {
		_ = json.Unmarshal(slot.Slot.BonusTags, &bonus)
	}
	allowedEwar, allowedBonus := map[string]bool{}, map[string]bool{}
	if slot.Fit != nil {
		fit, fitErr := s.setSvc.GetFit(slot.Fit.ID)
		if fitErr != nil {
			return fitErr
		}
		for _, tag := range fit.EWarTags {
			allowedEwar[tag] = true
		}
		for _, tag := range fit.BonusTags {
			allowedBonus[tag] = true
		}
	} else {
		for _, tag := range ewar {
			allowedEwar[tag] = true
		}
		for _, tag := range bonus {
			allowedBonus[tag] = true
		}
	}
	for _, tag := range ewar {
		if !allowedEwar[tag] {
			return fmt.Errorf("位置不支持电子战标签 %s", tag)
		}
	}
	for _, tag := range bonus {
		if !allowedBonus[tag] {
			return fmt.Errorf("位置不支持加成标签 %s", tag)
		}
	}
	userRows, err := s.users.ListByIDs([]uint{input.UserID})
	if err != nil || len(userRows) != 1 {
		return errors.New("用户不存在")
	}
	if userRows[0].TournamentCharID == nil || strings.TrimSpace(userRows[0].TournamentCharName) == "" {
		return errors.New("该用户未完善比赛服角色资料")
	}
	ewarJSON, _ := json.Marshal(ewar)
	bonusJSON, _ := json.Marshal(bonus)
	status := "assigned"
	if self {
		status = "pending"
	}
	err = s.repo.UpsertRegistration(&model.ATMatchRegistration{MatchID: matchID, UserID: input.UserID, Status: status, SlotOrder: input.SlotOrder, EWarTags: ewarJSON, BonusTags: bonusJSON})
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return errors.New("该阵容位置已被其他成员占用")
	}
	if err != nil {
		return err
	}
	return s.syncActiveRoundRoster(match)
}

func (s *ATMatchService) snapshotRoster(match *model.ATMatch, previousB []model.ATMatchRoundRoster) ([]model.ATMatchRoundRoster, error) {
	set, err := s.setSvc.Get(match.SetID)
	if err != nil {
		return nil, err
	}
	regs, err := s.repo.ListRegistrations(match.ID)
	if err != nil {
		return nil, err
	}
	regBySlot := map[int]model.ATMatchRegistration{}
	var userIDs []uint
	for _, reg := range regs {
		if reg.Status != "rejected" && reg.SlotOrder > 0 {
			regBySlot[reg.SlotOrder] = reg
			userIDs = append(userIDs, reg.UserID)
		}
	}
	users, err := s.users.ListByIDs(userIDs)
	if err != nil {
		return nil, err
	}
	userByID := map[uint]model.User{}
	for _, user := range users {
		userByID[user.ID] = user
	}
	roster := make([]model.ATMatchRoundRoster, 0, len(set.Slots)+len(previousB))
	for _, bundle := range set.Slots {
		entry := model.ATMatchRoundRoster{Side: "A", SlotOrder: bundle.Slot.SlotOrder, ShipTypeID: bundle.Slot.ShipTypeID}
		if reg, ok := regBySlot[bundle.Slot.SlotOrder]; ok {
			uid := reg.UserID
			entry.UserID = &uid
			entry.CharacterName = userByID[uid].TournamentCharName
		}
		roster = append(roster, entry)
	}
	for _, entry := range previousB {
		entry.ID = 0
		entry.RoundID = 0
		entry.Side = "B"
		roster = append(roster, entry)
	}
	return roster, nil
}

func (s *ATMatchService) syncActiveRoundRoster(match *model.ATMatch) error {
	rounds, err := s.repo.ListRounds(match.ID)
	if err != nil {
		return err
	}
	for i := len(rounds) - 1; i >= 0; i-- {
		if rounds[i].Status != "in_progress" {
			continue
		}
		current, err := s.repo.GetRoster(rounds[i].ID)
		if err != nil {
			return err
		}
		var opponents []model.ATMatchRoundRoster
		for _, entry := range current {
			if entry.Side == "B" {
				opponents = append(opponents, entry)
			}
		}
		roster, err := s.snapshotRoster(match, opponents)
		if err != nil {
			return err
		}
		return s.repo.ReplaceRoster(rounds[i].ID, roster)
	}
	return nil
}

func (s *ATMatchService) createAutomatedRound(match *model.ATMatch, actor uint, previousB []model.ATMatchRoundRoster) (*model.ATMatchRound, error) {
	rounds, err := s.repo.ListRounds(match.ID)
	if err != nil {
		return nil, err
	}
	for _, round := range rounds {
		if round.Status == "in_progress" || round.Status == "br_review" {
			return nil, errors.New("仍有未完成的比赛局")
		}
	}
	roundNo := 1
	for _, round := range rounds {
		if round.Status != "void" && round.RoundNo >= roundNo {
			roundNo = round.RoundNo + 1
		}
	}
	roster, err := s.snapshotRoster(match, previousB)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	round := &model.ATMatchRound{MatchID: match.ID, RoundNo: roundNo, Attempt: 1, Status: "in_progress", StartedAt: &now}
	if err := s.repo.CreateRound(round, roster, actor); err != nil {
		return nil, err
	}
	return round, nil
}

func (s *ATMatchService) PositionEFT(actor, matchID uint, slotOrder int, admin bool) (*ATSetSlotEFT, error) {
	match, err := s.repo.GetMatch(matchID)
	if err != nil {
		return nil, err
	}
	regs, err := s.repo.ListRegistrations(matchID)
	if err != nil {
		return nil, err
	}
	var selected *model.ATMatchRegistration
	for i := range regs {
		if regs[i].SlotOrder == slotOrder && regs[i].Status != "rejected" {
			selected = &regs[i]
			break
		}
	}
	if selected == nil {
		return nil, errors.New("该位置尚未分配成员")
	}
	if selected.UserID != actor && !admin && match.FCUserID != actor {
		return nil, errors.New("只能复制自己位置的装配")
	}
	var ewar, bonus []string
	_ = json.Unmarshal(selected.EWarTags, &ewar)
	_ = json.Unmarshal(selected.BonusTags, &bonus)
	return s.setSvc.GeneratePositionEFT(match.SetID, slotOrder, ewar, bonus)
}

type ATCreateRoundInput struct {
	Roster []model.ATMatchRoundRoster `json:"roster"`
}

func (s *ATMatchService) CreateRound(actor, matchID uint, input ATCreateRoundInput, admin bool) (*model.ATMatchRound, error) {
	match, err := s.repo.GetMatch(matchID)
	if err != nil {
		return nil, err
	}
	if !admin && match.FCUserID != actor {
		return nil, errors.New("无权管理该比赛")
	}
	if match.Status != "in_progress" {
		return nil, errors.New("比赛尚未进行")
	}
	rounds, err := s.repo.ListRounds(matchID)
	if err != nil {
		return nil, err
	}
	roundNo := 1
	for _, round := range rounds {
		if round.Status != "void" && round.RoundNo >= roundNo {
			roundNo = round.RoundNo + 1
		}
	}
	if err := s.validateRoundRoster(match, input.Roster, false); err != nil {
		return nil, err
	}
	now := time.Now()
	round := &model.ATMatchRound{MatchID: matchID, RoundNo: roundNo, Attempt: 1, Status: "in_progress", StartedAt: &now}
	if err := s.repo.CreateRound(round, input.Roster, actor); err != nil {
		return nil, err
	}
	return round, nil
}

func (s *ATMatchService) validateRoundRoster(match *model.ATMatch, roster []model.ATMatchRoundRoster, requireBoth bool) error {
	max := 7
	if match.Format == "10v10" {
		max = 10
	}
	counts := map[string]int{}
	slots := map[string]bool{}
	characterNames := map[string]bool{}
	userIDs := []uint{}
	seenUsers := map[uint]bool{}
	for _, entry := range roster {
		if entry.Side != "A" && entry.Side != "B" {
			return errors.New("阵容 side 仅支持 A/B")
		}
		if entry.SlotOrder < 1 || entry.SlotOrder > max {
			return errors.New("阵容槽位超出赛制上限")
		}
		key := fmt.Sprintf("%s:%d", entry.Side, entry.SlotOrder)
		if slots[key] {
			return errors.New("阵容槽位重复")
		}
		slots[key] = true
		counts[entry.Side]++
		if strings.TrimSpace(entry.CharacterName) == "" || entry.ShipTypeID <= 0 {
			return errors.New("阵容角色名和舰船不能为空")
		}
		identity := strings.ToLower(strings.Join(strings.Fields(entry.CharacterName), " "))
		if characterNames[identity] {
			return errors.New("同一角色不能重复占用阵容槽位")
		}
		characterNames[identity] = true
		if entry.Side == "A" {
			if entry.UserID == nil || *entry.UserID == 0 {
				return errors.New("己方阵容必须关联系统用户")
			}
			if seenUsers[*entry.UserID] {
				return errors.New("同一用户不能重复上场")
			}
			seenUsers[*entry.UserID] = true
			userIDs = append(userIDs, *entry.UserID)
		}
	}
	if requireBoth && counts["A"] != max {
		return fmt.Errorf("己方阵容必须完整填写 %d 人", max)
	}
	if requireBoth && counts["B"] != max {
		return fmt.Errorf("对方阵容必须完整填写 %d 人", max)
	}
	users, err := s.users.ListByIDs(userIDs)
	if err != nil {
		return err
	}
	profiles := make(map[uint]model.User, len(users))
	for _, user := range users {
		profiles[user.ID] = user
	}
	for _, entry := range roster {
		if entry.Side != "A" || entry.UserID == nil {
			continue
		}
		user, ok := profiles[*entry.UserID]
		if !ok || user.TournamentCharID == nil || *user.TournamentCharID <= 0 || strings.TrimSpace(user.TournamentCharName) == "" {
			return fmt.Errorf("用户 %d 未完善比赛服资料", *entry.UserID)
		}
		if !strings.EqualFold(strings.TrimSpace(entry.CharacterName), strings.TrimSpace(user.TournamentCharName)) {
			return fmt.Errorf("用户 %d 的比赛服角色名不匹配", *entry.UserID)
		}
	}
	set, err := s.setSvc.Get(match.SetID)
	if err != nil {
		return err
	}
	expected, actual := map[int64]int{}, map[int64]int{}
	for _, slot := range set.Slots {
		expected[slot.Slot.ShipTypeID]++
	}
	for _, entry := range roster {
		if entry.Side == "A" {
			actual[entry.ShipTypeID]++
		}
	}
	for typeID, count := range expected {
		if actual[typeID] > count || ((requireBoth || counts["A"] == max) && actual[typeID] != count) {
			return errors.New("己方每局舰船必须与比赛固定 Set 一致")
		}
	}
	for typeID := range actual {
		if expected[typeID] == 0 {
			return errors.New("己方每局舰船必须与比赛固定 Set 一致")
		}
	}
	return nil
}

func (s *ATMatchService) SetRoundRoster(actor, roundID uint, roster []model.ATMatchRoundRoster, admin bool) error {
	round, err := s.repo.GetRound(roundID)
	if err != nil {
		return err
	}
	match, err := s.repo.GetMatch(round.MatchID)
	if err != nil {
		return err
	}
	if !admin && match.FCUserID != actor {
		return errors.New("无权管理该比赛")
	}
	if round.Status == "confirmed" || round.Status == "void" {
		return errors.New("该局阵容已锁定")
	}
	if err := s.validateRoundRoster(match, roster, false); err != nil {
		return err
	}
	return s.repo.ReplaceRoster(roundID, roster)
}
func (s *ATMatchService) GetRoundRoster(actor, roundID uint, admin bool) ([]model.ATMatchRoundRoster, error) {
	round, err := s.repo.GetRound(roundID)
	if err != nil {
		return nil, err
	}
	match, err := s.repo.GetMatch(round.MatchID)
	if err != nil {
		return nil, err
	}
	if !admin && match.FCUserID != actor {
		return nil, errors.New("无权查看该局阵容")
	}
	return s.repo.GetRoster(roundID)
}
func (s *ATMatchService) FinishRound(actor, roundID uint, admin bool) error {
	round, err := s.repo.GetRound(roundID)
	if err != nil {
		return err
	}
	match, err := s.repo.GetMatch(round.MatchID)
	if err != nil {
		return err
	}
	if !admin && match.FCUserID != actor {
		return errors.New("无权管理该比赛")
	}
	if round.Status != "in_progress" {
		return errors.New("该局不在进行中")
	}
	from := round.Status
	round.Status = "br_review"
	return s.repo.UpdateRound(round, actor, from)
}
func (s *ATMatchService) VoidAndReopen(actor, roundID uint, reason string, admin bool) (*model.ATMatchRound, error) {
	if strings.TrimSpace(reason) == "" {
		return nil, errors.New("作废重开必须填写原因")
	}
	round, err := s.repo.GetRound(roundID)
	if err != nil {
		return nil, err
	}
	match, err := s.repo.GetMatch(round.MatchID)
	if err != nil {
		return nil, err
	}
	if !admin && match.FCUserID != actor {
		return nil, errors.New("无权管理该比赛")
	}
	roster, _ := s.repo.GetRoster(roundID)
	from := round.Status
	round.Status = "void"
	round.VoidReason = reason
	if err := s.repo.UpdateRound(round, actor, from); err != nil {
		return nil, err
	}
	now := time.Now()
	replacement := &model.ATMatchRound{MatchID: round.MatchID, RoundNo: round.RoundNo, Attempt: round.Attempt + 1, Status: "in_progress", StartedAt: &now}
	if err := s.repo.CreateRound(replacement, roster, actor); err != nil {
		return nil, err
	}
	return replacement, nil
}

func (s *ATMatchService) ConfirmRound(actor, roundID uint, result string, admin bool) error {
	if result != "a_win" && result != "b_win" && result != "draw" {
		return errors.New("赛果仅支持 a_win/b_win/draw")
	}
	round, err := s.repo.GetRound(roundID)
	if err != nil {
		return err
	}
	match, err := s.repo.GetMatch(round.MatchID)
	if err != nil {
		return err
	}
	if !admin && match.FCUserID != actor {
		return errors.New("无权管理该比赛")
	}
	if round.Status != "br_review" {
		return errors.New("该局未进入 BR 审核")
	}
	brExists, err := s.brRepo.Exists(roundID)
	if err != nil {
		return err
	}
	if !brExists {
		return errors.New("该局必须先固化 BR")
	}
	roster, err := s.repo.GetRoster(roundID)
	if err != nil {
		return err
	}
	if err := s.validateRoundRoster(match, roster, true); err != nil {
		return err
	}
	from := round.Status
	now := time.Now()
	round.Status = "confirmed"
	round.Result = result
	round.CompletedAt = &now
	if err := s.repo.UpdateRound(round, actor, from); err != nil {
		return err
	}
	oldStatus := match.Status
	if result == "a_win" {
		match.TeamAWins++
	}
	if result == "b_win" {
		match.TeamBWins++
	}
	target := 2
	if match.Series == "BO5" {
		target = 3
	}
	if match.TeamAWins >= target || match.TeamBWins >= target {
		match.Status = "completed"
	}
	if err := s.repo.UpdateWinsAndStatus(match, actor, oldStatus); err != nil {
		return err
	}
	if match.Status == "completed" {
		return s.Settle(match.ID, actor)
	}
	previousRoster, err := s.repo.GetRoster(roundID)
	if err != nil {
		return err
	}
	previousB := make([]model.ATMatchRoundRoster, 0)
	for _, entry := range previousRoster {
		if entry.Side == "B" {
			previousB = append(previousB, entry)
		}
	}
	_, err = s.createAutomatedRound(match, actor, previousB)
	return err
}

func (s *ATMatchService) Settle(matchID, actor uint) error {
	match, err := s.repo.GetMatch(matchID)
	if err != nil {
		return err
	}
	ids, err := s.repo.ParticipatingUserIDs(matchID)
	if err != nil {
		return err
	}
	entries := make([]model.ATContributionLedger, len(ids))
	for i, userID := range ids {
		mid := matchID
		entries[i] = model.ATContributionLedger{SeasonID: match.SeasonID, UserID: userID, Amount: match.ContributionValue, SourceType: "match", SourceKey: fmt.Sprintf("match:%d:user:%d", matchID, userID), MatchID: &mid, Reason: "match participation", ActorUserID: actor}
	}
	return s.repo.InsertLedger(entries)
}
func (s *ATMatchService) Reset(actor, matchID uint, reason string) error {
	if strings.TrimSpace(reason) == "" {
		return errors.New("整场重置必须填写原因")
	}
	match, err := s.repo.GetMatch(matchID)
	if err != nil {
		return err
	}
	ledger, err := s.repo.MatchLedger(matchID)
	if err != nil {
		return err
	}
	reversals := make([]model.ATContributionLedger, len(ledger))
	for i, entry := range ledger {
		mid := matchID
		reversals[i] = model.ATContributionLedger{SeasonID: entry.SeasonID, UserID: entry.UserID, Amount: -entry.Amount, SourceType: "reversal", SourceKey: fmt.Sprintf("reversal:%d", entry.ID), MatchID: &mid, Reason: reason, ActorUserID: actor}
	}
	if err := s.repo.InsertLedger(reversals); err != nil {
		return err
	}
	if err := s.repo.VoidAllRounds(matchID, actor, reason); err != nil {
		return err
	}
	return s.repo.UpdateMatchStatus(matchID, actor, match.Status, "draft", reason, map[string]any{"team_a_wins": 0, "team_b_wins": 0, "end_reason": ""})
}

func (s *ATMatchService) ManualContribution(actor, seasonID, userID uint, amount float64, reason string) error {
	if amount == 0 || strings.TrimSpace(reason) == "" {
		return errors.New("手工贡献必须填写非零数值和原因")
	}
	entry := model.ATContributionLedger{SeasonID: seasonID, UserID: userID, Amount: amount, SourceType: "manual", SourceKey: fmt.Sprintf("manual:%d:%d", actor, time.Now().UnixNano()), Reason: reason, ActorUserID: actor}
	return s.repo.InsertLedger([]model.ATContributionLedger{entry})
}
func (s *ATMatchService) Leaderboard(seasonID *uint) ([]repository.ATContributionLeaderboardRow, error) {
	return s.repo.Leaderboard(seasonID)
}

func (s *ATMatchService) UserLedger(userID uint, seasonID *uint) ([]model.ATContributionLedger, error) {
	return s.repo.UserLedger(userID, seasonID)
}
