package service

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"amiya-eden/pkg/eve/esi"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// FleetKMRefreshFunc 触发单个角色 KM 刷新的钩子，由 jobs 层注入以避免循环依赖
var FleetKMRefreshFunc func(characterID int64)

// FleetAutoSRPFunc 自动 SRP 处理钩子，由 jobs 层注入
var FleetAutoSRPFunc func(fleetID string)

// FleetService 舰队业务逻辑层
type FleetService struct {
	repo       *repository.FleetRepository
	papRepo    *repository.AlliancePAPRepository
	userRepo   *repository.UserRepository
	charRepo   *repository.EveCharacterRepository
	configRepo *repository.SysConfigRepository
	walletRepo *repository.SysWalletRepository
	ssoSvc     *EveSSOService
	walletSvc  *SysWalletService
	webhookSvc *WebhookService
	rateRepo   *repository.PAPTypeRateRepository
	esiClient  *esi.Client
}

func NewFleetService() *FleetService {
	return &FleetService{
		repo:       repository.NewFleetRepository(),
		papRepo:    repository.NewAlliancePAPRepository(),
		userRepo:   repository.NewUserRepository(),
		charRepo:   repository.NewEveCharacterRepository(),
		configRepo: repository.NewSysConfigRepository(),
		walletRepo: repository.NewSysWalletRepository(),
		ssoSvc:     NewEveSSOService(),
		walletSvc:  NewSysWalletService(),
		webhookSvc: NewWebhookService(),
		rateRepo:   repository.NewPAPTypeRateRepository(),
		esiClient:  esi.NewClientWithConfig(global.Config.EveSSO.ESIBaseURL, global.Config.EveSSO.ESIAPIPrefix),
	}
}

const (
	CorporationPapPeriodCurrentMonth = "current_month"
	CorporationPapPeriodLastMonth    = "last_month"
	CorporationPapPeriodAtYear       = "at_year"
	CorporationPapPeriodAll          = "all"
)

// CorporationPapSummaryItem 军团 PAP 汇总项
type CorporationPapSummaryItem struct {
	UserID            uint    `json:"user_id"`
	Nickname          string  `json:"nickname"`
	CorpTicker        string  `json:"corp_ticker"`
	MainCharacterName string  `json:"main_character_name"`
	CharacterCount    int     `json:"character_count"`
	StratOpPaps       float64 `json:"strat_op_paps"`
	SkirmishPaps      float64 `json:"skirmish_paps"`
	AllianceStratPaps float64 `json:"alliance_strat_paps"`
}

// CorporationPapOverview 军团 PAP 页面顶部概览
type CorporationPapOverview struct {
	FilteredPapTotal     float64 `json:"filtered_pap_total"`
	FilteredStratOpTotal float64 `json:"filtered_strat_op_total"`
	AllPapTotal          float64 `json:"all_pap_total"`
	FilteredUserCount    int64   `json:"filtered_user_count"`
	Period               string  `json:"period"`
	Year                 *int    `json:"year,omitempty"`
}

// CorporationPapSummaryResponse 军团 PAP 汇总响应
type CorporationPapSummaryResponse struct {
	List     []CorporationPapSummaryItem `json:"list"`
	Total    int64                       `json:"total"`
	Page     int                         `json:"page"`
	PageSize int                         `json:"pageSize"`
	Overview CorporationPapOverview      `json:"overview"`
}

// ─────────────────────────────────────────────
//  舰队 CRUD
// ─────────────────────────────────────────────

// CreateFleetRequest 创建舰队请求
type CreateFleetRequest struct {
	Title         string  `json:"title" binding:"required"`
	Description   string  `json:"description"`
	StartAt       string  `json:"start_at" binding:"required"` // RFC3339
	EndAt         string  `json:"end_at" binding:"required"`   // RFC3339
	Importance    string  `json:"importance" binding:"required,oneof=strat_op cta other"`
	PapCount      float64 `json:"pap_count"`
	CharacterID   int64   `json:"character_id" binding:"required"` // FC 角色 ID
	SendPing      bool    `json:"send_ping"`                       // 是否发送 Ping 通知
	FleetConfigID *uint   `json:"fleet_config_id"`                 // 舰队配置 ID
	AutoSrpMode   string  `json:"auto_srp_mode"`                   // disabled/submit_only/auto_approve
}

// CreateFleet 创建舰队
func (s *FleetService) CreateFleet(userID uint, req *CreateFleetRequest) (*model.Fleet, error) {
	// 验证角色属于当前用户
	char, err := s.charRepo.GetByCharacterID(req.CharacterID)
	if err != nil {
		return nil, errors.New("角色不存在")
	}
	if char.UserID != userID {
		return nil, errors.New("该角色不属于当前用户")
	}

	startAt, err := time.Parse(time.RFC3339, req.StartAt)
	if err != nil {
		return nil, errors.New("起始时间格式错误（需 RFC3339）")
	}
	endAt, err := time.Parse(time.RFC3339, req.EndAt)
	if err != nil {
		return nil, errors.New("结束时间格式错误（需 RFC3339）")
	}

	if endAt.Before(startAt) {
		return nil, errors.New("结束时间不能早于起始时间")
	}

	fleet := &model.Fleet{
		ID:              uuid.New().String(),
		Title:           req.Title,
		Description:     req.Description,
		StartAt:         startAt,
		EndAt:           endAt,
		Importance:      req.Importance,
		PapCount:        req.PapCount,
		FCUserID:        userID,
		FCCharacterID:   req.CharacterID,
		FCCharacterName: char.CharacterName,
		FleetConfigID:   req.FleetConfigID,
		AutoSrpMode:     normalizeAutoSrpMode(req.AutoSrpMode),
	}

	// 自动从 ESI 拉取当前 ESI 舰队 ID
	ctx := context.Background()
	if accessToken, tokenErr := s.ssoSvc.GetValidToken(ctx, req.CharacterID); tokenErr == nil {
		path := fmt.Sprintf("/characters/%d/fleet/", req.CharacterID)
		var info CharacterFleetInfo
		if esiErr := s.esiGet(ctx, path, accessToken, &info); esiErr == nil {
			fleet.ESIFleetID = &info.FleetID
		} else {
			global.Logger.Warn("CreateFleet: 拉取 ESI fleet_id 失败", zap.Error(esiErr))
		}
	} else {
		global.Logger.Warn("CreateFleet: 获取 Token 失败，跳过 ESI fleet_id", zap.Error(tokenErr))
	}

	if err := s.repo.Create(fleet); err != nil {
		return nil, err
	}

	// 确保 FC 用户有钱包
	_, _ = s.walletSvc.GetMyWallet(userID)

	// 异步发送 Webhook Ping
	if req.SendPing {
		go func() {
			if pingErr := s.webhookSvc.SendFleetPing(fleet); pingErr != nil {
				global.Logger.Warn("CreateFleet: webhook ping 失败", zap.Error(pingErr))
			}
		}()
	}

	return fleet, nil
}

