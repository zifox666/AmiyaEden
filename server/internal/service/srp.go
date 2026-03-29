package service

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"
)

// slotCategory 将 HiSlot0, HiSlot1 等 flagName 归类为 "HiSlot"
func slotCategory(flagName string) string {
	return strings.TrimRight(flagName, "0123456789")
}

// SrpService 补损业务逻辑层
type SrpService struct {
	repo      *repository.SrpRepository
	fleetRepo *repository.FleetRepository
	charRepo  *repository.EveCharacterRepository
	userRepo  *repository.UserRepository
	sdeRepo   *repository.SdeRepository
	kmRepo    *repository.KillmailRepository
	ssoSvc    *EveSSOService
	walletSvc *SysWalletService
}

func NewSrpService() *SrpService {
	return &SrpService{
		repo:      repository.NewSrpRepository(),
		fleetRepo: repository.NewFleetRepository(),
		charRepo:  repository.NewEveCharacterRepository(),
		userRepo:  repository.NewUserRepository(),
		sdeRepo:   repository.NewSdeRepository(),
		kmRepo:    repository.NewKillmailRepository(),
		ssoSvc:    NewEveSSOService(),
		walletSvc: NewSysWalletService(),
	}
}

const (
	SrpPayoutModeManualTransfer = "manual_transfer"
	SrpPayoutModeFuxiCoin       = "fuxi_coin"
)

// ─────────────────────────────────────────────
//  KM 解析辅助
// ─────────────────────────────────────────────

// resolveCharacterKillmail 确认 killmailID 与 characterID 有关联，并返回 EveKillmailList
func (s *SrpService) resolveCharacterKillmail(killmailID int64, characterID int64) (*model.EveKillmailList, error) {
	// 验证人物-KM 关联关系
	if _, err := s.kmRepo.GetCharacterKillmailLink(characterID, killmailID); err != nil {
		return nil, errors.New("该 KM 不属于指定人物，或尚未被 ESI 刷新任务录入")
	}
	// 加载 KM 详情
	km, err := s.kmRepo.GetKillmailByID(killmailID)
	if err != nil {
		return nil, errors.New("KM 详情不存在")
	}
	return km, nil
}

// ─────────────────────────────────────────────
//  舰船价格表
// ─────────────────────────────────────────────

// ListShipPrices 返回所有（可按关键字过滤）舰船价格
func (s *SrpService) ListShipPrices(keyword string) ([]model.SrpShipPrice, error) {
	return s.repo.ListShipPrices(keyword)
}

// UpsertShipPriceRequest 创建/更新舰船价格请求
type UpsertShipPriceRequest struct {
	ID         uint    `json:"id"` // 0=新建，非0=更新
	ShipTypeID int64   `json:"ship_type_id" binding:"required"`
	ShipName   string  `json:"ship_name"    binding:"required"`
	Amount     float64 `json:"amount"       binding:"required,min=0"`
}

// UpsertShipPrice 创建或更新舰船价格
func (s *SrpService) UpsertShipPrice(userID uint, req *UpsertShipPriceRequest) (*model.SrpShipPrice, error) {
	p := &model.SrpShipPrice{
		ID:         req.ID,
		ShipTypeID: req.ShipTypeID,
		ShipName:   req.ShipName,
		Amount:     req.Amount,
		UpdatedBy:  userID,
	}
	if req.ID == 0 {
		p.CreatedBy = userID
	} else {
		// 保留原始 created_by
		existing, err := s.repo.GetShipPriceByTypeID(req.ShipTypeID)
		if err == nil {
			p.CreatedBy = existing.CreatedBy
		}
	}
	if err := s.repo.UpsertShipPrice(p); err != nil {
		return nil, err
	}
	return p, nil
}

// DeleteShipPrice 删除舰船价格
func (s *SrpService) DeleteShipPrice(id uint) error {
	return s.repo.DeleteShipPrice(id)
}

// ─────────────────────────────────────────────
//  申请提交
// ─────────────────────────────────────────────

// SubmitApplicationRequest 提交补损申请请求
type SubmitApplicationRequest struct {
	CharacterID int64   `json:"character_id"  binding:"required"` // 受损人物 ID
	KillmailID  int64   `json:"killmail_id"   binding:"required"` // zkillboard killmail id
	FleetID     *string `json:"fleet_id"`                         // 关联舰队（可选）
	Note        string  `json:"note"`                             // 备注（无舰队时必填）
}

