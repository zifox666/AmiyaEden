package service

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"errors"
	"time"
)

// SrpService 补损业务逻辑层
type SrpService struct {
	repo      *repository.SrpRepository
	fleetRepo *repository.FleetRepository
	charRepo  *repository.EveCharacterRepository
}

func NewSrpService() *SrpService {
	return &SrpService{
		repo:      repository.NewSrpRepository(),
		fleetRepo: repository.NewFleetRepository(),
		charRepo:  repository.NewEveCharacterRepository(),
	}
}

// ─────────────────────────────────────────────
//  KM 解析辅助
// ─────────────────────────────────────────────

// resolveCharacterKillmail 确认 killmailID 与 characterID 有关联，并返回 EveKillmailList
func resolveCharacterKillmail(killmailID int64, characterID int64) (*model.EveKillmailList, error) {
	// 验证角色-KM 关联关系
	var ckm model.EveCharacterKillmail
	if err := global.DB.Where("character_id = ? AND killmail_id = ?", characterID, killmailID).First(&ckm).Error; err != nil {
		return nil, errors.New("该 KM 不属于指定角色，或尚未被 ESI 刷新任务录入")
	}
	// 加载 KM 详情
	var km model.EveKillmailList
	if err := global.DB.Where("kill_mail_id = ?", killmailID).First(&km).Error; err != nil {
		return nil, errors.New("KM 详情不存在")
	}
	return &km, nil
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
	CharacterID int64   `json:"character_id"  binding:"required"` // 受损角色 ID
	KillmailID  int64   `json:"killmail_id"   binding:"required"` // zkillboard killmail id
	FleetID     *string `json:"fleet_id"`                         // 关联舰队（可选）
	Note        string  `json:"note"`                             // 备注（无舰队时必填）
	FinalAmount float64 `json:"final_amount"`                     // 用户可以修改推荐金额（后台也可修改）
}

// SubmitApplication 提交补损申请
func (s *SrpService) SubmitApplication(userID uint, req *SubmitApplicationRequest) (*model.SrpApplication, error) {
	// 1. 验证角色属于当前用户
	char, err := s.charRepo.GetByCharacterID(req.CharacterID)
	if err != nil || char.UserID != userID {
		return nil, errors.New("角色不属于当前用户或不存在")
	}

	// 2. 无舰队时需要填写备注
	if req.FleetID == nil && req.Note == "" {
		return nil, errors.New("未关联舰队时，备注不能为空")
	}

	// 3. 检查是否重复提交
	if s.repo.ExistsApplicationByKillmail(req.KillmailID, req.CharacterID) {
		return nil, errors.New("该 KM 已提交过补损申请，不能重复提交")
	}

	// 4. 获取 KM 详情（验证角色与 KM 关联）
	km, err := resolveCharacterKillmail(req.KillmailID, req.CharacterID)
	if err != nil {
		return nil, err
	}

	// 5. 确认该 KM 的受害者确实是这个角色
	if km.CharacterID != req.CharacterID {
		return nil, errors.New("该 KM 的受害者不是指定角色，无法申请补损")
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
		// 角色必须是舰队成员
		members, _ := s.fleetRepo.ListMembers(*req.FleetID)
		isMember := false
		for _, m := range members {
			if m.CharacterID == req.CharacterID {
				isMember = true
				break
			}
		}
		if !isMember {
			return nil, errors.New("该角色不是该舰队的成员，无法申请补损")
		}
	}

	// 7. 查找推荐金额
	recommended := 0.0
	if priceRecord, perr := s.repo.GetShipPriceByTypeID(km.ShipTypeID); perr == nil {
		recommended = priceRecord.Amount
	}

	finalAmount := req.FinalAmount
	if finalAmount <= 0 {
		finalAmount = recommended
	}

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
		FinalAmount:       finalAmount,
		ReviewStatus:      model.SrpReviewPending,
		PayoutStatus:      model.SrpPayoutPending,
	}

	if err := s.repo.CreateApplication(app); err != nil {
		return nil, err
	}
	return app, nil
}

// ─────────────────────────────────────────────
//  申请列表（管理端）
// ─────────────────────────────────────────────

