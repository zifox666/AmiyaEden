package service

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"amiya-eden/pkg/eve/esi"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// WelfareService 福利业务逻辑层
type welfareDeliveryMailSender func(ctx context.Context, reviewerID uint, deliveredWelfare *model.Welfare, deliveredApp *model.WelfareApplication) (MailAttemptSummary, error)

type WelfareService struct {
	repo      *repository.WelfareRepository
	userRepo  *repository.UserRepository
	charRepo  *repository.EveCharacterRepository
	fleetRepo *repository.FleetRepository
	skillRepo *repository.EveSkillRepository
	planRepo  *repository.SkillPlanRepository
	ssoSvc    *EveSSOService
	esiClient *esi.Client

	deliveryMailSender welfareDeliveryMailSender
}

func NewWelfareService() *WelfareService {
	svc := &WelfareService{
		repo:      repository.NewWelfareRepository(),
		userRepo:  repository.NewUserRepository(),
		charRepo:  repository.NewEveCharacterRepository(),
		fleetRepo: repository.NewFleetRepository(),
		skillRepo: repository.NewEveSkillRepository(),
		planRepo:  repository.NewSkillPlanRepository(),
		ssoSvc:    newConfiguredEveSSOService(),
		esiClient: newConfiguredESIClient(),
	}
	svc.deliveryMailSender = svc.sendDeliveryMail
	return svc
}

// ─────────────────────────────────────────────
//  管理员端 - 福利定义 CRUD
// ─────────────────────────────────────────────

// AdminCreateWelfare 创建福利
func (s *WelfareService) AdminCreateWelfare(w *model.Welfare) error {
	if w.Name == "" {
		return errors.New("福利名称不能为空")
	}
	if w.DistMode != model.WelfareDistModePerUser && w.DistMode != model.WelfareDistModePerCharacter {
		return errors.New("无效的发放模式")
	}
	if w.PayByFuxiCoin != nil && *w.PayByFuxiCoin < 0 {
		return errors.New("伏羲币发放数量不能小于 0")
	}
	if w.RequireSkillPlan && len(w.SkillPlanIDs) == 0 {
		return errors.New("需要技能计划时必须选择至少一个技能计划")
	}

	skillPlanIDs := w.SkillPlanIDs
	if err := s.repo.CreateWelfare(w); err != nil {
		return err
	}

	if w.RequireSkillPlan && len(skillPlanIDs) > 0 {
		if err := s.repo.ReplaceWelfareSkillPlans(w.ID, skillPlanIDs); err != nil {
			return err
		}
	}
	w.SkillPlanIDs = skillPlanIDs
	if w.SkillPlanIDs == nil {
		w.SkillPlanIDs = []uint{}
	}
	return nil
}

// AdminUpdateWelfareRequest 更新福利请求
type AdminUpdateWelfareRequest struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	DistMode         string `json:"dist_mode"`
	PayByFuxiCoin    *int   `json:"pay_by_fuxi_coin"`
	RequireSkillPlan bool   `json:"require_skill_plan"`
	SkillPlanIDs     []uint `json:"skill_plan_ids"`
	MaxCharAgeMonths *int   `json:"max_char_age_months"`
	MinimumPap       *int   `json:"minimum_pap"`
	RequireEvidence  bool   `json:"require_evidence"`
	ExampleEvidence  string `json:"example_evidence"`
	Status           int8   `json:"status"`
	SortOrder        *int   `json:"sort_order"`
}

// AdminUpdateWelfare 更新福利
func (s *WelfareService) AdminUpdateWelfare(id uint, req *AdminUpdateWelfareRequest) (*model.Welfare, error) {
	w, err := s.repo.GetWelfareByID(id)
	if err != nil {
		return nil, errors.New("福利不存在")
	}

	if req.Name == "" {
		return nil, errors.New("福利名称不能为空")
	}
	if req.DistMode != model.WelfareDistModePerUser && req.DistMode != model.WelfareDistModePerCharacter {
		return nil, errors.New("无效的发放模式")
	}
	if req.PayByFuxiCoin != nil && *req.PayByFuxiCoin < 0 {
		return nil, errors.New("伏羲币发放数量不能小于 0")
	}
	if req.RequireSkillPlan && len(req.SkillPlanIDs) == 0 {
		return nil, errors.New("需要技能计划时必须选择至少一个技能计划")
	}

	w.Name = req.Name
	w.Description = req.Description
	w.DistMode = req.DistMode
	w.PayByFuxiCoin = req.PayByFuxiCoin
	w.RequireSkillPlan = req.RequireSkillPlan
	w.MaxCharAgeMonths = req.MaxCharAgeMonths
	w.MinimumPap = req.MinimumPap
	w.RequireEvidence = req.RequireEvidence
	w.ExampleEvidence = req.ExampleEvidence
	w.Status = req.Status
	if req.SortOrder != nil {
		w.SortOrder = *req.SortOrder
	}

	if err := s.repo.UpdateWelfare(w); err != nil {
		return nil, err
	}

	// 更新关联的技能计划
	var planIDs []uint
	if req.RequireSkillPlan {
		planIDs = req.SkillPlanIDs
	}
	if err := s.repo.ReplaceWelfareSkillPlans(w.ID, planIDs); err != nil {
		return nil, err
	}
	w.SkillPlanIDs = planIDs
	if w.SkillPlanIDs == nil {
		w.SkillPlanIDs = []uint{}
	}

	return w, nil
}

// AdminDeleteWelfare 删除福利（仅当无发放记录时允许）
func (s *WelfareService) AdminDeleteWelfare(id uint) error {
	if _, err := s.repo.GetWelfareByID(id); err != nil {
		return errors.New("福利不存在")
	}

	count, err := s.repo.CountApplicationsByWelfareID(id)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("该福利已有发放记录，无法删除")
	}

	return s.repo.DeleteWelfare(id)
}