// SubmitApplication 提交补损申请
func (s *SrpService) SubmitApplication(userID uint, req *SubmitApplicationRequest) (*model.SrpApplication, error) {
	// 1. 验证人物属于当前用户
	char, err := s.charRepo.GetByCharacterID(req.CharacterID)
	if err != nil || char.UserID != userID {
		return nil, errors.New("人物不属于当前用户或不存在")
	}

	// 2. 无舰队时需要填写备注
	if req.FleetID == nil && req.Note == "" {
		return nil, errors.New("未关联舰队时，备注不能为空")
	}

	// 3. 检查是否重复提交
	if s.repo.ExistsApplicationByKillmail(req.KillmailID, req.CharacterID) {
		return nil, errors.New("该 KM 已提交过补损申请，不能重复提交")
	}

	// 4. 获取 KM 详情（验证人物与 KM 关联）
	km, err := s.resolveCharacterKillmail(req.KillmailID, req.CharacterID)
	if err != nil {
		return nil, err
	}

	// 5. 确认该 KM 的受害者确实是这个人物
	if km.CharacterID != req.CharacterID {
		return nil, errors.New("该 KM 的受害者不是指定人物，无法申请补损")
	}

	// 6. 关联舰队时验证
	if req.FleetID != nil && *req.FleetID != "" {
		fleet, ferr := s.fleetRepo.GetByID(*req.FleetID)
		if ferr != nil {
			return nil, errors.New("关联的舰队不存在")
		}
		// KM 时间必须在舰队时间范围内
		if km.KillmailTime.Before(fleet.StartAt) || km.KillmailTime.After(fleet.EndAt) {
			return nil, errors.New("KM 时间不在舰队活动时间范围内")
		}
		// 人物必须是舰队成员
		members, _ := s.fleetRepo.ListMembers(*req.FleetID)
		isMember := false
		for _, m := range members {
			if m.CharacterID == req.CharacterID {
				isMember = true
				break
			}
		}
		if !isMember {
			return nil, errors.New("该人物不是该舰队的成员，无法申请补损")
		}
	}

	// 7. 计算 SRP 推荐金额（共用推荐金额计算逻辑）
	autoSrpSvc := NewAutoSrpService()
	recommended, _ := autoSrpSvc.RecommendSrpAmount(km.ShipTypeID, req.KillmailID, req.FleetID)

	// 8. 构建申请
	app := &model.SrpApplication{
		UserID:            userID,
		CharacterID:       req.CharacterID,
		CharacterName:     char.CharacterName,
		KillmailID:        req.KillmailID,
		FleetID:           req.FleetID,
		Note:              req.Note,
		ShipTypeID:        km.ShipTypeID,
		ShipName:          "", // 由前端或 SDE 填写；此处留空
		SolarSystemID:     km.SolarSystemID,
		SolarSystemName:   "", // 同上
		KillmailTime:      km.KillmailTime,
		CorporationID:     km.CorporationID,
		AllianceID:        km.AllianceID,
		RecommendedAmount: recommended,
		FinalAmount:       recommended,
		ReviewStatus:      model.SrpReviewSubmitted,
		PayoutStatus:      model.SrpPayoutNotPaid,
	}

	if err := s.repo.CreateApplication(app); err != nil {
		return nil, err
	}
	return app, nil
}

// ─────────────────────────────────────────────
//  申请列表（管理端）
// ─────────────────────────────────────────────

// SrpApplicationResponse 补损申请响应（含舰队信息）
type SrpApplicationResponse struct {
	model.SrpApplication
	FleetTitle  string `json:"fleet_title,omitempty"`
	FleetFCName string `json:"fleet_fc_name,omitempty"`
	Nickname    string `json:"nickname,omitempty"`
}

// SrpBatchPayoutSummaryResponse 按用户聚合的批量发放摘要
type SrpBatchPayoutSummaryResponse struct {
	UserID            uint    `json:"user_id"`
	Nickname          string  `json:"nickname,omitempty"`
	MainCharacterID   int64   `json:"main_character_id"`
	MainCharacterName string  `json:"main_character_name"`
	TotalAmount       float64 `json:"total_amount"`
	ApplicationCount  int64   `json:"application_count"`
}

// enrichWithFleetInfo 为申请列表填充舰队信息
func (s *SrpService) enrichWithFleetInfo(apps []model.SrpApplication) []SrpApplicationResponse {
	result := make([]SrpApplicationResponse, len(apps))
	userIDSet := make(map[uint]bool)
	// 收集所有非空 fleet_id
	fleetIDSet := make(map[string]bool)
	for _, app := range apps {
		userIDSet[app.UserID] = true
		if app.FleetID != nil && *app.FleetID != "" {
			fleetIDSet[*app.FleetID] = true
		}
	}
	userIDs := make([]uint, 0, len(userIDSet))
	for userID := range userIDSet {
		userIDs = append(userIDs, userID)
	}
	userMap := make(map[uint]model.User)
	if len(userIDs) > 0 {
		users, err := s.userRepo.ListByIDs(userIDs)
		if err == nil {
			userMap = make(map[uint]model.User, len(users))
			for _, user := range users {
				userMap[user.ID] = user
			}
		}
	}
	// 批量查询舰队信息
	fleetIDs := make([]string, 0, len(fleetIDSet))
	for fleetID := range fleetIDSet {
		fleetIDs = append(fleetIDs, fleetID)
	}
	fleetMap := make(map[string]model.Fleet)
	if len(fleetIDs) > 0 {
		fleets, err := s.fleetRepo.ListByIDs(fleetIDs)
		if err == nil {
			for index := range fleets {
				fleet := fleets[index]
				fleetMap[fleet.ID] = fleet
			}
		}
	}
	// 组装响应
	for i, app := range apps {
		resp := SrpApplicationResponse{SrpApplication: app}
		if user, ok := userMap[app.UserID]; ok {
			resp.Nickname = user.Nickname
		}
		if app.FleetID != nil && *app.FleetID != "" {
			if fleet, ok := fleetMap[*app.FleetID]; ok {
				resp.FleetTitle = fleet.Title
				resp.FleetFCName = fleet.FCCharacterName
			}
		}
		result[i] = resp
	}
	return result
}

// ListApplications 管理员端分页查询申请列表
func (s *SrpService) ListApplications(page, pageSize int, filter repository.SrpApplicationFilter) ([]SrpApplicationResponse, int64, error) {
	page = normalizePage(page)
	pageSize = normalizeLedgerPageSize(pageSize)

	apps, total, err := s.repo.ListApplications(page, pageSize, filter)
	if err != nil {
		return nil, 0, err
	}
	return s.enrichWithFleetInfo(apps), total, nil
}