// PingFleet 手动触发舰队 Ping（仅 FC 或管理员）
func (s *FleetService) PingFleet(fleetID string, userID uint, userRoles []string) error {
	fleet, err := s.repo.GetByID(fleetID)
	if err != nil {
		return errors.New("舰队不存在")
	}
	if !s.canManageFleet(fleet, userID, userRoles) {
		return errors.New("权限不足")
	}
	return s.webhookSvc.SendFleetPing(fleet)
}

// UpdateFleetRequest 更新舰队请求
type UpdateFleetRequest struct {
	Title         *string  `json:"title"`
	Description   *string  `json:"description"`
	StartAt       *string  `json:"start_at"`
	EndAt         *string  `json:"end_at"`
	Importance    *string  `json:"importance"`
	PapCount      *float64 `json:"pap_count"`
	CharacterID   *int64   `json:"character_id"`
	ESIFleetID    *int64   `json:"esi_fleet_id"`
	FleetConfigID *uint    `json:"fleet_config_id"`
	AutoSrpMode   *string  `json:"auto_srp_mode"`
}

type ManualAddFleetMembersRequest struct {
	CharacterNames []string `json:"character_names"`
}

type ManualAddFleetMembersResult struct {
	AddedCharacterNames   []string `json:"added_character_names"`
	MissingCharacterNames []string `json:"missing_character_names"`
}

// UpdateFleet 更新舰队信息
func (s *FleetService) UpdateFleet(fleetID string, userID uint, userRoles []string, req *UpdateFleetRequest) (*model.Fleet, error) {
	fleet, err := s.repo.GetByID(fleetID)
	if err != nil {
		return nil, errors.New("舰队不存在")
	}

	if !s.canManageFleet(fleet, userID, userRoles) {
		return nil, errors.New("权限不足")
	}

	if req.Title != nil {
		fleet.Title = *req.Title
	}
	if req.Description != nil {
		fleet.Description = *req.Description
	}
	if req.StartAt != nil {
		t, err := time.Parse(time.RFC3339, *req.StartAt)
		if err != nil {
			return nil, errors.New("起始时间格式错误")
		}
		fleet.StartAt = t
	}
	if req.EndAt != nil {
		t, err := time.Parse(time.RFC3339, *req.EndAt)
		if err != nil {
			return nil, errors.New("结束时间格式错误")
		}
		fleet.EndAt = t
	}
	if req.Importance != nil {
		fleet.Importance = *req.Importance
	}
	if req.PapCount != nil {
		fleet.PapCount = *req.PapCount
	}
	if req.CharacterID != nil {
		char, err := s.charRepo.GetByCharacterID(*req.CharacterID)
		if err != nil {
			return nil, errors.New("角色不存在")
		}
		if char.UserID != userID && !model.ContainsAnyRole(userRoles, model.RoleSuperAdmin, model.RoleAdmin) {
			return nil, errors.New("该角色不属于当前用户")
		}
		fleet.FCCharacterID = *req.CharacterID
		fleet.FCCharacterName = char.CharacterName
	}
	if req.ESIFleetID != nil {
		fleet.ESIFleetID = req.ESIFleetID
	}
	if req.FleetConfigID != nil {
		if *req.FleetConfigID == 0 {
			fleet.FleetConfigID = nil
		} else {
			fleet.FleetConfigID = req.FleetConfigID
		}
	}
	if req.AutoSrpMode != nil {
		fleet.AutoSrpMode = normalizeAutoSrpMode(*req.AutoSrpMode)
	}

	if err := s.repo.Update(fleet); err != nil {
		return nil, err
	}
	return fleet, nil
}

// DeleteFleet 删除舰队
func (s *FleetService) DeleteFleet(fleetID string, userID uint, userRoles []string) error {
	if _, err := s.repo.GetByID(fleetID); err != nil {
		return errors.New("舰队不存在")
	}
	if !s.canDeleteFleet(userRoles) {
		return errors.New("权限不足")
	}
	return s.repo.SoftDelete(fleetID)
}

// GetFleet 获取舰队详情
func (s *FleetService) GetFleet(fleetID string) (*model.Fleet, error) {
	return s.repo.GetByID(fleetID)
}

// RefreshESIFleetID 从 ESI 刷新舰队的 esi_fleet_id 并持久化
func (s *FleetService) RefreshESIFleetID(fleetID string, userID uint, userRoles []string) (*model.Fleet, error) {
	fleet, err := s.repo.GetByID(fleetID)
	if err != nil {
		return nil, errors.New("舰队不存在")
	}
	if !s.canManageFleet(fleet, userID, userRoles) {
		return nil, errors.New("权限不足")
	}

	ctx := context.Background()
	accessToken, err := s.ssoSvc.GetValidToken(ctx, fleet.FCCharacterID)
	if err != nil {
		return nil, fmt.Errorf("获取 Token 失败: %w", err)
	}

	path := fmt.Sprintf("/characters/%d/fleet/", fleet.FCCharacterID)
	var info CharacterFleetInfo
	if err := s.esiGet(ctx, path, accessToken, &info); err != nil {
		return nil, fmt.Errorf("ESI 查询失败: %w", err)
	}

	fleet.ESIFleetID = &info.FleetID
	if err := s.repo.Update(fleet); err != nil {
		return nil, err
	}
	return fleet, nil
}