// ListApplications 管理员端分页查询申请列表
func (s *SrpService) ListApplications(page, pageSize int, filter repository.SrpApplicationFilter) ([]model.SrpApplication, int64, error) {
	return s.repo.ListApplications(page, pageSize, filter)
}

// ListMyApplications 当前用户申请列表
func (s *SrpService) ListMyApplications(userID uint, page, pageSize int) ([]model.SrpApplication, int64, error) {
	return s.repo.ListMyApplications(userID, page, pageSize)
}

// GetApplication 查询单条申请
func (s *SrpService) GetApplication(id uint) (*model.SrpApplication, error) {
	return s.repo.GetApplicationByID(id)
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

// ReviewApplication 审批补损申请（srp/fc/admin 可操作）
func (s *SrpService) ReviewApplication(reviewerID uint, appID uint, req *ReviewApplicationRequest) (*model.SrpApplication, error) {
	app, err := s.repo.GetApplicationByID(appID)
	if err != nil {
		return nil, errors.New("申请不存在")
	}
	if app.ReviewStatus != model.SrpReviewPending {
		return nil, errors.New("该申请已被审批，不能重复操作")
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

// ─────────────────────────────────────────────
//  发放
// ─────────────────────────────────────────────

// PayoutRequest 发放请求
type SrpPayoutRequest struct {
	FinalAmount float64 `json:"final_amount"` // 允许最终覆盖金额（0=保持原值）
}

// Payout 发放补损（srp/admin 可操作）
func (s *SrpService) Payout(payerID uint, appID uint, req *SrpPayoutRequest) (*model.SrpApplication, error) {
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
	now := time.Now()
	app.PayoutStatus = model.SrpPayoutPaid
	app.PaidBy = &payerID
	app.PaidAt = &now

	if err := s.repo.UpdateApplication(app); err != nil {
		return nil, err
	}
	return app, nil
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
	VictimName    string    `json:"victim_name"`
}

// GetMyKillmails 获取当前用户所有角色作为受害者的 KM 列表（不限舰队，最近 200 条）
// 若 characterID > 0，则只返回指定角色的 KM（需属于当前用户）
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

	var ckmList []model.EveCharacterKillmail
	if err := global.DB.Where("character_id IN ?", charIDs).Find(&ckmList).Error; err != nil {
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

	var kms []model.EveKillmailList
	if err := global.DB.Where("kill_mail_id IN ?", kmIDs).
		Order("kill_mail_time DESC").
		Limit(200).
		Find(&kms).Error; err != nil {
		return nil, err
	}

	charIDSet := make(map[int64]bool)
	for _, id := range charIDs {
		charIDSet[id] = true
	}

	result := make([]FleetKillmailItem, 0, len(kms))
	for _, km := range kms {
		// 只返回受害者是当前用户角色的 KM
		if !charIDSet[km.CharacterID] {
			continue
		}
		result = append(result, FleetKillmailItem{
			KillmailID:    km.KillmailID,
			KillmailTime:  km.KillmailTime,
			ShipTypeID:    km.ShipTypeID,
			SolarSystemID: km.SolarSystemID,
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

	// 2. 获取当前用户绑定的角色
	chars, err := s.charRepo.ListByUserID(userID)
	if err != nil || len(chars) == 0 {
		return nil, errors.New("当前用户未绑定角色")
	}

	// 3. 筛选出参与过该舰队的角色 ID
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

	// 4. 查询这些角色在舰队时间段内的 KM
	var ckmList []model.EveCharacterKillmail
	if err := global.DB.Where("character_id IN ?", validCharIDs).Find(&ckmList).Error; err != nil {
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

	var kms []model.EveKillmailList
	if err := global.DB.Where("kill_mail_id IN ? AND kill_mail_time >= ? AND kill_mail_time <= ?",
		kmIDs, fleet.StartAt, fleet.EndAt).Find(&kms).Error; err != nil {
		return nil, err
	}

	// 5. 只返回受害角色是用户自己角色的 KM
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
			VictimName:    name,
		})
	}
	return result, nil
}