// ListMyApplications 当前用户申请列表
func (s *SrpService) ListMyApplications(userID uint, page, pageSize int) ([]SrpApplicationResponse, int64, error) {
	page = normalizePage(page)
	pageSize = normalizePageSize(pageSize, 20, 100)

	apps, total, err := s.repo.ListMyApplications(userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	return s.enrichWithFleetInfo(apps), total, nil
}

// GetApplication 查询单条申请
func (s *SrpService) GetApplication(id uint) (*SrpApplicationResponse, error) {
	app, err := s.repo.GetApplicationByID(id)
	if err != nil {
		return nil, err
	}
	resp := &SrpApplicationResponse{SrpApplication: *app}
	if user, uerr := s.userRepo.GetByID(app.UserID); uerr == nil {
		resp.Nickname = user.Nickname
	}
	if app.FleetID != nil && *app.FleetID != "" {
		if fleet, ferr := s.fleetRepo.GetByID(*app.FleetID); ferr == nil {
			resp.FleetTitle = fleet.Title
			resp.FleetFCName = fleet.FCCharacterName
		}
	}
	return resp, nil
}

// enrichBatchPayoutSummaryRows 批量填充用户昵称和主人物名到汇总行
func (s *SrpService) enrichBatchPayoutSummaryRows(rows []repository.SrpBatchPayoutSummaryRow) ([]SrpBatchPayoutSummaryResponse, error) {
	if len(rows) == 0 {
		return []SrpBatchPayoutSummaryResponse{}, nil
	}

	userIDs := make([]uint, 0, len(rows))
	for _, row := range rows {
		userIDs = append(userIDs, row.UserID)
	}

	users, err := s.userRepo.ListByIDs(userIDs)
	if err != nil {
		return nil, err
	}
	chars, err := s.charRepo.ListByUserIDs(userIDs)
	if err != nil {
		return nil, err
	}

	userMap := make(map[uint]model.User, len(users))
	for _, user := range users {
		userMap[user.ID] = user
	}

	userChars := make(map[uint][]model.EveCharacter)
	charNameByID := make(map[int64]string, len(chars))
	for _, char := range chars {
		userChars[char.UserID] = append(userChars[char.UserID], char)
		charNameByID[char.CharacterID] = char.CharacterName
	}

	result := make([]SrpBatchPayoutSummaryResponse, 0, len(rows))
	for _, row := range rows {
		resp := SrpBatchPayoutSummaryResponse{
			UserID:           row.UserID,
			TotalAmount:      row.TotalAmount,
			ApplicationCount: row.ApplicationCount,
		}

		if user, ok := userMap[row.UserID]; ok {
			resp.Nickname = user.Nickname
			resp.MainCharacterID = user.PrimaryCharacterID
			resp.MainCharacterName = charNameByID[user.PrimaryCharacterID]
		}
		if resp.MainCharacterName == "" {
			if chars := userChars[row.UserID]; len(chars) > 0 {
				resp.MainCharacterID = chars[0].CharacterID
				resp.MainCharacterName = chars[0].CharacterName
			}
		}

		result = append(result, resp)
	}
	return result, nil
}

// ListBatchPayoutSummary 查询管理端批量发放汇总
func (s *SrpService) ListBatchPayoutSummary() ([]SrpBatchPayoutSummaryResponse, error) {
	rows, err := s.repo.ListBatchPayoutSummary()
	if err != nil {
		return nil, err
	}
	return s.enrichBatchPayoutSummaryRows(rows)
}

// ─────────────────────────────────────────────
//  审批
// ─────────────────────────────────────────────

// ReviewApplicationRequest 审批请求
type ReviewApplicationRequest struct {
	Action      string  `json:"action"       binding:"required,oneof=approve reject"` // "approve" | "reject"
	ReviewNote  string  `json:"review_note"`                                          // 拒绝时必须填写
	FinalAmount float64 `json:"final_amount"`                                         // 批准时可以修改金额
}

type RunFleetAutoApprovalResponse struct {
	CheckedCount  int `json:"checked_count"`
	ApprovedCount int `json:"approved_count"`
	SkippedCount  int `json:"skipped_count"`
}

type RunFleetAutoApprovalRequest struct {
	FleetID string `json:"fleet_id" binding:"required"`
}

func canManualAutoApproveApplication(app *model.SrpApplication, fleet *model.Fleet, selectedFleetID string) bool {
	if app == nil || fleet == nil {
		return false
	}
	if selectedFleetID == "" {
		return false
	}
	if app.ReviewStatus != model.SrpReviewSubmitted {
		return false
	}
	if app.FleetID == nil || *app.FleetID == "" {
		return false
	}
	if *app.FleetID != selectedFleetID {
		return false
	}
	if fleet.ID != *app.FleetID {
		return false
	}
	if fleet.AutoSrpMode != model.FleetAutoSrpAutoApprove {
		return false
	}
	return fleet.FleetConfigID != nil && *fleet.FleetConfigID > 0
}

func applyAutoApprovalToApplication(
	app *model.SrpApplication,
	reviewerID uint,
	recommendedAmount float64,
	finalAmount float64,
	reviewedAt time.Time,
) {
	app.RecommendedAmount = recommendedAmount
	app.FinalAmount = finalAmount
	app.ReviewStatus = model.SrpReviewApproved
	app.ReviewNote = autoApproveReviewNote()
	app.ReviewedBy = &reviewerID
	app.ReviewedAt = &reviewedAt
}

// ReviewApplication 审批补损申请（srp/fc/admin 可操作）
// 支持对已批准/已拒绝的申请重新审批（编辑/重新拒绝）
func (s *SrpService) ReviewApplication(reviewerID uint, appID uint, req *ReviewApplicationRequest) (*model.SrpApplication, error) {
	app, err := s.repo.GetApplicationByID(appID)
	if err != nil {
		return nil, errors.New("申请不存在")
	}
	// 已发放的申请不允许重新审批
	if app.PayoutStatus == model.SrpPayoutPaid {
		return nil, errors.New("该申请已发放，不能修改审批状态")
	}
	if req.Action == "reject" && req.ReviewNote == "" {
		return nil, errors.New("拒绝时必须填写审批备注")
	}

	now := time.Now()
	app.ReviewedBy = &reviewerID
	app.ReviewedAt = &now
	app.ReviewNote = req.ReviewNote

	switch req.Action {
	case "approve":
		app.ReviewStatus = model.SrpReviewApproved
		if req.FinalAmount > 0 {
			app.FinalAmount = req.FinalAmount
		}
	case "reject":
		app.ReviewStatus = model.SrpReviewRejected
	}

	if err := s.repo.UpdateApplication(app); err != nil {
		return nil, err
	}
	return app, nil
}

func (s *SrpService) RunFleetAutoApproval(reviewerID uint, fleetID string) (*RunFleetAutoApprovalResponse, error) {
	if fleetID == "" {
		return nil, errors.New("fleet_id 不能为空")
	}

	apps, err := s.repo.ListSubmittedLinkedApplicationsByFleet(fleetID)
	if err != nil {
		return nil, err
	}

	result := &RunFleetAutoApprovalResponse{CheckedCount: len(apps)}
	if len(apps) == 0 {
		return result, nil
	}

	autoSrpSvc := NewAutoSrpService()
	contextCache := make(map[string]*autoSRPFleetContext)
	reviewedAt := time.Now()

	for i := range apps {
		app := &apps[i]
		if app.FleetID == nil || *app.FleetID == "" {
			result.SkippedCount++
			continue
		}

		ctx, ok := contextCache[*app.FleetID]
		if !ok {
			builtCtx, ctxErr := autoSrpSvc.buildFleetContext(*app.FleetID)
			if ctxErr != nil {
				contextCache[*app.FleetID] = nil
			} else {
				contextCache[*app.FleetID] = builtCtx
			}
			ctx = contextCache[*app.FleetID]
		}

		var fleet *model.Fleet
		if ctx != nil {
			fleet = ctx.fleet
		}
		if !canManualAutoApproveApplication(app, fleet, fleetID) {
			result.SkippedCount++
			continue
		}

		recommendedAmount, finalAmount, _, eligible := autoSrpSvc.evaluateApplicationWithContext(ctx, app)
		if !eligible {
			result.SkippedCount++
			continue
		}

		applyAutoApprovalToApplication(
			app,
			reviewerID,
			recommendedAmount,
			finalAmount,
			reviewedAt,
		)
		if err := s.repo.UpdateApplication(app); err != nil {
			return nil, err
		}
		result.ApprovedCount++
	}

	result.SkippedCount = result.CheckedCount - result.ApprovedCount
	return result, nil
}

// ─────────────────────────────────────────────
//  发放
// ─────────────────────────────────────────────

type SrpBatchFuxiPayoutSummary struct {
	ApplicationCount int     `json:"application_count"`
	UserCount        int     `json:"user_count"`
	TotalISKAmount   float64 `json:"total_isk_amount"`
	TotalFuxiCoin    float64 `json:"total_fuxi_coin"`
}

func normalizeSrpPayoutMode(mode string) string {
	switch mode {
	case "", SrpPayoutModeManualTransfer:
		return SrpPayoutModeManualTransfer
	case SrpPayoutModeFuxiCoin:
		return SrpPayoutModeFuxiCoin
	default:
		return ""
	}
}

func convertSrpAmountToFuxiCoin(iskAmount float64) float64 {
	return math.Round((iskAmount/1_000_000)*100) / 100
}

func buildSrpPayoutWalletReason(app *model.SrpApplication, fleetTitle string) string {
	shipName := strings.TrimSpace(app.ShipName)
	if shipName == "" {
		shipName = fmt.Sprintf("TypeID:%d", app.ShipTypeID)
	}
	reason := fmt.Sprintf("SRP#%d %s", app.ID, shipName)
	if strings.TrimSpace(fleetTitle) != "" {
		reason = fmt.Sprintf("%s | %s", reason, strings.TrimSpace(fleetTitle))
	}
	return reason
}

func buildSrpPayoutWalletRefID(appID uint) string {
	return fmt.Sprintf("srp:%d", appID)
}

func markSrpApplicationPaid(app *model.SrpApplication, payerID uint, paidAt time.Time) {
	app.PayoutStatus = model.SrpPayoutPaid
	app.PaidBy = &payerID
	app.PaidAt = &paidAt
}

func (s *SrpService) resolveSrpPayoutShipName(app *model.SrpApplication) string {
	if strings.TrimSpace(app.ShipName) != "" {
		return strings.TrimSpace(app.ShipName)
	}
	names, err := s.sdeRepo.GetNames(map[string][]int{"type": {int(app.ShipTypeID)}}, "zh")
	if err != nil {
		return fmt.Sprintf("TypeID:%d", app.ShipTypeID)
	}
	if name := strings.TrimSpace(names["type"][int(app.ShipTypeID)]); name != "" {
		return name
	}
	return fmt.Sprintf("TypeID:%d", app.ShipTypeID)
}

func (s *SrpService) resolveSrpPayoutFleetTitle(app *model.SrpApplication) string {
	if app.FleetID == nil || *app.FleetID == "" {
		return ""
	}
	fleet, err := s.fleetRepo.GetByID(*app.FleetID)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(fleet.Title)
}

func (s *SrpService) buildSrpWalletPayoutData(app *model.SrpApplication) (float64, string, string, error) {
	fuxiCoinAmount := convertSrpAmountToFuxiCoin(app.FinalAmount)
	if fuxiCoinAmount <= 0 {
		return 0, "", "", errors.New("最终金额必须大于 0 才能发放伏羲币")
	}

	appCopy := *app
	appCopy.ShipName = s.resolveSrpPayoutShipName(app)
	fleetTitle := s.resolveSrpPayoutFleetTitle(app)
	reason := buildSrpPayoutWalletReason(&appCopy, fleetTitle)

	return fuxiCoinAmount, reason, buildSrpPayoutWalletRefID(app.ID), nil
}

func (s *SrpService) payoutApplicationWithFuxiCoinTx(tx *gorm.DB, payerID uint, app *model.SrpApplication) error {
	fuxiCoinAmount, reason, refID, err := s.buildSrpWalletPayoutData(app)
	if err != nil {
		return err
	}
	if err := s.walletSvc.ApplyWalletDeltaByOperatorTx(tx, app.UserID, payerID, fuxiCoinAmount, reason, model.WalletRefSrpPayout, refID); err != nil {
		return err
	}
	paidAt := time.Now()
	markSrpApplicationPaid(app, payerID, paidAt)
	return s.repo.UpdateApplicationTx(tx, app)
}

// PayoutRequest 发放请求
type SrpPayoutRequest struct {
	FinalAmount float64 `json:"final_amount"` // 允许最终覆盖金额（0=保持原值）
	Mode        string  `json:"mode"`         // manual_transfer / fuxi_coin
}

// Payout 发放补损（srp/admin 可操作）
func (s *SrpService) Payout(payerID uint, appID uint, req *SrpPayoutRequest) (*model.SrpApplication, error) {
	mode := normalizeSrpPayoutMode(req.Mode)
	if mode == "" {
		return nil, errors.New("无效的发放方式")
	}

	app, err := s.repo.GetApplicationByID(appID)
	if err != nil {
		return nil, errors.New("申请不存在")
	}
	if app.ReviewStatus != model.SrpReviewApproved {
		return nil, errors.New("申请未被批准，无法发放")
	}
	if app.PayoutStatus == model.SrpPayoutPaid {
		return nil, errors.New("该申请已发放，不能重复操作")
	}
	if req.FinalAmount > 0 {
		app.FinalAmount = req.FinalAmount
	}

	if mode == SrpPayoutModeFuxiCoin {
		err := global.DB.Transaction(func(tx *gorm.DB) error {
			lockedApp, lockErr := s.repo.GetApplicationByIDForUpdate(tx, appID)
			if lockErr != nil {
				return errors.New("申请不存在")
			}
			if lockedApp.ReviewStatus != model.SrpReviewApproved {
				return errors.New("申请未被批准，无法发放")
			}
			if lockedApp.PayoutStatus == model.SrpPayoutPaid {
				return errors.New("该申请已发放，不能重复操作")
			}
			if req.FinalAmount > 0 {
				lockedApp.FinalAmount = req.FinalAmount
			}
			if err := s.payoutApplicationWithFuxiCoinTx(tx, payerID, lockedApp); err != nil {
				return err
			}
			*app = *lockedApp
			return nil
		})
		if err != nil {
			return nil, err
		}
		return app, nil
	}

	now := time.Now()
	markSrpApplicationPaid(app, payerID, now)

	if err := s.repo.UpdateApplication(app); err != nil {
		return nil, err
	}
	return app, nil
}

// BatchPayoutByUser 批量发放某用户所有已批准且未发放的 SRP
func (s *SrpService) BatchPayoutByUser(payerID uint, userID uint) (*SrpBatchPayoutSummaryResponse, error) {
	now := time.Now()
	summary, err := s.repo.BatchPayoutApplicationsByUser(userID, payerID, now)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNoApprovedUnpaidBatchPayoutApplications):
			return nil, errors.New("该用户没有可批量发放的 SRP 申请")
		case errors.Is(err, repository.ErrBatchPayoutSelectionChanged):
			return nil, errors.New("待发放申请已变更，请刷新后重试")
		default:
			return nil, err
		}
	}

	enriched, err := s.enrichBatchPayoutSummaryRows([]repository.SrpBatchPayoutSummaryRow{*summary})
	if err != nil {
		return nil, err
	}
	return &enriched[0], nil
}