// AdminListWelfares 查询福利列表
func (s *WelfareService) AdminListWelfares(page, pageSize int, filter repository.WelfareFilter) ([]model.Welfare, int64, error) {
	page = normalizePage(page)
	pageSize = normalizePageSize(pageSize, 20, 100)
	return s.repo.ListWelfares(page, pageSize, filter)
}

// AdminReorderWelfares 按给定 ID 顺序更新 sort_order
func (s *WelfareService) AdminReorderWelfares(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	updates := make([]repository.WelfareSortUpdate, len(ids))
	for i, id := range ids {
		updates[i] = repository.WelfareSortUpdate{ID: id, SortOrder: i}
	}
	return s.repo.UpdateWelfareSortOrders(updates)
}

// ─────────────────────────────────────────────
//  用户端 - 福利申请
// ─────────────────────────────────────────────

// EligibleCharacterResp 可申请人物
type EligibleCharacterResp struct {
	CharacterID      int64  `json:"character_id"`
	CharacterName    string `json:"character_name"`
	CanApplyNow      bool   `json:"can_apply_now"`
	IneligibleReason string `json:"ineligible_reason,omitempty"`
}

// EligibleWelfareResp 可申请福利
type EligibleWelfareResp struct {
	ID                 uint                    `json:"id"`
	Name               string                  `json:"name"`
	Description        string                  `json:"description"`
	DistMode           string                  `json:"dist_mode"`
	SkillPlanNames     []string                `json:"skill_plan_names"`
	RequireEvidence    bool                    `json:"require_evidence"`
	ExampleEvidence    string                  `json:"example_evidence"`
	CanApplyNow        bool                    `json:"can_apply_now"`
	IneligibleReason   string                  `json:"ineligible_reason,omitempty"`
	EligibleCharacters []EligibleCharacterResp `json:"eligible_characters"`
}

// buildIneligibleReason 根据不满足的条件构建原因字符串
// 可能的值："pap"、"skill"、"pap_skill"
func buildIneligibleReason(papBlocked bool, skillBlocked bool) string {
	if papBlocked && skillBlocked {
		return "pap_skill"
	}
	if papBlocked {
		return "pap"
	}
	if skillBlocked {
		return "skill"
	}
	return ""
}

func skillPlanNamesForWelfare(planIDs []uint, planNamesByID map[uint]string) []string {
	names := make([]string, 0, len(planIDs))
	for _, planID := range planIDs {
		name := strings.TrimSpace(planNamesByID[planID])
		if name == "" {
			continue
		}
		names = append(names, name)
	}
	return names
}

func (s *WelfareService) fillWelfareSkillPlanNames(welfares []model.Welfare) error {
	planIDSet := make(map[uint]struct{})
	for _, welfare := range welfares {
		for _, planID := range welfare.SkillPlanIDs {
			planIDSet[planID] = struct{}{}
		}
	}

	if len(planIDSet) == 0 {
		for index := range welfares {
			welfares[index].SkillPlanNames = []string{}
		}
		return nil
	}

	planIDs := make([]uint, 0, len(planIDSet))
	for planID := range planIDSet {
		planIDs = append(planIDs, planID)
	}

	plans, err := s.planRepo.ListByIDs(planIDs)
	if err != nil {
		return err
	}

	planNamesByID := make(map[uint]string, len(plans))
	for _, plan := range plans {
		planNamesByID[plan.ID] = strings.TrimSpace(plan.Title)
	}

	for index := range welfares {
		welfares[index].SkillPlanNames = skillPlanNamesForWelfare(
			welfares[index].SkillPlanIDs,
			planNamesByID,
		)
		if welfares[index].SkillPlanNames == nil {
			welfares[index].SkillPlanNames = []string{}
		}
	}

	return nil
}

