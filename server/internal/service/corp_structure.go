package service

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"errors"
	"fmt"
	"math"
	"time"

	"gorm.io/gorm"
)

type CorpStructureService struct {
	repo        *repository.CorpStructureRepository
	charRepo    *repository.EveCharacterRepository
	settingRepo *repository.CorpStructureFuelSettingRepository
	taskRepo    *repository.CorpStructureFuelTaskRepository
	walletSvc   *SysWalletService
}

func NewCorpStructureService() *CorpStructureService {
	return &CorpStructureService{
		repo:        repository.NewCorpStructureRepository(),
		charRepo:    repository.NewEveCharacterRepository(),
		settingRepo: repository.NewCorpStructureFuelSettingRepository(),
		taskRepo:    repository.NewCorpStructureFuelTaskRepository(),
		walletSvc:   NewSysWalletService(),
	}
}

// CorpStructureListRequest 建筑列表请求
type CorpStructureListRequest struct {
	Current         int    `json:"current"           binding:"required,min=1"`
	Size            int    `json:"size"              binding:"required,min=1,max=100"`
	CorpID          int64  `json:"corp_id"`
	State           string `json:"state"`
	FuelExpiresSoon bool   `json:"fuel_expires_soon"`
	Keyword         string `json:"keyword"`
}