// BatchPayoutAsFuxiCoin 将全部已批准未发放的申请换算为伏羲币并发放到系统钱包
func (s *SrpService) BatchPayoutAsFuxiCoin(payerID uint) (*SrpBatchFuxiPayoutSummary, error) {
	summary := &SrpBatchFuxiPayoutSummary{}
	userIDs := make(map[uint]struct{})

	err := global.DB.Transaction(func(tx *gorm.DB) error {
		apps, err := s.repo.ListApprovedUnpaidApplicationsForUpdate(tx)
		if err != nil {
			return err
		}
		if len(apps) == 0 {
			return errors.New("暂无可发放的 SRP 申请")
		}

		for i := range apps {
			app := &apps[i]
			fuxiCoinAmount, _, _, buildErr := s.buildSrpWalletPayoutData(app)
			if buildErr != nil {
				return buildErr
			}
			if err := s.payoutApplicationWithFuxiCoinTx(tx, payerID, app); err != nil {
				return err
			}

			summary.ApplicationCount++
			summary.TotalISKAmount += app.FinalAmount
			summary.TotalFuxiCoin += fuxiCoinAmount
			userIDs[app.UserID] = struct{}{}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	summary.UserCount = len(userIDs)
	summary.TotalISKAmount = math.Round(summary.TotalISKAmount*100) / 100
	summary.TotalFuxiCoin = math.Round(summary.TotalFuxiCoin*100) / 100
	return summary, nil
}

// ─────────────────────────────────────────────
//  ESI: Open Information Window
// ─────────────────────────────────────────────

// OpenInfoWindowRequest 打开人物信息窗口请求
type OpenInfoWindowRequest struct {
	CharacterID int64 `json:"character_id" binding:"required"` // 操作者人物 ID（用于获取 token）
	TargetID    int64 `json:"target_id"    binding:"required"` // 要打开信息窗口的目标 ID
}

// OpenInfoWindow 通过 ESI 在客户端打开人物信息窗口
// POST /ui/openwindow/information?target_id=xxx
// 需要 scope: esi-ui.open_window.v1
func (s *SrpService) OpenInfoWindow(userID uint, req *OpenInfoWindowRequest) error {
	// 1. 验证人物属于当前用户
	char, err := s.charRepo.GetByCharacterID(req.CharacterID)
	if err != nil || char.UserID != userID {
		return errors.New("人物不属于当前用户或不存在")
	}

	// 2. 获取有效 token
	ctx := context.Background()
	token, err := s.ssoSvc.GetValidToken(ctx, req.CharacterID)
	if err != nil {
		return fmt.Errorf("获取 token 失败: %w", err)
	}

	// 3. 调用 ESI Open Information Window
	url := fmt.Sprintf("%s/ui/openwindow/information/?target_id=%d", global.Config.EveSSO.ESIBaseURL, req.TargetID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("构建请求失败: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("调用 ESI Open Window 失败: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ESI error %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ─────────────────────────────────────────────
//  快捷申请：获取符合舰队的我的 KM 列表
// ─────────────────────────────────────────────

// FleetKillmailItem KM 列表条目（给前端下拉用）
type FleetKillmailItem struct {
	KillmailID    int64     `json:"killmail_id"`
	KillmailTime  time.Time `json:"killmail_time"`
	ShipTypeID    int64     `json:"ship_type_id"`
	SolarSystemID int64     `json:"solar_system_id"`
	CharacterID   int64     `json:"character_id"`
	VictimName    string    `json:"victim_name"`
}

// GetMyKillmails 获取当前用户所有人物作为受害者的 KM 列表（不限舰队，最近 200 条）
// 若 characterID > 0，则只返回指定人物的 KM（需属于当前用户）
func (s *SrpService) GetMyKillmails(userID uint, characterID int64) ([]FleetKillmailItem, error) {
	chars, err := s.charRepo.ListByUserID(userID)
	if err != nil || len(chars) == 0 {
		return []FleetKillmailItem{}, nil
	}
	charNameMap := make(map[int64]string)
	var charIDs []int64
	if characterID > 0 {
		for _, c := range chars {
			if c.CharacterID == characterID {
				charIDs = []int64{characterID}
				charNameMap[characterID] = c.CharacterName
				break
			}
		}
		if len(charIDs) == 0 {
			return []FleetKillmailItem{}, nil
		}
	} else {
		for _, c := range chars {
			charIDs = append(charIDs, c.CharacterID)
			charNameMap[c.CharacterID] = c.CharacterName
		}
	}

	ckmList, err := s.kmRepo.ListCharacterKillmailsByCharacterIDs(charIDs)
	if err != nil {
		return nil, err
	}
	if len(ckmList) == 0 {
		return []FleetKillmailItem{}, nil
	}

	kmIDs := make([]int64, 0, len(ckmList))
	kmCharMap := make(map[int64]int64) // killmail_id -> character_id
	for _, ckm := range ckmList {
		kmIDs = append(kmIDs, ckm.KillmailID)
		kmCharMap[ckm.KillmailID] = ckm.CharacterID
	}

	// 只查询最近 30 天的 KM
	since := time.Now().AddDate(0, 0, -30)

	kms, err := s.kmRepo.ListKillmailsByIDsSince(kmIDs, since, 200)
	if err != nil {
		return nil, err
	}

	charIDSet := make(map[int64]bool)
	for _, id := range charIDs {
		charIDSet[id] = true
	}

	result := make([]FleetKillmailItem, 0, len(kms))
	for _, km := range kms {
		// 只返回受害者是当前用户人物的 KM
		if !charIDSet[km.CharacterID] {
			continue
		}
		result = append(result, FleetKillmailItem{
			KillmailID:    km.KillmailID,
			KillmailTime:  km.KillmailTime,
			ShipTypeID:    km.ShipTypeID,
			SolarSystemID: km.SolarSystemID,
			CharacterID:   km.CharacterID,
			VictimName:    charNameMap[km.CharacterID],
		})
	}
	return result, nil
}

// GetFleetKillmails 获取符合舰队时间范围和成员资格的当前用户 KM 列表
func (s *SrpService) GetFleetKillmails(userID uint, fleetID string) ([]FleetKillmailItem, error) {
	// 1. 获取舰队信息
	fleet, err := s.fleetRepo.GetByID(fleetID)
	if err != nil {
		return nil, errors.New("舰队不存在")
	}

	// 2. 获取当前用户绑定的人物
	chars, err := s.charRepo.ListByUserID(userID)
	if err != nil || len(chars) == 0 {
		return nil, errors.New("当前用户未绑定人物")
	}

	// 3. 筛选出参与过该舰队的人物 ID
	members, err := s.fleetRepo.ListMembers(fleetID)
	if err != nil {
		return nil, err
	}
	memberSet := make(map[int64]bool)
	for _, m := range members {
		memberSet[m.CharacterID] = true
	}
	var validCharIDs []int64
	charNameMap := make(map[int64]string)
	for _, c := range chars {
		if memberSet[c.CharacterID] {
			validCharIDs = append(validCharIDs, c.CharacterID)
			charNameMap[c.CharacterID] = c.CharacterName
		}
	}
	if len(validCharIDs) == 0 {
		return []FleetKillmailItem{}, nil
	}

	// 4. 查询这些人物在舰队时间段内的 KM
	ckmList, err := s.kmRepo.ListCharacterKillmailsByCharacterIDs(validCharIDs)
	if err != nil {
		return nil, err
	}
	if len(ckmList) == 0 {
		return []FleetKillmailItem{}, nil
	}
	kmIDSet := make(map[int64]int64) // killmail_id -> character_id
	for _, ckm := range ckmList {
		kmIDSet[ckm.KillmailID] = ckm.CharacterID
	}
	kmIDs := make([]int64, 0, len(kmIDSet))
	for kid := range kmIDSet {
		kmIDs = append(kmIDs, kid)
	}

	kms, err := s.kmRepo.ListKillmailsByIDsInTimeRange(kmIDs, fleet.StartAt, fleet.EndAt)
	if err != nil {
		return nil, err
	}

	// 5. 只返回受害人物是用户自己人物的 KM
	result := make([]FleetKillmailItem, 0, len(kms))
	for _, km := range kms {
		if !memberSet[km.CharacterID] {
			continue
		}
		name := charNameMap[km.CharacterID]
		result = append(result, FleetKillmailItem{
			KillmailID:    km.KillmailID,
			KillmailTime:  km.KillmailTime,
			ShipTypeID:    km.ShipTypeID,
			SolarSystemID: km.SolarSystemID,
			CharacterID:   km.CharacterID,
			VictimName:    name,
		})
	}
	return result, nil
}

// ─────────────────────────────────────────────
//  KM 装配详情
// ─────────────────────────────────────────────

// KillmailDetailRequest 请求参数
type KillmailDetailRequest struct {
	KillmailID int64  `json:"killmail_id" binding:"required"`
	Language   string `json:"language"` // "zh" / "en"
}

// KillmailSlotItem 单个槽位中合并后的物品
type KillmailSlotItem struct {
	ItemID   int    `json:"item_id"`
	ItemName string `json:"item_name"`
	Quantity int64  `json:"quantity"`
	Dropped  bool   `json:"dropped"` // true=掉落, false=摧毁
}

// KillmailSlotGroup 按槽位分组
type KillmailSlotGroup struct {
	FlagID   int                `json:"flag_id"`
	FlagName string             `json:"flag_name"`
	FlagText string             `json:"flag_text"`
	OrderID  int                `json:"order_id"`
	Items    []KillmailSlotItem `json:"items"`
}

// KillmailDetailResponse KM 装配详情响应
type KillmailDetailResponse struct {
	KillmailID    int64               `json:"killmail_id"`
	KillmailTime  time.Time           `json:"killmail_time"`
	ShipTypeID    int64               `json:"ship_type_id"`
	ShipName      string              `json:"ship_name"`
	SolarSystemID int64               `json:"solar_system_id"`
	SystemName    string              `json:"system_name"`
	CharacterID   int64               `json:"character_id"`
	CharacterName string              `json:"character_name"`
	JaniceAmount  *float64            `json:"janice_amount"`
	Slots         []KillmailSlotGroup `json:"slots"`
}

// slotCategoryNames 槽位类别的中英文显示名
var slotCategoryNames = map[string]map[string]string{
	"HiSlot":              {"zh": "高槽", "en": "High Slots"},
	"MedSlot":             {"zh": "中槽", "en": "Medium Slots"},
	"LoSlot":              {"zh": "低槽", "en": "Low Slots"},
	"RigSlot":             {"zh": "改装件", "en": "Rig Slots"},
	"SubSystemSlot":       {"zh": "子系统", "en": "Subsystem Slots"},
	"DroneBay":            {"zh": "无人机舱", "en": "Drone Bay"},
	"FighterBay":          {"zh": "战斗机机库", "en": "Fighter Bay"},
	"Cargo":               {"zh": "货柜舱", "en": "Cargo"},
	"FleetHangar":         {"zh": "舰队机库", "en": "Fleet Hangar"},
	"Implant":             {"zh": "植入体", "en": "Implants"},
	"SpecializedFuelBay":  {"zh": "燃料舱", "en": "Fuel Bay"},
	"SpecializedOreHold":  {"zh": "矿石舱", "en": "Ore Hold"},
	"SpecializedAmmoHold": {"zh": "弹药舱", "en": "Ammo Hold"},
}

// GetKillmailDetail 查询 KM 装配详情
func (s *SrpService) GetKillmailDetail(req *KillmailDetailRequest) (*KillmailDetailResponse, error) {
	lang := req.Language
	if lang == "" {
		lang = "zh"
	}

	// 1. 查询 KM 主记录
	kmPtr, err := s.kmRepo.GetKillmailByID(req.KillmailID)
	if err != nil {
		return nil, errors.New("KM 不存在")
	}
	km := *kmPtr

	// 2. 查询 KM 所有物品
	items, err := s.kmRepo.ListKillmailItemsByKillmailID(req.KillmailID)
	if err != nil {
		return nil, err
	}

	// 3. 收集所有 flagID 查 invFlags
	flagIDSet := make(map[int]bool)
	for _, it := range items {
		flagIDSet[it.Flag] = true
	}
	flagIDs := make([]int, 0, len(flagIDSet))
	for fid := range flagIDSet {
		flagIDs = append(flagIDs, fid)
	}
	flags, err := s.sdeRepo.GetFlags(flagIDs)
	if err != nil {
		return nil, err
	}
	flagMap := make(map[int]repository.FlagInfo)
	for _, f := range flags {
		flagMap[f.FlagID] = f
	}

	// 4. 收集所有 typeID（物品 + 舰船），查翻译名
	typeIDSet := make(map[int]bool)
	typeIDSet[int(km.ShipTypeID)] = true
	for _, it := range items {
		typeIDSet[it.ItemID] = true
	}
	typeIDs := make([]int, 0, len(typeIDSet))
	for tid := range typeIDSet {
		typeIDs = append(typeIDs, tid)
	}
	nameMap, err := s.sdeRepo.GetNames(map[string][]int{"type": typeIDs}, lang)
	if err != nil {
		return nil, err
	}
	typeNames := nameMap["type"]

	// 5. 查星系名
	sysNameMap, _ := s.sdeRepo.GetNames(map[string][]int{"solar_system": {int(km.SolarSystemID)}}, lang)
	solarSystemNames := sysNameMap["solar_system"]

	// 6. 查人物名
	charName := ""
	if char, cerr := s.charRepo.GetByCharacterID(km.CharacterID); cerr == nil {
		charName = char.CharacterName
	}

	// 7. 按 (槽位类别, item_id, dropped) 合并，同时按类别分组
	type mergeKey struct {
		Category string
		ItemID   int
		Dropped  bool
	}
	merged := make(map[mergeKey]*KillmailSlotItem)

	catMap := make(map[string]*KillmailSlotGroup)
	catOrder := make([]string, 0)

	for _, it := range items {
		dropped := it.DropType != nil && *it.DropType
		fi := flagMap[it.Flag]
		cat := slotCategory(fi.FlagName)

		// 确保类别组已创建
		if _, ok := catMap[cat]; !ok {
			displayName := fi.FlagText
			if names, exists := slotCategoryNames[cat]; exists {
				if n, ok := names[lang]; ok {
					displayName = n
				}
			}
			catMap[cat] = &KillmailSlotGroup{
				FlagID:   it.Flag,
				FlagName: cat,
				FlagText: displayName,
				OrderID:  fi.OrderID,
				Items:    []KillmailSlotItem{},
			}
			catOrder = append(catOrder, cat)
		} else if fi.OrderID < catMap[cat].OrderID {
			catMap[cat].OrderID = fi.OrderID
		}

		// 按 (category, item_id, dropped) 合并
		key := mergeKey{Category: cat, ItemID: it.ItemID, Dropped: dropped}
		if existing, ok := merged[key]; ok {
			existing.Quantity += it.ItemNum
		} else {
			itemName := typeNames[it.ItemID]
			if itemName == "" {
				itemName = "Unknown"
			}
			si := &KillmailSlotItem{
				ItemID:   it.ItemID,
				ItemName: itemName,
				Quantity: it.ItemNum,
				Dropped:  dropped,
			}
			merged[key] = si
			catMap[cat].Items = append(catMap[cat].Items, *si)
		}
	}

	// 回写合并后数量（指针合并后 slice 中是副本，需要同步）
	for cat, g := range catMap {
		for i := range g.Items {
			key := mergeKey{Category: cat, ItemID: g.Items[i].ItemID, Dropped: g.Items[i].Dropped}
			g.Items[i].Quantity = merged[key].Quantity
		}
	}

	// 按 orderID 排序
	slots := make([]KillmailSlotGroup, 0, len(catOrder))
	for _, cat := range catOrder {
		slots = append(slots, *catMap[cat])
	}
	for i := 1; i < len(slots); i++ {
		for j := i; j > 0 && slots[j].OrderID < slots[j-1].OrderID; j-- {
			slots[j], slots[j-1] = slots[j-1], slots[j]
		}
	}

	shipName := typeNames[int(km.ShipTypeID)]
	if shipName == "" {
		shipName = "Unknown"
	}
	sysName := solarSystemNames[int(km.SolarSystemID)]

	return &KillmailDetailResponse{
		KillmailID:    km.KillmailID,
		KillmailTime:  km.KillmailTime,
		ShipTypeID:    km.ShipTypeID,
		ShipName:      shipName,
		SolarSystemID: km.SolarSystemID,
		SystemName:    sysName,
		CharacterID:   km.CharacterID,
		CharacterName: charName,
		JaniceAmount:  km.JaniceAmount,
		Slots:         slots,
	}, nil
}