// GetEligibleWelfares 获取用户可申请的福利列表
func (s *WelfareService) GetEligibleWelfares(userID uint) ([]EligibleWelfareResp, error) {
	// 1. 获取用户信息
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	// 2. 获取用户所有人物
	characters, err := s.charRepo.ListByUserID(userID)
	if err != nil {
		return nil, errors.New("获取人物列表失败")
	}

	// 3. 获取所有启用的福利
	welfares, err := s.repo.ListActiveWelfares()
	if err != nil {
		return nil, errors.New("获取福利列表失败")
	}
	if len(welfares) == 0 {
		return []EligibleWelfareResp{}, nil
	}
	if err := s.fillWelfareSkillPlanNames(welfares); err != nil {
		return nil, errors.New("获取技能计划失败")
	}

	// 4. 获取这些福利的所有申请记录
	welfareIDs := make([]uint, len(welfares))
	for i, w := range welfares {
		welfareIDs[i] = w.ID
	}
	allApps, err := s.repo.ListApplicationsByWelfareIDs(welfareIDs)
	if err != nil {
		return nil, errors.New("获取申请记录失败")
	}

	// 按 welfareID 分组申请记录
	appsByWelfare := make(map[uint][]model.WelfareApplication)
	for _, app := range allApps {
		appsByWelfare[app.WelfareID] = append(appsByWelfare[app.WelfareID], app)
	}

	// 5. 预加载技能检查数据（仅当有需要技能计划的福利时）
	skillCheckCache := make(map[int64]map[uint]bool) // characterID -> planID -> satisfied
	skillCheckReady := true
	needsSkillCheck := false
	for _, w := range welfares {
		if w.RequireSkillPlan && len(w.SkillPlanIDs) > 0 {
			needsSkillCheck = true
			break
		}
	}
	if needsSkillCheck && len(characters) > 0 {
		skillCheckCache, skillCheckReady = s.buildSkillCheckCache(characters, welfares)
	}

	// 5b. 预加载人物生日（仅当有人物年龄限制的福利时）
	needsAgeCheck := false
	for _, w := range welfares {
		if w.MaxCharAgeMonths != nil && *w.MaxCharAgeMonths > 0 {
			needsAgeCheck = true
			break
		}
	}
	if needsAgeCheck && len(characters) > 0 {
		s.ensureBirthdays(characters)
	}

	// 5c. 预加载军团 PAP 总数（仅当有最低 PAP 限制的福利时）
	needsMinimumPapCheck := false
	for _, w := range welfares {
		if w.MinimumPap != nil && *w.MinimumPap > 0 {
			needsMinimumPapCheck = true
			break
		}
	}
	var totalPap float64
	if needsMinimumPapCheck {
		total, err := s.fleetRepo.SumPapByUserTotal(user.ID)
		if err != nil {
			return nil, errors.New("获取 PAP 统计失败")
		}
		totalPap = total
	}

	now := time.Now()

	// 6. 计算每个福利的资格
	result := make([]EligibleWelfareResp, 0)
	for _, w := range welfares {
		apps := appsByWelfare[w.ID]
		if welfareAgeRestrictionFailed(characters, w.MaxCharAgeMonths, now) {
			continue
		}
		minimumPapBlocked := welfareMinimumPapRestrictionFailed(w.MinimumPap, totalPap)
		if w.RequireSkillPlan && len(w.SkillPlanIDs) > 0 && !skillCheckReady {
			continue
		}
		resp, ok := s.buildEligibleWelfareResp(user, characters, apps, w, skillCheckCache, minimumPapBlocked)
		if !ok {
			continue
		}
		result = append(result, resp)
	}

	return result, nil
}

// welfareAgeRestrictionFailed 检查福利的年龄限制是否阻止该用户申请
func welfareAgeRestrictionFailed(characters []model.EveCharacter, maxMonths *int, now time.Time) bool {
	if maxMonths == nil || *maxMonths <= 0 {
		return false
	}
	return anyCharacterTooOld(characters, *maxMonths, now)
}

// welfareMinimumPapRestrictionFailed checks the minimum PAP restriction.
func welfareMinimumPapRestrictionFailed(minimumPap *int, totalPap float64) bool {
	if minimumPap == nil || *minimumPap <= 0 {
		return false
	}
	return totalPap <= float64(*minimumPap)
}

// isUserIneligible 检查 per_user 福利中用户是否已申请过（通过 QQ 或 DiscordID 匹配）
func (s *WelfareService) isUserIneligible(user *model.User, apps []model.WelfareApplication) bool {
	userQQ := strings.TrimSpace(user.QQ)
	userDiscord := strings.TrimSpace(user.DiscordID)

	for _, app := range apps {
		if userQQ != "" && strings.TrimSpace(app.QQ) == userQQ {
			return true
		}
		if userDiscord != "" && strings.TrimSpace(app.DiscordID) == userDiscord {
			return true
		}
	}
	return false
}

// buildEligibleWelfareResp 组装单个福利的可申请状态
func (s *WelfareService) buildEligibleWelfareResp(
	user *model.User,
	characters []model.EveCharacter,
	apps []model.WelfareApplication,
	w model.Welfare,
	skillCheckCache map[int64]map[uint]bool,
	minimumPapBlocked bool,
) (EligibleWelfareResp, bool) {
	skillPlanNames := append([]string(nil), w.SkillPlanNames...)
	if skillPlanNames == nil {
		skillPlanNames = []string{}
	}

	resp := EligibleWelfareResp{
		ID:              w.ID,
		Name:            w.Name,
		Description:     w.Description,
		DistMode:        w.DistMode,
		SkillPlanNames:  skillPlanNames,
		RequireEvidence: w.RequireEvidence,
		ExampleEvidence: w.ExampleEvidence,
	}

	if w.DistMode == model.WelfareDistModePerUser {
		if s.isUserIneligible(user, apps) {
			return EligibleWelfareResp{}, false
		}
		if w.RequireSkillPlan && len(w.SkillPlanIDs) > 0 && len(characters) == 0 {
			return EligibleWelfareResp{}, false
		}
		skillBlocked := false
		if w.RequireSkillPlan && len(w.SkillPlanIDs) > 0 {
			skillBlocked = !s.anyCharacterSatisfiesSkillPlan(characters, w.SkillPlanIDs, skillCheckCache)
		}
		resp.CanApplyNow = !minimumPapBlocked && !skillBlocked
		if !resp.CanApplyNow {
			resp.IneligibleReason = buildIneligibleReason(minimumPapBlocked, skillBlocked)
		}
		return resp, true
	}

	eligible := s.filterEligibleCharacters(characters, apps, w, skillCheckCache, minimumPapBlocked)
	if len(eligible) == 0 {
		return EligibleWelfareResp{}, false
	}
	resp.EligibleCharacters = eligible
	return resp, true
}