// ListFleets 分页查询舰队列表
func (s *FleetService) ListFleets(page, pageSize int, filter repository.FleetFilter) ([]model.FleetListItem, int64, error) {
	page = normalizePage(page)
	pageSize = normalizePageSize(pageSize, 10, 100)
	return s.repo.List(page, pageSize, filter)
}

// GetMyFleets 返回当前用户参与过的舰队列表（按 fleet_member.user_id 过滤）
func (s *FleetService) GetMyFleets(userID uint) ([]model.Fleet, error) {
	return s.repo.ListFleetsByMemberUserID(userID, 200)
}

// ─────────────────────────────────────────────
//  舰队成员
// ─────────────────────────────────────────────

// GetMembers 获取舰队成员列表
func (s *FleetService) GetMembers(fleetID string) ([]model.FleetMember, error) {
	return s.repo.ListMembers(fleetID)
}

// ManualAddMembers 手动按角色名添加成员到舰队
func (s *FleetService) ManualAddMembers(fleetID string, userID uint, userRoles []string, req *ManualAddFleetMembersRequest) (*ManualAddFleetMembersResult, error) {
	fleet, err := s.repo.GetByID(fleetID)
	if err != nil {
		return nil, errors.New("舰队不存在")
	}
	if !s.canManageFleet(fleet, userID, userRoles) {
		return nil, errors.New("权限不足")
	}

	characterNames := normalizeCharacterNames(req.CharacterNames)
	if len(characterNames) == 0 {
		return nil, errors.New("请至少填写一个角色名")
	}

	result := &ManualAddFleetMembersResult{
		AddedCharacterNames:   make([]string, 0, len(characterNames)),
		MissingCharacterNames: make([]string, 0),
	}

	for _, name := range characterNames {
		char, err := s.charRepo.GetByCharacterName(name)
		if err != nil {
			result.MissingCharacterNames = append(result.MissingCharacterNames, name)
			continue
		}

		member := &model.FleetMember{
			FleetID:       fleet.ID,
			CharacterID:   char.CharacterID,
			CharacterName: char.CharacterName,
			UserID:        char.UserID,
		}
		if err := s.repo.AddMember(member); err != nil {
			return nil, err
		}

		result.AddedCharacterNames = append(result.AddedCharacterNames, char.CharacterName)

		if fleet.ESIFleetID != nil {
			go s.esiInviteMember(fleet, char)
		}
	}

	return result, nil
}

// JoinFleet 通过邀请码加入舰队
func (s *FleetService) JoinFleet(code string, userID uint, characterID int64) error {
	invite, err := s.repo.GetInviteByCode(code)
	if err != nil {
		return errors.New("邀请链接无效")
	}
	if !invite.Active {
		return errors.New("邀请链接已失效")
	}
	if time.Now().After(invite.ExpiresAt) {
		return errors.New("邀请链接已过期")
	}

	// 验证角色属于当前用户
	char, err := s.charRepo.GetByCharacterID(characterID)
	if err != nil {
		return errors.New("角色不存在")
	}
	if char.UserID != userID {
		return errors.New("该角色不属于当前用户")
	}

	fleet, err := s.repo.GetByID(invite.FleetID)
	if err != nil {
		return errors.New("舰队不存在")
	}

	// 记录成员
	member := &model.FleetMember{
		FleetID:       fleet.ID,
		CharacterID:   characterID,
		CharacterName: char.CharacterName,
		UserID:        userID,
	}
	if err := s.repo.AddMember(member); err != nil {
		return err
	}

	// 尝试通过 ESI 邀请角色加入游戏内舰队
	if fleet.ESIFleetID != nil {
		go s.esiInviteMember(fleet, char)
	}

	return nil
}

// esiInviteMember 通过 ESI 邀请角色加入游戏内舰队
func (s *FleetService) esiInviteMember(fleet *model.Fleet, char *model.EveCharacter) {
	ctx := context.Background()

	accessToken, err := s.ssoSvc.GetValidToken(ctx, fleet.FCCharacterID)
	if err != nil {
		global.Logger.Warn("[Fleet] 获取 FC Token 失败",
			zap.String("fleet_id", fleet.ID),
			zap.Error(err),
		)
		return
	}

	path := fmt.Sprintf("/fleets/%d/members/", *fleet.ESIFleetID)
	body := map[string]interface{}{
		"character_id": char.CharacterID,
		"role":         "squad_member",
	}

	if err := s.esiPost(ctx, path, accessToken, body); err != nil {
		global.Logger.Warn("[Fleet] ESI 邀请成员失败",
			zap.String("fleet_id", fleet.ID),
			zap.Int64("character_id", char.CharacterID),
			zap.Error(err),
		)
	}
}

// ─────────────────────────────────────────────
//  PAP 发放
// ─────────────────────────────────────────────