type FuelTaskBrief struct {
	ID              uint       `json:"id"`
	Status          string     `json:"status"`
	ClaimerUserID   uint       `json:"claimer_user_id"`
	AddedHours      float64    `json:"added_hours"`
	WalletAmount    float64    `json:"wallet_amount"`
	IskAmount       float64    `json:"isk_amount"`
	IskPayoutStatus string     `json:"isk_payout_status"`
	ClaimedAt       time.Time  `json:"claimed_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
}

type CorpStructureListItem struct {
	model.CorpStructureInfo
	FuelTask          *FuelTaskBrief `json:"fuel_task,omitempty"`
	CanClaim          bool           `json:"can_claim"`
	CanSettle         bool           `json:"can_settle"`
	ClaimDeniedReason string         `json:"claim_denied_reason,omitempty"`
}

type FuelSettingUpsertRequest struct {
	CorporationID        int64    `json:"corporation_id" binding:"required,gt=0"`
	Enabled              bool     `json:"enabled"`
	ClaimMode            string   `json:"claim_mode"`
	ManualStructureIDs   []int64  `json:"manual_structure_ids"`
	ConditionFuelHoursLE *float64 `json:"condition_fuel_hours_le"`
	ConditionStates      []string `json:"condition_states"`
	ContributionUnit     string   `json:"contribution_unit"`
	WalletEnabled        bool     `json:"wallet_enabled"`
	WalletCalcMode       string   `json:"wallet_calc_mode"`
	WalletValue          float64  `json:"wallet_value"`
	IskEnabled           bool     `json:"isk_enabled"`
	IskCalcMode          string   `json:"isk_calc_mode"`
	IskValue             float64  `json:"isk_value"`
}

type FuelSettleResult struct {
	TaskID        uint    `json:"task_id"`
	StructureID   int64   `json:"structure_id"`
	AddedHours    float64 `json:"added_hours"`
	WalletAmount  float64 `json:"wallet_amount"`
	IskAmount     float64 `json:"isk_amount"`
	IskNeedManual bool    `json:"isk_need_manual"`
}

// ListCorpStructures 获取用户可见的军团建筑列表
func (s *CorpStructureService) ListCorpStructures(userID uint, req *CorpStructureListRequest) (interface{}, error) {
	corpID, err := s.resolveCorpID(userID, req.CorpID)
	if err != nil {
		return nil, err
	}
	if corpID == 0 {
		return map[string]interface{}{
			"list":     []interface{}{},
			"total":    0,
			"page":     req.Current,
			"pageSize": req.Size,
		}, nil
	}

	list, total, err := s.repo.ListByCorpID(corpID, req.Current, req.Size, req.State, req.FuelExpiresSoon, req.Keyword)
	if err != nil {
		return nil, err
	}

	setting, _ := s.GetFuelSetting(corpID)

	structureIDs := make([]int64, 0, len(list))
	for _, item := range list {
		structureIDs = append(structureIDs, item.StructureID)
	}
	tasks, _ := s.taskRepo.ListLatestByStructureIDs(structureIDs)
	taskMap := make(map[int64]model.CorpStructureFuelTask, len(tasks))
	for _, t := range tasks {
		taskMap[t.StructureID] = t
	}

	rows := make([]CorpStructureListItem, 0, len(list))
	for _, item := range list {
		row := CorpStructureListItem{CorpStructureInfo: item}
		task, hasTask := taskMap[item.StructureID]
		if hasTask {
			row.FuelTask = &FuelTaskBrief{
				ID:              task.ID,
				Status:          task.Status,
				ClaimerUserID:   task.ClaimerUserID,
				AddedHours:      task.AddedHours,
				WalletAmount:    task.WalletAmount,
				IskAmount:       task.IskAmount,
				IskPayoutStatus: task.IskPayoutStatus,
				ClaimedAt:       task.ClaimedAt,
				CompletedAt:     task.CompletedAt,
			}
		}

		if hasTask && task.Status == model.FuelTaskStatusClaimed {
			row.CanSettle = task.ClaimerUserID == userID
			row.CanClaim = false
			if task.ClaimerUserID != userID {
				row.ClaimDeniedReason = "已被其他成员承接"
			}
		} else {
			row.CanClaim = s.isStructureClaimable(setting, &item)
			if !row.CanClaim {
				row.ClaimDeniedReason = "当前建筑不在可承接范围内"
			}
		}
		rows = append(rows, row)
	}

	return map[string]interface{}{
		"list":     rows,
		"total":    total,
		"page":     req.Current,
		"pageSize": req.Size,
	}, nil
}

// GetUserCorpIDs 获取用户关联的所有军团 ID
func (s *CorpStructureService) GetUserCorpIDs(userID uint) ([]int64, error) {
	return s.repo.GetCorpIDsByUserID(userID)
}

func (s *CorpStructureService) GetFuelSetting(corpID int64) (*model.CorpStructureFuelSetting, error) {
	setting, err := s.settingRepo.GetByCorpID(corpID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &model.CorpStructureFuelSetting{
				CorporationID:    corpID,
				Enabled:          false,
				ClaimMode:        model.FuelClaimModeAll,
				ContributionUnit: model.FuelContributionUnitHour,
				WalletCalcMode:   model.FuelCalcModePerHour,
				IskCalcMode:      model.FuelCalcModePerHour,
			}, nil
		}
		return nil, err
	}
	return setting, nil
}

func (s *CorpStructureService) UpsertFuelSetting(operatorID uint, req *FuelSettingUpsertRequest) error {
	if req.WalletValue < 0 || req.IskValue < 0 {
		return errors.New("贡献值不能为负数")
	}
	if req.ConditionFuelHoursLE != nil && *req.ConditionFuelHoursLE < 0 {
		return errors.New("燃料小时阈值不能为负数")
	}

	switch req.ClaimMode {
	case model.FuelClaimModeAll, model.FuelClaimModeManual, model.FuelClaimModeCondition, model.FuelClaimModeMixed:
	default:
		return errors.New("无效的承接模式")
	}
	switch req.ContributionUnit {
	case "", model.FuelContributionUnitHour:
	default:
		return errors.New("无效的贡献单位")
	}
	switch req.WalletCalcMode {
	case model.FuelCalcModeFixed, model.FuelCalcModePerHour:
	default:
		return errors.New("无效的钱包计算模式")
	}
	switch req.IskCalcMode {
	case model.FuelCalcModeFixed, model.FuelCalcModePerHour:
	default:
		return errors.New("无效的ISK计算模式")
	}

	setting, err := s.settingRepo.GetByCorpID(req.CorporationID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		setting = &model.CorpStructureFuelSetting{CorporationID: req.CorporationID}
	}

	setting.Enabled = req.Enabled
	setting.ClaimMode = req.ClaimMode
	setting.ManualStructureIDs = model.Int64List(uniqueInt64(req.ManualStructureIDs))
	setting.ConditionFuelHoursLE = req.ConditionFuelHoursLE
	setting.ConditionStates = model.StringList(uniqueStrings(req.ConditionStates))
	setting.ContributionUnit = model.FuelContributionUnitHour
	setting.WalletEnabled = req.WalletEnabled
	setting.WalletCalcMode = req.WalletCalcMode
	setting.WalletValue = req.WalletValue
	setting.IskEnabled = req.IskEnabled
	setting.IskCalcMode = req.IskCalcMode
	setting.IskValue = req.IskValue
	setting.UpdatedBy = operatorID
	return s.settingRepo.Save(setting)
}

func (s *CorpStructureService) ClaimFuelTask(userID uint, structureID int64) error {
	info, err := s.repo.GetByStructureID(structureID)
	if err != nil {
		return errors.New("建筑不存在")
	}
	setting, err := s.GetFuelSetting(info.CorporationID)
	if err != nil {
		return err
	}
	if !setting.Enabled {
		return errors.New("当前军团未启用建筑承接")
	}
	if !s.isStructureClaimable(setting, info) {
		return errors.New("当前建筑不在可承接范围内")
	}

	if active, err := s.taskRepo.GetActiveByStructureID(structureID); err == nil && active.ID > 0 {
		return errors.New("该建筑已被承接")
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	before, err := parseFuelExpires(info.FuelExpires)
	if err != nil {
		return errors.New("当前建筑燃料时间不可用，请稍后重试")
	}
	now := time.Now()
	task := &model.CorpStructureFuelTask{
		CorporationID:     info.CorporationID,
		StructureID:       info.StructureID,
		ClaimerUserID:     userID,
		Status:            model.FuelTaskStatusClaimed,
		BeforeFuelExpires: before,
		IskPayoutStatus:   model.IskPayoutStatusPending,
		ClaimedAt:         now,
	}
	return s.taskRepo.Create(task)
}

func (s *CorpStructureService) SettleFuelTask(userID uint, structureID int64) (*FuelSettleResult, error) {
	task, err := s.taskRepo.GetClaimedByStructureAndUser(structureID, userID)
	if err != nil {
		return nil, errors.New("未找到你的承接任务")
	}
	info, err := s.repo.GetByStructureID(structureID)
	if err != nil {
		return nil, errors.New("建筑不存在")
	}

	after, err := parseFuelExpires(info.FuelExpires)
	if err != nil {
		return nil, errors.New("当前建筑燃料时间不可用")
	}

	delta := round2(after.Sub(task.BeforeFuelExpires).Hours())
	if delta <= 0 {
		return nil, errors.New("燃料到期时间未增加，无法结算贡献")
	}

	setting, err := s.GetFuelSetting(info.CorporationID)
	if err != nil {
		return nil, err
	}
	walletAmount := calcContribution(setting.WalletEnabled, setting.WalletCalcMode, setting.WalletValue, delta)
	iskAmount := calcContribution(setting.IskEnabled, setting.IskCalcMode, setting.IskValue, delta)

	now := time.Now()
	task.Status = model.FuelTaskStatusCompleted
	task.AfterFuelExpires = &after
	task.CompletedAt = &now
	task.AddedHours = delta
	task.WalletAmount = walletAmount
	task.IskAmount = iskAmount
	if iskAmount > 0 {
		task.IskPayoutStatus = model.IskPayoutStatusPending
	} else {
		task.IskPayoutStatus = model.IskPayoutStatusWaived
	}

	tx := global.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if walletAmount > 0 {
		reason := fmt.Sprintf("建筑加油贡献 [%s]", info.Name)
		refID := fmt.Sprintf("%d:%d", task.StructureID, task.ID)
		if err := s.walletSvc.ApplyWalletDeltaTx(tx, userID, walletAmount, reason, model.WalletRefStructureFuelReward, refID); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("钱包发放失败: %w", err)
		}
	}

	if err := s.taskRepo.UpdateTx(tx, task); err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &FuelSettleResult{
		TaskID:        task.ID,
		StructureID:   structureID,
		AddedHours:    delta,
		WalletAmount:  walletAmount,
		IskAmount:     iskAmount,
		IskNeedManual: iskAmount > 0,
	}, nil
}

func (s *CorpStructureService) MarkIskPaid(taskID uint, operatorID uint, note string) error {
	task, err := s.taskRepo.GetByID(taskID)
	if err != nil {
		return errors.New("任务不存在")
	}
	if task.Status != model.FuelTaskStatusCompleted {
		return errors.New("任务尚未结算，不能标记发放")
	}
	if task.IskAmount <= 0 {
		return errors.New("该任务没有ISK奖励")
	}
	if task.IskPayoutStatus != model.IskPayoutStatusPending {
		return errors.New("该任务不在待发放状态")
	}
	return s.taskRepo.MarkIskPaid(taskID, operatorID, note)
}

func (s *CorpStructureService) resolveCorpID(userID uint, reqCorpID int64) (int64, error) {
	if reqCorpID > 0 {
		return reqCorpID, nil
	}
	corpIDs, err := s.repo.GetCorpIDsByUserID(userID)
	if err != nil {
		return 0, err
	}
	if len(corpIDs) == 0 {
		return 0, nil
	}
	return corpIDs[0], nil
}

func (s *CorpStructureService) isStructureClaimable(setting *model.CorpStructureFuelSetting, info *model.CorpStructureInfo) bool {
	if setting == nil || !setting.Enabled {
		return false
	}
	manualMatch := containsInt64([]int64(setting.ManualStructureIDs), info.StructureID)
	conditionMatch := s.matchCondition(setting, info)

	switch setting.ClaimMode {
	case model.FuelClaimModeAll:
		return true
	case model.FuelClaimModeManual:
		return manualMatch
	case model.FuelClaimModeCondition:
		return conditionMatch
	case model.FuelClaimModeMixed:
		return manualMatch || conditionMatch
	default:
		return false
	}
}

func (s *CorpStructureService) matchCondition(setting *model.CorpStructureFuelSetting, info *model.CorpStructureInfo) bool {
	if len(setting.ConditionStates) > 0 && !containsString([]string(setting.ConditionStates), info.State) {
		return false
	}
	if setting.ConditionFuelHoursLE != nil {
		exp, err := parseFuelExpires(info.FuelExpires)
		if err != nil {
			return false
		}
		remainHours := exp.Sub(time.Now()).Hours()
		if remainHours > *setting.ConditionFuelHoursLE {
			return false
		}
	}
	return true
}

func parseFuelExpires(v string) (time.Time, error) {
	if v == "" {
		return time.Time{}, errors.New("empty")
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
	}
	for _, layout := range layouts {
		if ts, err := time.Parse(layout, v); err == nil {
			return ts, nil
		}
	}
	return time.Time{}, errors.New("invalid format")
}

func calcContribution(enabled bool, mode string, value float64, deltaHours float64) float64 {
	if !enabled || value <= 0 {
		return 0
	}
	switch mode {
	case model.FuelCalcModeFixed:
		return round2(value)
	case model.FuelCalcModePerHour:
		return round2(value * deltaHours)
	default:
		return 0
	}
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func containsInt64(list []int64, target int64) bool {
	for _, v := range list {
		if v == target {
			return true
		}
	}
	return false
}

func containsString(list []string, target string) bool {
	for _, v := range list {
		if v == target {
			return true
		}
	}
	return false
}

func uniqueInt64(list []int64) []int64 {
	set := make(map[int64]struct{}, len(list))
	result := make([]int64, 0, len(list))
	for _, v := range list {
		if v <= 0 {
			continue
		}
		if _, ok := set[v]; ok {
			continue
		}
		set[v] = struct{}{}
		result = append(result, v)
	}
	return result
}

func uniqueStrings(list []string) []string {
	set := make(map[string]struct{}, len(list))
	result := make([]string, 0, len(list))
	for _, v := range list {
		if v == "" {
			continue
		}
		if _, ok := set[v]; ok {
			continue
		}
		set[v] = struct{}{}
		result = append(result, v)
	}
	return result
}