// filterEligibleCharacters 过滤 per_character 福利中可见的人物
func (s *WelfareService) filterEligibleCharacters(
	characters []model.EveCharacter,
	apps []model.WelfareApplication,
	w model.Welfare,
	skillCheckCache map[int64]map[uint]bool,
	minimumPapBlocked bool,
) []EligibleCharacterResp {
	// 构建已申请的人物集合
	appliedCharIDs := make(map[int64]bool)
	appliedCharNames := make(map[string]bool)
	for _, app := range apps {
		appliedCharIDs[app.CharacterID] = true
		if name := strings.TrimSpace(app.CharacterName); name != "" {
			appliedCharNames[name] = true
		}
	}

	var eligible []EligibleCharacterResp
	for _, char := range characters {
		// 已申请过的人物跳过
		if appliedCharIDs[char.CharacterID] || appliedCharNames[strings.TrimSpace(char.CharacterName)] {
			continue
		}
		skillBlocked := false
		if w.RequireSkillPlan && len(w.SkillPlanIDs) > 0 {
			skillBlocked = !s.characterSatisfiesAnySkillPlan(char.CharacterID, w.SkillPlanIDs, skillCheckCache)
		}
		canApplyNow := !minimumPapBlocked && !skillBlocked
		charResp := EligibleCharacterResp{
			CharacterID:   char.CharacterID,
			CharacterName: char.CharacterName,
			CanApplyNow:   canApplyNow,
		}
		if !canApplyNow {
			charResp.IneligibleReason = buildIneligibleReason(minimumPapBlocked, skillBlocked)
		}
		eligible = append(eligible, charResp)
	}
	return eligible
}

// buildSkillCheckCache 批量构建人物技能计划满足状态缓存
func (s *WelfareService) buildSkillCheckCache(
	characters []model.EveCharacter,
	welfares []model.Welfare,
) (map[int64]map[uint]bool, bool) {
	// 收集所有需要的技能计划 ID
	planIDSet := make(map[uint]struct{})
	for _, w := range welfares {
		if w.RequireSkillPlan {
			for _, pid := range w.SkillPlanIDs {
				planIDSet[pid] = struct{}{}
			}
		}
	}
	planIDs := make([]uint, 0, len(planIDSet))
	for pid := range planIDSet {
		planIDs = append(planIDs, pid)
	}

	// 获取技能计划的技能要求
	planSkills, err := s.planRepo.ListSkillsByPlanIDs(planIDs)
	if err != nil {
		return nil, false
	}
	planSkillsMap := make(map[uint][]model.SkillPlanSkill)
	for _, skill := range planSkills {
		planSkillsMap[skill.SkillPlanID] = append(planSkillsMap[skill.SkillPlanID], skill)
	}

	// 为每个人物检查每个计划
	cache := make(map[int64]map[uint]bool)
	for _, char := range characters {
		skills, err := s.skillRepo.GetSkillList(int(char.CharacterID))
		if err != nil {
			continue
		}
		levelMap := buildCharacterSkillLevelMap(skills)

		cache[char.CharacterID] = make(map[uint]bool)
		for _, pid := range planIDs {
			requirements := planSkillsMap[pid]
			satisfied := true
			for _, req := range requirements {
				if levelMap[req.SkillTypeID] < req.RequiredLevel {
					satisfied = false
					break
				}
			}
			cache[char.CharacterID][pid] = satisfied
		}
	}
	return cache, true
}

// anyCharacterSatisfiesSkillPlan 检查是否有任一人物满足任一技能计划
func (s *WelfareService) anyCharacterSatisfiesSkillPlan(
	characters []model.EveCharacter,
	planIDs []uint,
	cache map[int64]map[uint]bool,
) bool {
	for _, char := range characters {
		if s.characterSatisfiesAnySkillPlan(char.CharacterID, planIDs, cache) {
			return true
		}
	}
	return false
}

// characterSatisfiesAnySkillPlan 检查人物是否满足任一技能计划
func (s *WelfareService) characterSatisfiesAnySkillPlan(
	characterID int64,
	planIDs []uint,
	cache map[int64]map[uint]bool,
) bool {
	charCache, ok := cache[characterID]
	if !ok {
		return false
	}
	for _, pid := range planIDs {
		if charCache[pid] {
			return true
		}
	}
	return false
}

// ApplyForWelfareRequest 申请福利请求
type ApplyForWelfareRequest struct {
	WelfareID     uint   `json:"welfare_id"`
	CharacterID   int64  `json:"character_id"`
	EvidenceImage string `json:"evidence_image"`
}

func initialWelfareApplicationRequestedStatus() string {
	return model.WelfareAppStatusRequested
}