// IssuePap 发放 PAP 到舰队所有成员
func (s *FleetService) IssuePap(fleetID string, userID uint, userRoles []string) error {
	fleet, err := s.repo.GetByID(fleetID)
	if err != nil {
		return errors.New("舰队不存在")
	}
	if !s.canManageFleet(fleet, userID, userRoles) {
		return errors.New("权限不足")
	}
	if fleet.PapCount <= 0 {
		return errors.New("PAP 数量必须大于 0")
	}

	// 1. 先尝试 ESI 同步成员（失败不阻断发放）
	if fleet.ESIFleetID != nil {
		if _, syncErr := s.SyncESIMembers(fleetID, userID, userRoles); syncErr != nil {
			global.Logger.Warn("[Fleet] IssuePap ESI 同步失败，继续发放",
				zap.String("fleet_id", fleetID),
				zap.Error(syncErr),
			)
		}
	}

	// 2. 获取最新成员列表
	members, err := s.repo.ListMembers(fleetID)
	if err != nil {
		return err
	}
	if len(members) == 0 {
		return errors.New("舰队中没有成员")
	}

	// 3. 获取旧 PAP 记录，用于钱包差量计算（在事务外读取，快照一致即可）
	oldLogs, err := s.repo.ListPapLogsByFleet(fleetID)
	if err != nil {
		return err
	}
	rateMap := s.rateRepo.GetRateMap()
	walletRate := papImportanceToWalletRate(fleet.Importance, rateMap)
	fcSalary := s.getPAPFCSalary()
	fcSalaryMonthlyLimit := s.getPAPFCSalaryMonthlyLimit()
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	nextMonthStart := monthStart.AddDate(0, 1, 0)

	fcInMembers := false
	for _, m := range members {
		if m.UserID == fleet.FCUserID {
			fcInMembers = true
			break
		}
	}

	oldWalletPerUser := buildPapWalletByUser(toPapWalletEntriesFromLogs(oldLogs), walletRate)
	oldFCSalaryAmount := 0.0
	if tx, lookupErr := s.walletRepo.GetTransactionByUserRefTypeRefIDInRange(
		fleet.FCUserID,
		model.WalletRefPapFCSalary,
		fleetID,
		monthStart,
		nextMonthStart,
	); lookupErr == nil {
		oldFCSalaryAmount = tx.Amount
	} else if !errors.Is(lookupErr, gorm.ErrRecordNotFound) {
		global.Logger.Warn("[Fleet] 查询 FC 工资历史流水失败",
			zap.String("fleet_id", fleetID),
			zap.Uint("user_id", fleet.FCUserID),
			zap.Error(lookupErr),
		)
	}

	// 4. 构建新 PAP 记录
	newLogs := make([]model.FleetPapLog, 0, len(members))
	newEntries := make([]papWalletEntry, 0, len(members))
	for _, m := range members {
		newLogs = append(newLogs, model.FleetPapLog{
			FleetID:     fleetID,
			CharacterID: m.CharacterID,
			UserID:      m.UserID,
			PapCount:    fleet.PapCount,
			IssuedBy:    userID,
		})
		newEntries = append(newEntries, papWalletEntry{
			UserID:   m.UserID,
			PapCount: fleet.PapCount,
		})
	}

	// 5. 事务：更新 PAP 记录 + 钱包差量
	tx := global.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Where("fleet_id = ?", fleetID).Delete(&model.FleetPapLog{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Create(&newLogs).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 合并涉及的所有 user_id，计算并应用差量
	allUsers := make(map[uint]struct{})
	for uid := range oldWalletPerUser {
		allUsers[uid] = struct{}{}
	}
	newWalletPerUser := buildPapWalletByUser(newEntries, walletRate)
	newFCSalaryAmount := 0.0
	if fcInMembers || oldFCSalaryAmount > 0 {
		// monthlyCount is only needed when this is a new entry (no prior salary for this fleet).
		// When oldFCSalaryAmount > 0, calculateFCSalaryAmount ignores monthlyCount entirely,
		// so skip the DB query in that case.
		var monthlyCount int64
		if fcInMembers && oldFCSalaryAmount == 0 {
			var countErr error
			monthlyCount, countErr = s.walletRepo.CountTransactionsByUserRefTypeInRange(
				fleet.FCUserID,
				model.WalletRefPapFCSalary,
				monthStart,
				nextMonthStart,
			)
			if countErr != nil {
				global.Logger.Warn("[Fleet] 统计 FC 工资次数失败",
					zap.String("fleet_id", fleetID),
					zap.Uint("user_id", fleet.FCUserID),
					zap.Error(countErr),
				)
				monthlyCount = int64(fcSalaryMonthlyLimit)
			}
		}
		newFCSalaryAmount = calculateFCSalaryAmount(fcInMembers, oldFCSalaryAmount, monthlyCount, fcSalaryMonthlyLimit, fcSalary)
	}
	for uid := range newWalletPerUser {
		allUsers[uid] = struct{}{}
	}

	reason := fmt.Sprintf("舰队 PAP 奖励: %s", fleetID)
	for uid := range allUsers {
		delta := newWalletPerUser[uid] - oldWalletPerUser[uid]
		if delta == 0 {
			continue
		}
		if err := s.walletSvc.ApplyWalletDeltaTx(tx, uid, delta, reason, model.WalletRefPapReward, fleetID); err != nil {
			global.Logger.Warn("[Fleet] 钱包差量更新失败",
				zap.Uint("user_id", uid),
				zap.Float64("delta", delta),
				zap.Error(err),
			)
			// 钱包失败不阻断整个操作，继续
		}
	}

	fcSalaryDelta := newFCSalaryAmount - oldFCSalaryAmount
	if fcSalaryDelta != 0 {
		if err := s.walletSvc.ApplyWalletDeltaTx(tx, fleet.FCUserID, fcSalaryDelta, fmt.Sprintf("舰队 FC 工资: %s", fleetID), model.WalletRefPapFCSalary, fleetID); err != nil {
			global.Logger.Warn("[Fleet] FC 工资差量更新失败",
				zap.String("fleet_id", fleetID),
				zap.Uint("user_id", fleet.FCUserID),
				zap.Float64("delta", fcSalaryDelta),
				zap.Error(err),
			)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	// 尝试更新 ESI 舰队 MOTD
	if fleet.ESIFleetID != nil {
		go s.updateFleetMotd(fleet)
	}

	// 异步触发新成员的 KM 刷新（只刷新本次 PAP 新增的成员）
	go s.triggerNewMembersKMRefresh(fleetID)

	return nil
}

type papWalletEntry struct {
	UserID   uint
	PapCount float64
}

func toPapWalletEntriesFromLogs(logs []model.FleetPapLog) []papWalletEntry {
	entries := make([]papWalletEntry, 0, len(logs))
	for _, log := range logs {
		entries = append(entries, papWalletEntry{
			UserID:   log.UserID,
			PapCount: log.PapCount,
		})
	}
	return entries
}

func buildPapWalletByUser(entries []papWalletEntry, walletRate float64) map[uint]float64 {
	result := make(map[uint]float64, len(entries))
	for _, entry := range entries {
		result[entry.UserID] += entry.PapCount * walletRate
	}
	return result
}

// calculateFCSalaryAmount returns the target FC salary amount for a fleet issuance.
// If the FC is not in the member list, returns 0 (revoke any prior salary for this fleet).
// If a salary was already issued for this fleet (existingSalaryAmount > 0), the monthly limit
// is intentionally bypassed — this is a re-issue that updates the amount to the current rate,
// not a new charge against the cap.
// monthlyCount is only consulted for genuinely new entries (existingSalaryAmount == 0).
func calculateFCSalaryAmount(fcInMembers bool, existingSalaryAmount float64, monthlyCount int64, monthlyLimit int, currentSalary float64) float64 {
	if !fcInMembers {
		return 0
	}
	if existingSalaryAmount > 0 {
		return currentSalary
	}
	if monthlyLimit <= 0 || monthlyCount >= int64(monthlyLimit) {
		return 0
	}
	return currentSalary
}

// papImportanceToWalletRate 将舰队重要性映射到对应的 PAP 兑换汇率（系统钱包 / 1 PAP）
func papImportanceToWalletRate(importance string, rateMap map[string]float64) float64 {
	var papType string
	switch importance {
	case model.FleetImportanceCTA:
		papType = model.PAPTypeCTA
	case model.FleetImportanceStratOp:
		papType = model.PAPTypeStratOp
	default:
		papType = model.PAPTypeSkirmish
	}
	if rate, ok := rateMap[papType]; ok {
		return rate
	}
	return 1
}

func (s *FleetService) getPAPFCSalary() float64 {
	return s.configRepo.GetFloat(model.SysConfigPAPFCSalary, model.SysConfigDefaultPAPFCSalary)
}

func (s *FleetService) getPAPFCSalaryMonthlyLimit() int {
	return s.configRepo.GetInt(model.SysConfigPAPFCSalaryLimit, model.SysConfigDefaultPAPFCSalaryLimit)
}

// triggerNewMembersKMRefresh 对本舰队中需要 KM 刷新的成员执行触发：
// 从未触发过，或上次触发时间超过 15 分钟（冷却期外）的成员才会被触发
func (s *FleetService) triggerNewMembersKMRefresh(fleetID string) {
	if FleetKMRefreshFunc == nil {
		return
	}
	newMembers, err := s.repo.ListMembersForKMRefresh(fleetID)
	if err != nil {
		global.Logger.Warn("[Fleet] 查询待 KM 刷新成员失败",
			zap.String("fleet_id", fleetID),
			zap.Error(err),
		)
		return
	}
	if len(newMembers) == 0 {
		return
	}

	charIDs := make([]int64, 0, len(newMembers))
	for _, m := range newMembers {
		charIDs = append(charIDs, m.CharacterID)
	}

	// 先标记，再触发——避免并发重复触发
	if err := s.repo.MarkMembersKMRefreshed(fleetID, charIDs); err != nil {
		global.Logger.Warn("[Fleet] 标记 KM 刷新状态失败",
			zap.String("fleet_id", fleetID),
			zap.Error(err),
		)
		// 即使标记失败也继续触发，下次 PAP 仍会重试
	}

	for _, charID := range charIDs {
		FleetKMRefreshFunc(charID)
	}
	global.Logger.Info("[Fleet] 已触发舰队成员 KM 刷新",
		zap.String("fleet_id", fleetID),
		zap.Int("count", len(charIDs)),
	)

	// KM 刷新完成后触发自动 SRP
	if FleetAutoSRPFunc != nil {
		FleetAutoSRPFunc(fleetID)
	}
}

// ListMembersWithPap 分页查询舰队成员（含 PAP 信息）
func (s *FleetService) ListMembersWithPap(fleetID string, page, pageSize int) ([]repository.MemberWithPap, int64, error) {
	page = normalizePage(page)
	pageSize = normalizePageSize(pageSize, 260, 260)
	return s.repo.ListMembersWithPap(fleetID, page, pageSize)
}

// updateFleetMotd 在 ESI 舰队 MOTD 中追加 PAP 发放记录
func (s *FleetService) updateFleetMotd(fleet *model.Fleet) {
	ctx := context.Background()

	accessToken, err := s.ssoSvc.GetValidToken(ctx, fleet.FCCharacterID)
	if err != nil {
		global.Logger.Warn("[Fleet] 获取 FC Token 失败（更新 MOTD）",
			zap.String("fleet_id", fleet.ID),
			zap.Error(err),
		)
		return
	}

	// 先获取当前 MOTD
	fleetPath := fmt.Sprintf("/fleets/%d/", *fleet.ESIFleetID)
	var fleetInfo struct {
		Motd string `json:"motd"`
	}
	if err := s.esiGet(ctx, fleetPath, accessToken, &fleetInfo); err != nil {
		global.Logger.Warn("[Fleet] 获取舰队信息失败",
			zap.String("fleet_id", fleet.ID),
			zap.Error(err),
		)
		return
	}

	// 追加 PAP 记录
	papNote := fmt.Sprintf("\n- %.1f PAP 已发放 %s -", fleet.PapCount, time.Now().Format("2006-01-02 15:04"))
	newMotd := fleetInfo.Motd + papNote

	body := map[string]interface{}{
		"motd": newMotd,
	}
	if err := s.esiPut(ctx, fleetPath, accessToken, body); err != nil {
		global.Logger.Warn("[Fleet] 更新舰队 MOTD 失败",
			zap.String("fleet_id", fleet.ID),
			zap.Error(err),
		)
	}
}

// GetPapLogs 获取舰队 PAP 发放记录
func (s *FleetService) GetPapLogs(fleetID string) ([]model.FleetPapLog, error) {
	return s.repo.ListPapLogsByFleet(fleetID)
}

// GetUserPapLogs 获取用户的 PAP 记录（含角色名、FC 名称、舰队信息）
func (s *FleetService) GetUserPapLogs(userID uint) ([]repository.PapLogDetail, error) {
	return s.repo.ListPapLogsDetailByUser(userID)
}

// GetCorporationPapSummary 获取军团维度 PAP 汇总
func (s *FleetService) GetCorporationPapSummary(page, pageSize int, period string, year int, corpTickers []string) (*CorporationPapSummaryResponse, error) {
	page = normalizePage(page)
	pageSize = normalizeLedgerPageSize(pageSize)

	now := time.Now()
	filter, normalizedPeriod, normalizedYear, err := s.buildCorporationPapFilter(period, year, now)
	if err != nil {
		return nil, err
	}

	rows, err := s.repo.ListCorporationPapSummaryAll(filter)
	if err != nil {
		return nil, err
	}
	allRows, err := s.repo.ListCorporationPapSummaryAll(repository.FleetPapSummaryFilter{})
	if err != nil {
		return nil, err
	}

	userIDs := make([]uint, 0, len(rows)+len(allRows))
	seenUsers := make(map[uint]struct{}, len(rows)+len(allRows))
	for _, row := range rows {
		if _, ok := seenUsers[row.UserID]; ok {
			continue
		}
		seenUsers[row.UserID] = struct{}{}
		userIDs = append(userIDs, row.UserID)
	}
	for _, row := range allRows {
		if _, ok := seenUsers[row.UserID]; ok {
			continue
		}
		seenUsers[row.UserID] = struct{}{}
		userIDs = append(userIDs, row.UserID)
	}

	profileByUserID, err := s.resolveCorporationPapProfiles(userIDs)
	if err != nil {
		return nil, err
	}

	allowedTickerSet := make(map[string]struct{})
	for _, ticker := range corpTickers {
		normalized := strings.ToUpper(strings.TrimSpace(ticker))
		if normalized != "" {
			allowedTickerSet[normalized] = struct{}{}
		}
	}

	matchesCorpFilter := func(userID uint) bool {
		if len(allowedTickerSet) == 0 {
			return true
		}
		profile := profileByUserID[userID]
		_, ok := allowedTickerSet[strings.ToUpper(profile.CorpTicker)]
		return ok
	}

	items := make([]CorporationPapSummaryItem, 0, len(rows))
	var filteredPapTotal float64
	var filteredStratOpTotal float64
	for _, row := range rows {
		if !matchesCorpFilter(row.UserID) {
			continue
		}
		profile := profileByUserID[row.UserID]
		items = append(items, CorporationPapSummaryItem{
			UserID:            row.UserID,
			Nickname:          profile.Nickname,
			CorpTicker:        profile.CorpTicker,
			MainCharacterName: profile.MainCharacterName,
			CharacterCount:    profile.CharacterCount,
			StratOpPaps:       row.StratOpPaps,
			SkirmishPaps:      row.SkirmishPaps,
		})
		filteredPapTotal += row.StratOpPaps + row.SkirmishPaps
		filteredStratOpTotal += row.StratOpPaps
	}

	var allPapTotal float64
	for _, row := range allRows {
		if !matchesCorpFilter(row.UserID) {
			continue
		}
		allPapTotal += row.StratOpPaps + row.SkirmishPaps
	}

	total := int64(len(items))
	start := (page - 1) * pageSize
	if start > len(items) {
		start = len(items)
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	pagedItems := items[start:end]

	mainChars := make([]string, 0, len(pagedItems))
	for _, item := range pagedItems {
		if item.MainCharacterName != "" {
			mainChars = append(mainChars, item.MainCharacterName)
		}
	}

	allianceStratPapByMainChar, err := s.papRepo.SumStrategicPapByMainCharacters(mainChars, filter.StartAt, filter.EndAt)
	if err != nil {
		return nil, err
	}
	for i := range pagedItems {
		pagedItems[i].AllianceStratPaps = allianceStratPapByMainChar[pagedItems[i].MainCharacterName]
	}

	overview := CorporationPapOverview{
		FilteredPapTotal:     filteredPapTotal,
		FilteredStratOpTotal: filteredStratOpTotal,
		AllPapTotal:          allPapTotal,
		FilteredUserCount:    total,
		Period:               normalizedPeriod,
	}
	if normalizedYear != nil {
		overview.Year = normalizedYear
	}

	return &CorporationPapSummaryResponse{
		List:     pagedItems,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Overview: overview,
	}, nil
}

// ─────────────────────────────────────────────
//  从 ESI 拉取舰队成员并记录
// ─────────────────────────────────────────────

// ESIFleetMember ESI 舰队成员响应
type ESIFleetMember struct {
	CharacterID   int64  `json:"character_id"`
	JoinTime      string `json:"join_time"`
	Role          string `json:"role"`
	RoleName      string `json:"role_name"`
	ShipTypeID    int64  `json:"ship_type_id"`
	SolarSystemID int64  `json:"solar_system_id"`
	SquadID       int64  `json:"squad_id"`
	WingID        int64  `json:"wing_id"`
}

// SyncESIMembers 从 ESI 获取当前舰队成员并记录到数据库
func (s *FleetService) SyncESIMembers(fleetID string, userID uint, userRoles []string) ([]ESIFleetMember, error) {
	fleet, err := s.repo.GetByID(fleetID)
	if err != nil {
		return nil, errors.New("舰队不存在")
	}
	if !s.canManageFleet(fleet, userID, userRoles) {
		return nil, errors.New("权限不足")
	}
	if fleet.ESIFleetID == nil {
		return nil, errors.New("未设置 ESI 舰队 ID")
	}

	ctx := context.Background()
	accessToken, err := s.ssoSvc.GetValidToken(ctx, fleet.FCCharacterID)
	if err != nil {
		return nil, fmt.Errorf("获取 FC Token 失败: %w", err)
	}

	path := fmt.Sprintf("/fleets/%d/members/", *fleet.ESIFleetID)
	var esiMembers []ESIFleetMember
	if err := s.esiGet(ctx, path, accessToken, &esiMembers); err != nil {
		return nil, fmt.Errorf("获取 ESI 舰队成员失败: %w", err)
	}

	// 将 ESI 成员记录到数据库
	for _, em := range esiMembers {
		char, err := s.charRepo.GetByCharacterID(em.CharacterID)
		if err != nil {
			// 角色不在系统中，跳过
			continue
		}
		shipTypeID := em.ShipTypeID
		solarSystemID := em.SolarSystemID
		member := &model.FleetMember{
			FleetID:       fleetID,
			CharacterID:   em.CharacterID,
			CharacterName: char.CharacterName,
			UserID:        char.UserID,
			ShipTypeID:    &shipTypeID,
			SolarSystemID: &solarSystemID,
		}
		_ = s.repo.AddMember(member)
	}

	return esiMembers, nil
}

// ─────────────────────────────────────────────
//  邀请链接
// ─────────────────────────────────────────────

// CreateInvite 创建舰队邀请链接
func (s *FleetService) CreateInvite(fleetID string, userID uint, userRoles []string) (*model.FleetInvite, error) {
	fleet, err := s.repo.GetByID(fleetID)
	if err != nil {
		return nil, errors.New("舰队不存在")
	}
	if !s.canManageFleet(fleet, userID, userRoles) {
		return nil, errors.New("权限不足")
	}

	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	code := hex.EncodeToString(b)

	// 邮证链接过期时间：取舰队结束时间，但至少保证 24 小时内有效
	expiresAt := fleet.EndAt
	if expiresAt.Before(time.Now().Add(24 * time.Hour)) {
		expiresAt = time.Now().Add(24 * time.Hour)
	}

	invite := &model.FleetInvite{
		FleetID:   fleetID,
		Code:      code,
		Active:    true,
		ExpiresAt: expiresAt,
	}

	if err := s.repo.CreateInvite(invite); err != nil {
		return nil, err
	}
	return invite, nil
}

// GetInvites 获取舰队邀请链接列表
func (s *FleetService) GetInvites(fleetID string) ([]model.FleetInvite, error) {
	return s.repo.ListInvitesByFleet(fleetID)
}

// DeactivateInvite 禁用邀请链接
func (s *FleetService) DeactivateInvite(inviteID uint, userID uint, userRoles []string) error {
	if !s.canManageFleet(nil, userID, userRoles) {
		return errors.New("权限不足")
	}
	return s.repo.DeactivateInvite(inviteID)
}

// ─────────────────────────────────────────────
//  权限判断
// ─────────────────────────────────────────────

// canManageFleet 判断用户是否有权管理舰队相关功能。
// super_admin、admin、senior_fc 可管理任何舰队；
// fc 角色仅能管理自己创建的舰队（fleet.FCUserID == userID）。
// fleet 为 nil 时退化为纯角色判断（用于不依赖具体舰队的操作）。
func (s *FleetService) canManageFleet(fleet *model.Fleet, userID uint, userRoles []string) bool {
	if model.ContainsAnyRole(userRoles, model.RoleSuperAdmin, model.RoleAdmin, model.RoleSeniorFC) {
		return true
	}
	if model.ContainsAnyRole(userRoles, model.RoleFC) {
		if fleet == nil {
			return true
		}
		return fleet.FCUserID == userID
	}
	return false
}

func (s *FleetService) canDeleteFleet(userRoles []string) bool {
	return model.ContainsAnyRole(userRoles, model.RoleSuperAdmin, model.RoleAdmin)
}

func normalizeCharacterNames(names []string) []string {
	result := make([]string, 0, len(names))
	seen := make(map[string]struct{}, len(names))
	for _, name := range names {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

func (s *FleetService) buildCorporationPapFilter(period string, year int, now time.Time) (repository.FleetPapSummaryFilter, string, *int, error) {
	location := now.Location()

	switch period {
	case CorporationPapPeriodCurrentMonth:
		startAt := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, location)
		return repository.FleetPapSummaryFilter{StartAt: &startAt}, CorporationPapPeriodCurrentMonth, nil, nil
	case "", CorporationPapPeriodLastMonth:
		startAt := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, location)
		endAt := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, location)
		return repository.FleetPapSummaryFilter{StartAt: &startAt, EndAt: &endAt}, CorporationPapPeriodLastMonth, nil, nil
	case CorporationPapPeriodAtYear, "last_year":
		if period == "last_year" {
			year = now.Year() - 1
		}
		if year <= 0 {
			year = now.Year()
		}
		startAt := time.Date(year, time.January, 1, 0, 0, 0, 0, location)
		endAt := time.Date(year+1, time.January, 1, 0, 0, 0, 0, location)
		return repository.FleetPapSummaryFilter{StartAt: &startAt, EndAt: &endAt}, CorporationPapPeriodAtYear, &year, nil
	case CorporationPapPeriodAll:
		return repository.FleetPapSummaryFilter{}, CorporationPapPeriodAll, nil, nil
	default:
		return repository.FleetPapSummaryFilter{}, "", nil, errors.New("无效的 PAP 时间筛选条件")
	}
}

type corporationPapProfile struct {
	Nickname          string
	MainCharacterName string
	CharacterCount    int
	MainCharacterID   int64
	CorporationID     int64
	CorpTicker        string
}

func (s *FleetService) resolveCorporationPapProfiles(userIDs []uint) (map[uint]corporationPapProfile, error) {
	profileByUserID := make(map[uint]corporationPapProfile, len(userIDs))
	if len(userIDs) == 0 {
		return profileByUserID, nil
	}

	users, err := s.userRepo.ListByIDs(userIDs)
	if err != nil {
		return nil, err
	}

	primaryCharacterIDs := make([]int64, 0, len(users))
	for _, user := range users {
		if user.PrimaryCharacterID > 0 {
			primaryCharacterIDs = append(primaryCharacterIDs, user.PrimaryCharacterID)
		}
	}

	primaryChars, err := s.charRepo.ListByCharacterIDs(primaryCharacterIDs)
	if err != nil {
		return nil, err
	}
	primaryNameByCharacterID := make(map[int64]string, len(primaryChars))
	primaryCharByCharacterID := make(map[int64]model.EveCharacter, len(primaryChars))
	for _, char := range primaryChars {
		primaryNameByCharacterID[char.CharacterID] = char.CharacterName
		primaryCharByCharacterID[char.CharacterID] = char
	}

	fallbackChars, err := s.charRepo.ListByUserIDs(userIDs)
	if err != nil {
		return nil, err
	}
	fallbackNameByUserID := make(map[uint]string, len(fallbackChars))
	characterCountByUserID := make(map[uint]int, len(fallbackChars))
	fallbackCorpIDByUserID := make(map[uint]int64, len(fallbackChars))
	for _, char := range fallbackChars {
		characterCountByUserID[char.UserID]++
		if fallbackNameByUserID[char.UserID] == "" {
			fallbackNameByUserID[char.UserID] = char.CharacterName
		}
		if fallbackCorpIDByUserID[char.UserID] == 0 {
			fallbackCorpIDByUserID[char.UserID] = char.CorporationID
		}
	}

	primaryCorpIDs := make([]int64, 0, len(users))
	for _, user := range users {
		profile := corporationPapProfile{
			Nickname:        user.Nickname,
			CharacterCount:  characterCountByUserID[user.ID],
			MainCharacterID: user.PrimaryCharacterID,
		}
		if primaryChar, ok := primaryCharByCharacterID[user.PrimaryCharacterID]; ok {
			profile.CorporationID = primaryChar.CorporationID
			if primaryChar.CorporationID > 0 {
				primaryCorpIDs = append(primaryCorpIDs, primaryChar.CorporationID)
			}
		} else if fallbackCorpIDByUserID[user.ID] > 0 {
			profile.CorporationID = fallbackCorpIDByUserID[user.ID]
			primaryCorpIDs = append(primaryCorpIDs, fallbackCorpIDByUserID[user.ID])
		}

		if name := primaryNameByCharacterID[user.PrimaryCharacterID]; name != "" {
			profile.MainCharacterName = name
		} else if name := fallbackNameByUserID[user.ID]; name != "" {
			profile.MainCharacterName = name
		} else if user.Nickname != "" {
			profile.MainCharacterName = user.Nickname
		} else {
			profile.MainCharacterName = fmt.Sprintf("User#%d", user.ID)
		}
		profileByUserID[user.ID] = profile
	}

	tickerByCorpID, err := s.resolveCorporationTickers(primaryCorpIDs)
	if err != nil {
		return nil, err
	}

	for _, userID := range userIDs {
		if _, ok := profileByUserID[userID]; !ok {
			profileByUserID[userID] = corporationPapProfile{
				MainCharacterName: fmt.Sprintf("User#%d", userID),
				CharacterCount:    characterCountByUserID[userID],
			}
		}
		profile := profileByUserID[userID]
		if profile.CorporationID > 0 {
			profile.CorpTicker = tickerByCorpID[profile.CorporationID]
		}
		profileByUserID[userID] = profile
	}

	return profileByUserID, nil
}

func (s *FleetService) resolveCorporationTickers(corporationIDs []int64) (map[int64]string, error) {
	tickerByCorpID := make(map[int64]string)
	seen := make(map[int64]struct{})

	for _, corpID := range corporationIDs {
		if corpID <= 0 {
			continue
		}
		if _, ok := seen[corpID]; ok {
			continue
		}
		seen[corpID] = struct{}{}

		var corpInfo struct {
			Ticker string `json:"ticker"`
		}
		if err := s.esiGetPublic(context.Background(), fmt.Sprintf("/corporations/%d/", corpID), &corpInfo); err != nil {
			global.Logger.Warn("解析军团 ticker 失败", zap.Int64("corporation_id", corpID), zap.Error(err))
			continue
		}
		tickerByCorpID[corpID] = corpInfo.Ticker
	}

	return tickerByCorpID, nil
}

// ─────────────────────────────────────────────
//  ESI: 获取角色当前舰队信息
// ─────────────────────────────────────────────

// CharacterFleetInfo 角色当前舰队信息
type CharacterFleetInfo struct {
	FleetID     int64  `json:"fleet_id"`
	FleetBossID int64  `json:"fleet_boss_id"`
	Role        string `json:"role"`
	SquadID     int64  `json:"squad_id"`
	WingID      int64  `json:"wing_id"`
}

// normalizeAutoSrpMode 规范化自动 SRP 模式值
func normalizeAutoSrpMode(mode string) string {
	switch mode {
	case model.FleetAutoSrpSubmitOnly, model.FleetAutoSrpAutoApprove:
		return mode
	default:
		return model.FleetAutoSrpDisabled
	}
}

// GetCharacterFleetInfo 获取角色当前所在的 ESI 舰队信息
func (s *FleetService) GetCharacterFleetInfo(userID uint, characterID int64) (*CharacterFleetInfo, error) {
	char, err := s.charRepo.GetByCharacterID(characterID)
	if err != nil {
		return nil, errors.New("角色不存在")
	}
	if char.UserID != userID {
		return nil, errors.New("该角色不属于当前用户")
	}

	ctx := context.Background()
	accessToken, err := s.ssoSvc.GetValidToken(ctx, characterID)
	if err != nil {
		return nil, fmt.Errorf("获取 Token 失败: %w", err)
	}

	path := fmt.Sprintf("/characters/%d/fleet/", characterID)
	var info CharacterFleetInfo
	if err := s.esiGet(ctx, path, accessToken, &info); err != nil {
		return nil, fmt.Errorf("获取舰队信息失败: %w", err)
	}

	return &info, nil
}

// ─────────────────────────────────────────────
//  ESI HTTP 辅助方法（避免循环依赖 esi 包）
// ─────────────────────────────────────────────

// esiGet GET 请求并解析 JSON 响应
func (s *FleetService) esiGet(ctx context.Context, path, accessToken string, out interface{}) error {
	return s.esiClient.Get(ctx, path, accessToken, out)
}

// esiPost POST 请求（不期望响应体）
func (s *FleetService) esiPost(ctx context.Context, path, accessToken string, body interface{}) error {
	return s.esiClient.PostNoContent(ctx, path, accessToken, body)
}

// esiPut PUT 请求（不期望响应体）
func (s *FleetService) esiPut(ctx context.Context, path, accessToken string, body interface{}) error {
	return s.esiClient.PutJSON(ctx, path, accessToken, body)
}

// esiGetPublic GET 公共 ESI 接口并解析 JSON 响应
func (s *FleetService) esiGetPublic(ctx context.Context, path string, out interface{}) error {
	return s.esiClient.Get(ctx, path, "", out)
}