// ApplyForWelfare 申请福利
func (s *WelfareService) ApplyForWelfare(userID uint, req *ApplyForWelfareRequest) (*model.WelfareApplication, error) {
	// 获取福利
	welfare, err := s.repo.GetWelfareByID(req.WelfareID)
	if err != nil {
		return nil, errors.New("福利不存在")
	}
	if welfare.Status != model.WelfareStatusActive {
		return nil, errors.New("该福利未启用")
	}

	// 获取用户信息
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	// per_user 需要至少一个联系方式
	if welfare.DistMode == model.WelfareDistModePerUser {
		if strings.TrimSpace(user.QQ) == "" && strings.TrimSpace(user.DiscordID) == "" {
			return nil, errors.New("请先设置QQ或Discord联系方式")
		}
	}

	// 获取用户人物
	characters, err := s.charRepo.ListByUserID(userID)
	if err != nil {
		return nil, errors.New("获取人物列表失败")
	}

	// 获取该福利的所有申请记录
	apps, err := s.repo.ListApplicationsByWelfareIDs([]uint{welfare.ID})
	if err != nil {
		return nil, errors.New("获取申请记录失败")
	}

	if welfare.MaxCharAgeMonths != nil && *welfare.MaxCharAgeMonths > 0 {
		s.ensureBirthdays(characters)
	}

	var selectedChar model.EveCharacter

	if welfare.DistMode == model.WelfareDistModePerUser {
		// per_user: 检查用户是否已申请
		if s.isUserIneligible(user, apps) {
			return nil, errors.New("您已申请过该福利")
		}
		// 找主人物或第一个人物
		for _, c := range characters {
			if c.CharacterID == user.PrimaryCharacterID {
				selectedChar = c
				break
			}
		}
		if selectedChar.CharacterID == 0 && len(characters) > 0 {
			selectedChar = characters[0]
		}
	} else {
		// per_character: 必须指定人物
		if req.CharacterID == 0 {
			return nil, errors.New("按人物模式必须指定人物")
		}
		// 验证人物属于用户
		found := false
		for _, c := range characters {
			if c.CharacterID == req.CharacterID {
				selectedChar = c
				found = true
				break
			}
		}
		if !found {
			return nil, errors.New("该人物不属于您")
		}
		// 检查人物是否已申请
		for _, app := range apps {
			if app.CharacterID == req.CharacterID || strings.TrimSpace(app.CharacterName) == strings.TrimSpace(selectedChar.CharacterName) {
				return nil, errors.New("该人物已申请过该福利")
			}
		}
	}

	// 人物年龄检查：任一人物超龄则该福利不可申请
	if welfareAgeRestrictionFailed(characters, welfare.MaxCharAgeMonths, time.Now()) {
		return nil, errors.New("您的人物年龄超过该福利限制")
	}

	// PAP 检查：军团 PAP 总数必须严格大于最低要求
	if welfare.MinimumPap != nil && *welfare.MinimumPap > 0 {
		totalPap, err := s.fleetRepo.SumPapByUserTotal(userID)
		if err != nil {
			return nil, errors.New("获取 PAP 统计失败")
		}
		if welfareMinimumPapRestrictionFailed(welfare.MinimumPap, totalPap) {
			return nil, errors.New("您的军团 PAP 未达到该福利限制")
		}
	}

	// 证明图片检查
	if welfare.RequireEvidence && strings.TrimSpace(req.EvidenceImage) == "" {
		return nil, errors.New("该福利需要上传证明图片")
	}

	// 技能计划检查
	if welfare.RequireSkillPlan {
		// 填充 SkillPlanIDs (GetWelfareByID 不会自动填充)
		planIDs, err := s.repo.GetSkillPlanIDsByWelfareID(welfare.ID)
		if err != nil {
			return nil, errors.New("获取技能计划失败")
		}
		welfare.SkillPlanIDs = planIDs

		if len(welfare.SkillPlanIDs) > 0 {
			welfares := []model.Welfare{*welfare}
			cache, ok := s.buildSkillCheckCache(characters, welfares)
			if !ok {
				return nil, errors.New("获取技能计划失败")
			}

			if welfare.DistMode == model.WelfareDistModePerUser {
				if !s.anyCharacterSatisfiesSkillPlan(characters, welfare.SkillPlanIDs, cache) {
					return nil, errors.New("您的人物不满足技能计划要求")
				}
			} else {
				if !s.characterSatisfiesAnySkillPlan(selectedChar.CharacterID, welfare.SkillPlanIDs, cache) {
					return nil, errors.New("该人物不满足技能计划要求")
				}
			}
		}
	}

	// 创建申请
	app := &model.WelfareApplication{
		WelfareID:     welfare.ID,
		UserID:        &userID,
		CharacterID:   selectedChar.CharacterID,
		CharacterName: selectedChar.CharacterName,
		QQ:            user.QQ,
		DiscordID:     user.DiscordID,
		EvidenceImage: req.EvidenceImage,
		Status:        initialWelfareApplicationRequestedStatus(),
	}

	if err := s.repo.CreateApplication(app); err != nil {
		return nil, errors.New("申请失败")
	}
	return app, nil
}

// ─────────────────────────────────────────────
//  管理端 - 福利审批
// ─────────────────────────────────────────────

// AdminApplicationResp 管理端福利申请记录响应
type AdminApplicationResp struct {
	ID                uint       `json:"id"`
	WelfareID         uint       `json:"welfare_id"`
	WelfareName       string     `json:"welfare_name"`
	WelfareDesc       string     `json:"welfare_description"`
	UserID            *uint      `json:"user_id"`
	ApplicantNickname string     `json:"applicant_nickname"`
	CharacterName     string     `json:"character_name"`
	QQ                string     `json:"qq"`
	DiscordID         string     `json:"discord_id"`
	EvidenceImage     string     `json:"evidence_image"`
	Status            string     `json:"status"`
	ReviewedBy        uint       `json:"reviewed_by"`
	ReviewerName      string     `json:"reviewer_name"`
	CreatedAt         time.Time  `json:"created_at"`
	ReviewedAt        *time.Time `json:"reviewed_at"`
}

// AdminListApplications 管理端查询福利申请列表
func (s *WelfareService) AdminListApplications(page, pageSize int, filter repository.WelfareApplicationFilter) ([]AdminApplicationResp, int64, error) {
	page = normalizePage(page)
	pageSize = normalizeLedgerPageSize(pageSize)

	apps, total, err := s.repo.ListApplicationsPaginated(page, pageSize, filter)
	if err != nil {
		return nil, 0, err
	}
	if len(apps) == 0 {
		return []AdminApplicationResp{}, 0, nil
	}

	// 批量收集需要的 welfare IDs 和 user IDs
	welfareIDSet := make(map[uint]struct{})
	userIDSet := make(map[uint]struct{})
	for _, app := range apps {
		welfareIDSet[app.WelfareID] = struct{}{}
		if app.UserID != nil {
			userIDSet[*app.UserID] = struct{}{}
		}
		if app.ReviewedBy > 0 {
			userIDSet[app.ReviewedBy] = struct{}{}
		}
	}

	// 批量获取 welfare 信息
	welfareIDs := make([]uint, 0, len(welfareIDSet))
	for wid := range welfareIDSet {
		welfareIDs = append(welfareIDs, wid)
	}
	welfareMap := make(map[uint]model.Welfare)
	if len(welfareIDs) > 0 {
		welfares, err := s.repo.ListWelfaresByIDs(welfareIDs)
		if err == nil {
			for index := range welfares {
				welfare := welfares[index]
				welfareMap[welfare.ID] = welfare
			}
		}
	}

	// 批量获取 user 昵称
	userIDs := make([]uint, 0, len(userIDSet))
	for uid := range userIDSet {
		userIDs = append(userIDs, uid)
	}
	nicknameMap := make(map[uint]string)
	if len(userIDs) > 0 {
		users, err := s.userRepo.ListByIDs(userIDs)
		if err == nil {
			for _, user := range users {
				nicknameMap[user.ID] = user.Nickname
			}
		}
	}

	result := make([]AdminApplicationResp, 0, len(apps))
	for _, app := range apps {
		resp := AdminApplicationResp{
			ID:            app.ID,
			WelfareID:     app.WelfareID,
			UserID:        app.UserID,
			CharacterName: app.CharacterName,
			QQ:            app.QQ,
			DiscordID:     app.DiscordID,
			EvidenceImage: app.EvidenceImage,
			Status:        app.Status,
			ReviewedBy:    app.ReviewedBy,
			CreatedAt:     app.CreatedAt,
			ReviewedAt:    app.ReviewedAt,
		}
		if welfare, ok := welfareMap[app.WelfareID]; ok {
			resp.WelfareName = welfare.Name
			resp.WelfareDesc = welfare.Description
		}
		if app.UserID != nil {
			resp.ApplicantNickname = nicknameMap[*app.UserID]
		}
		if app.ReviewedBy > 0 {
			resp.ReviewerName = nicknameMap[app.ReviewedBy]
		}
		result = append(result, resp)
	}

	return result, total, nil
}

// AdminReviewApplicationRequest 审批请求
type AdminReviewApplicationRequest struct {
	Action string `json:"action"` // "deliver" or "reject"
}

// validateReviewTransition 验证审批状态转换是否合法，返回目标状态
func validateReviewTransition(currentStatus, action string) (string, error) {
	switch action {
	case "deliver":
		if currentStatus != model.WelfareAppStatusRequested {
			return "", errors.New("只能对待发放的申请进行发放操作")
		}
		return model.WelfareAppStatusDelivered, nil
	case "reject":
		if currentStatus != model.WelfareAppStatusRequested {
			return "", errors.New("只能对待发放的申请进行拒绝操作")
		}
		return model.WelfareAppStatusRejected, nil
	default:
		return "", errors.New("无效的审批操作")
	}
}

// AdminReviewApplication 管理端审批福利申请
// AdminDeleteApplication 删除单条福利申请记录
func (s *WelfareService) AdminDeleteApplication(id uint) error {
	if _, err := s.repo.GetApplicationByID(id); err != nil {
		return errors.New("申请记录不存在")
	}
	return s.repo.DeleteApplication(id)
}

func (s *WelfareService) AdminReviewApplication(appID uint, reviewerID uint, req *AdminReviewApplicationRequest) (MailAttemptSummary, error) {
	var deliveredWelfare *model.Welfare
	var deliveredApp *model.WelfareApplication

	err := global.DB.Transaction(func(tx *gorm.DB) error {
		app, err := s.repo.GetApplicationByIDForUpdateTx(tx, appID)
		if err != nil {
			return errors.New("申请记录不存在")
		}

		newStatus, err := validateReviewTransition(app.Status, strings.TrimSpace(req.Action))
		if err != nil {
			return err
		}
		app.Status = newStatus

		now := time.Now()
		app.ReviewedBy = reviewerID
		app.ReviewedAt = &now

		if newStatus == model.WelfareAppStatusDelivered {
			welfare, err := s.repo.GetWelfareByIDTx(tx, app.WelfareID)
			if err != nil {
				return errors.New("福利不存在")
			}
			if welfare.PayByFuxiCoin != nil && *welfare.PayByFuxiCoin > 0 {
				if app.UserID == nil || *app.UserID == 0 {
					return errors.New("该福利配置了伏羲币发放，但申请记录缺少用户信息")
				}
				reason := fmt.Sprintf("Welfare#%d Application#%d %s", welfare.ID, app.ID, welfare.Name)
				refID := fmt.Sprintf("welfare_application:%d", app.ID)
				if err := NewSysWalletService().ApplyWalletDeltaByOperatorTx(
					tx,
					*app.UserID,
					reviewerID,
					float64(*welfare.PayByFuxiCoin),
					reason,
					model.WalletRefWelfarePayout,
					refID,
				); err != nil {
					return err
				}
			}

			appCopy := *app
			welfareCopy := *welfare
			deliveredApp = &appCopy
			deliveredWelfare = &welfareCopy
		}

		return s.repo.UpdateApplicationTx(tx, app)
	})
	if err != nil {
		return MailAttemptSummary{}, err
	}

	return s.attemptDeliveryMail(reviewerID, deliveredWelfare, deliveredApp), nil
}

func (s *WelfareService) attemptDeliveryMail(reviewerID uint, deliveredWelfare *model.Welfare, deliveredApp *model.WelfareApplication) MailAttemptSummary {
	if deliveredWelfare == nil || deliveredApp == nil || s.deliveryMailSender == nil {
		return MailAttemptSummary{}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	summary, err := s.deliveryMailSender(ctx, reviewerID, deliveredWelfare, deliveredApp)
	if err != nil {
		if global.Logger != nil {
			global.Logger.Warn("福利发放后邮件尝试失败",
				zap.Uint("reviewer_user_id", reviewerID),
				zap.Uint("welfare_id", deliveredWelfare.ID),
				zap.Uint("application_id", deliveredApp.ID),
				zap.Error(err),
			)
		}
		return summary.withError(err)
	}
	return summary
}

func (s *WelfareService) sendDeliveryMail(
	ctx context.Context,
	reviewerID uint,
	deliveredWelfare *model.Welfare,
	deliveredApp *model.WelfareApplication,
) (MailAttemptSummary, error) {
	if deliveredWelfare == nil || deliveredApp == nil {
		return MailAttemptSummary{}, nil
	}
	summary := MailAttemptSummary{}
	if deliveredApp.UserID == nil || *deliveredApp.UserID == 0 {
		return summary, errors.New("福利申请缺少收件用户信息")
	}

	mailSupport := newInGameMailSupport(s.userRepo, s.charRepo, s.ssoSvc, s.esiClient)
	sender, err := mailSupport.resolveSender(ctx, reviewerID)
	summary.MailSenderCharacterID = sender.CharacterID
	summary.MailSenderCharacterName = sender.CharacterName
	if err != nil {
		return summary, err
	}

	recipient, err := mailSupport.resolveUserPrimaryCharacter(*deliveredApp.UserID)
	summary.MailRecipientCharacterID = recipient.CharacterID
	summary.MailRecipientCharacterName = recipient.CharacterName
	if err != nil {
		return summary, err
	}

	subject, body := buildWelfareDeliveryMailContent(deliveredWelfare.Name, sender.DisplayName)
	mailID, err := mailSupport.send(ctx, sender.CharacterID, sender.AccessToken, recipient.CharacterID, subject, body)
	summary.MailID = mailID
	return summary, err
}

func buildWelfareDeliveryMailContent(welfareName, officerDisplayName string) (string, string) {
	welfareName = strings.TrimSpace(welfareName)
	if welfareName == "" {
		welfareName = "福利"
	}
	officerDisplayName = strings.TrimSpace(officerDisplayName)
	if officerDisplayName == "" {
		officerDisplayName = "Officer"
	}

	subject := fmt.Sprintf("福利发放通知 / Welfare Delivery Notice %s", welfareName)
	var bodyBuilder strings.Builder
	bodyBuilder.WriteString("你好，\n\n")
	fmt.Fprintf(&bodyBuilder, "你的福利「%s」已由福利官 %s 发放。\n", welfareName, officerDisplayName)
	fmt.Fprintf(&bodyBuilder, "福利名称：%s\n", welfareName)
	bodyBuilder.WriteString("请检查你的伏羲币钱包或合同。\n")
	bodyBuilder.WriteString("如有疑问，请联系处理此申请的福利官。\n")
	bodyBuilder.WriteString("感谢你的支持，祝你飞行顺利。\n")
	bodyBuilder.WriteString("================\n\n")
	bodyBuilder.WriteString("Hello,\n\n")
	fmt.Fprintf(&bodyBuilder, "Your welfare \"%s\" has been delivered by officer %s.\n", welfareName, officerDisplayName)
	fmt.Fprintf(&bodyBuilder, "Welfare: %s\n", welfareName)
	bodyBuilder.WriteString("Please check your FuxiCoin wallet or contract.\n")
	bodyBuilder.WriteString("If anything looks incorrect, please contact the officer who handled this delivery.\n")
	bodyBuilder.WriteString("Thank you for your support, and fly safe.\n")

	return subject, bodyBuilder.String()
}

// MyApplicationResp 用户申请记录响应
type MyApplicationResp struct {
	ID            uint       `json:"id"`
	WelfareID     uint       `json:"welfare_id"`
	WelfareName   string     `json:"welfare_name"`
	CharacterName string     `json:"character_name"`
	Status        string     `json:"status"`
	ReviewerName  string     `json:"reviewer_name"`
	CreatedAt     time.Time  `json:"created_at"`
	ReviewedAt    *time.Time `json:"reviewed_at"`
}

func buildMyApplicationResponses(
	apps []model.WelfareApplication,
	welfareNames map[uint]string,
	reviewerNames map[uint]string,
) []MyApplicationResp {
	result := make([]MyApplicationResp, 0, len(apps))
	for _, app := range apps {
		reviewerName := ""
		if app.ReviewedBy > 0 {
			reviewerName = reviewerNames[app.ReviewedBy]
		}
		result = append(result, MyApplicationResp{
			ID:            app.ID,
			WelfareID:     app.WelfareID,
			WelfareName:   welfareNames[app.WelfareID],
			CharacterName: app.CharacterName,
			Status:        app.Status,
			ReviewerName:  reviewerName,
			CreatedAt:     app.CreatedAt,
			ReviewedAt:    app.ReviewedAt,
		})
	}
	return result
}

// ListMyApplications 查询用户的福利申请列表
func (s *WelfareService) ListMyApplications(userID uint, page, pageSize int, status string) ([]MyApplicationResp, int64, error) {
	page = normalizePage(page)
	pageSize = normalizePageSize(pageSize, 10, 100)

	apps, total, err := s.repo.ListApplicationsByUserIDPaginated(userID, page, pageSize, status)
	if err != nil {
		return nil, 0, errors.New("获取申请记录失败")
	}

	if len(apps) == 0 {
		return []MyApplicationResp{}, total, nil
	}

	welfareIDSet := make(map[uint]struct{})
	reviewerIDSet := make(map[uint]struct{})
	for _, app := range apps {
		welfareIDSet[app.WelfareID] = struct{}{}
		if app.ReviewedBy > 0 {
			reviewerIDSet[app.ReviewedBy] = struct{}{}
		}
	}

	welfareIDs := make([]uint, 0, len(welfareIDSet))
	for welfareID := range welfareIDSet {
		welfareIDs = append(welfareIDs, welfareID)
	}

	welfareNames := make(map[uint]string)
	if len(welfareIDs) > 0 {
		welfares, err := s.repo.ListWelfaresByIDs(welfareIDs)
		if err == nil {
			for _, welfare := range welfares {
				welfareNames[welfare.ID] = welfare.Name
			}
		}
	}

	reviewerIDs := make([]uint, 0, len(reviewerIDSet))
	for reviewerID := range reviewerIDSet {
		reviewerIDs = append(reviewerIDs, reviewerID)
	}

	reviewerNames := make(map[uint]string)
	if len(reviewerIDs) > 0 {
		users, err := s.userRepo.ListByIDs(reviewerIDs)
		if err == nil {
			for _, user := range users {
				reviewerNames[user.ID] = user.Nickname
			}
		}
	}

	return buildMyApplicationResponses(apps, welfareNames, reviewerNames), total, nil
}

// ─────────────────────────────────────────────
//  管理端 - 导入历史记录
// ─────────────────────────────────────────────

// ImportWelfareRecordsRequest 导入历史记录请求
type ImportWelfareRecordsRequest struct {
	WelfareID uint   `json:"welfare_id"`
	CSV       string `json:"csv"`
}

func parseImportedWelfareApplications(welfareID uint, csvText string) ([]model.WelfareApplication, error) {
	normalizedCSV := strings.ReplaceAll(csvText, "\r\n", "\n")
	normalizedCSV = strings.ReplaceAll(normalizedCSV, "\r", "\n")

	lines := strings.Split(normalizedCSV, "\n")
	apps := make([]model.WelfareApplication, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var parts []string
		if strings.Contains(line, "\t") {
			parts = strings.SplitN(line, "\t", 2)
		} else {
			parts = strings.SplitN(line, ",", 2)
		}

		characterName := strings.TrimSpace(parts[0])
		if characterName == "" {
			continue
		}

		qq := ""
		if len(parts) > 1 {
			qq = strings.TrimSpace(parts[1])
		}

		apps = append(apps, model.WelfareApplication{
			WelfareID:     welfareID,
			CharacterName: characterName,
			QQ:            qq,
			Status:        model.WelfareAppStatusDelivered,
		})
	}

	if len(apps) == 0 {
		return nil, errors.New("未解析到有效记录")
	}

	return apps, nil
}

// ImportWelfareRecords 解析 CSV 文本并批量创建福利申请记录（历史导入）
func (s *WelfareService) ImportWelfareRecords(req *ImportWelfareRecordsRequest) (int, error) {
	if req.WelfareID == 0 {
		return 0, errors.New("福利 ID 不能为空")
	}
	if strings.TrimSpace(req.CSV) == "" {
		return 0, errors.New("CSV 内容不能为空")
	}

	// 验证福利存在
	if _, err := s.repo.GetWelfareByID(req.WelfareID); err != nil {
		return 0, errors.New("福利不存在")
	}

	apps, err := parseImportedWelfareApplications(req.WelfareID, req.CSV)
	if err != nil {
		return 0, err
	}

	if err := s.repo.BulkCreateApplications(apps); err != nil {
		return 0, fmt.Errorf("导入失败: %w", err)
	}
	return len(apps), nil
}

// ─────────────────────────────────────────────
//  人物年龄检查
// ─────────────────────────────────────────────

// characterAgeTooOld 检查人物年龄是否超过限制月数
// 返回 true 表示人物太老（不符合资格）
func characterAgeTooOld(birthday *time.Time, maxMonths int, now time.Time) bool {
	if birthday == nil {
		return false // 未知生日不限制
	}
	cutoff := now.AddDate(0, -maxMonths, 0)
	return birthday.Before(cutoff)
}

// anyCharacterTooOld 检查用户是否拥有任何年龄超限的人物
func anyCharacterTooOld(characters []model.EveCharacter, maxMonths int, now time.Time) bool {
	for _, c := range characters {
		if characterAgeTooOld(c.Birthday, maxMonths, now) {
			return true
		}
	}
	return false
}

// esiCharacterPublicInfo ESI GET /characters/{id}/ 公开信息（只取 birthday）
type esiCharacterPublicInfo struct {
	Birthday string `json:"birthday"`
}

// ensureBirthdays 确保人物列表中的 Birthday 字段已填充，缺失的从 ESI 获取并持久化
func (s *WelfareService) ensureBirthdays(characters []model.EveCharacter) {
	for i := range characters {
		if characters[i].Birthday != nil {
			continue
		}
		birthday := s.fetchBirthdayFromESI(characters[i].CharacterID)
		if birthday == nil {
			continue
		}
		characters[i].Birthday = birthday
		// 持久化
		if dbChar, err := s.charRepo.GetByCharacterID(characters[i].CharacterID); err == nil {
			dbChar.Birthday = birthday
			_ = s.charRepo.Save(dbChar)
		}
	}
}

// fetchBirthdayFromESI 从 ESI 公开接口获取人物生日
func (s *WelfareService) fetchBirthdayFromESI(characterID int64) *time.Time {
	url := fmt.Sprintf("https://esi.evetech.net/latest/characters/%d/?datasource=tranquility", characterID)
	resp, err := http.Get(url) //nolint:gosec // ESI public endpoint, character ID is trusted internal data
	if err != nil {
		return nil
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return nil
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	var info esiCharacterPublicInfo
	if err := json.Unmarshal(body, &info); err != nil || info.Birthday == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, info.Birthday)
	if err != nil {
		return nil
	}
	return &t
}
