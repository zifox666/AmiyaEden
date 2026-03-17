package service

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

// AutoSrpService 自动 SRP 处理服务
type AutoSrpService struct {
	fleetRepo       *repository.FleetRepository
	fleetConfigRepo *repository.FleetConfigRepository
	srpRepo         *repository.SrpRepository
	charRepo        *repository.EveCharacterRepository
	sdeRepo         *repository.SdeRepository
}

// NewAutoSrpService 创建自动 SRP 服务
func NewAutoSrpService() *AutoSrpService {
	return &AutoSrpService{
		fleetRepo:       repository.NewFleetRepository(),
		fleetConfigRepo: repository.NewFleetConfigRepository(),
		srpRepo:         repository.NewSrpRepository(),
		charRepo:        repository.NewEveCharacterRepository(),
		sdeRepo:         repository.NewSdeRepository(),
	}
}

// ProcessAutoSRP 自动 SRP 处理入口
func (s *AutoSrpService) ProcessAutoSRP(fleetID string) {
	fleet, err := s.fleetRepo.GetByID(fleetID)
	if err != nil {
		global.Logger.Warn("[AutoSRP] 获取舰队失败", zap.String("fleet_id", fleetID), zap.Error(err))
		return
	}

	if fleet.AutoSrpMode == model.FleetAutoSrpDisabled {
		return
	}
	if fleet.FleetConfigID == nil || *fleet.FleetConfigID == 0 {
		global.Logger.Warn("[AutoSRP] 舰队未关联配置，跳过", zap.String("fleet_id", fleetID))
		return
	}

	// 获取舰队配置的装配列表
	fittings, err := s.fleetConfigRepo.ListFittingsByConfigID(*fleet.FleetConfigID)
	if err != nil || len(fittings) == 0 {
		global.Logger.Warn("[AutoSRP] 获取配置装配失败或为空", zap.String("fleet_id", fleetID), zap.Error(err))
		return
	}

	// 按 ship_type_id → fitting 映射
	fittingByShip := make(map[int64]*model.FleetConfigFitting)
	fittingIDs := make([]uint, len(fittings))
	for i := range fittings {
		fittingByShip[fittings[i].ShipTypeID] = &fittings[i]
		fittingIDs[i] = fittings[i].ID
	}

	// 预加载所有配置物品
	configItems, err := s.fleetConfigRepo.ListItemsByFittingIDs(fittingIDs)
	if err != nil {
		global.Logger.Warn("[AutoSRP] 获取配置物品失败", zap.String("fleet_id", fleetID), zap.Error(err))
		return
	}
	itemsByFitting := make(map[uint][]model.FleetConfigFittingItem)
	allItemIDs := make([]uint, 0, len(configItems))
	for _, item := range configItems {
		itemsByFitting[item.FleetConfigFittingID] = append(itemsByFitting[item.FleetConfigFittingID], item)
		allItemIDs = append(allItemIDs, item.ID)
	}

	// 预加载所有替代品
	allReplacements, err := s.fleetConfigRepo.ListReplacementsByItemIDs(allItemIDs)
	if err != nil {
		global.Logger.Warn("[AutoSRP] 获取替代品失败", zap.String("fleet_id", fleetID), zap.Error(err))
		return
	}
	repByItem := make(map[uint][]model.FleetConfigFittingItemReplacement)
	for _, r := range allReplacements {
		repByItem[r.FleetConfigFittingItemID] = append(repByItem[r.FleetConfigFittingItemID], r)
	}

	// 获取舰队成员
	members, err := s.fleetRepo.ListMembers(fleetID)
	if err != nil {
		global.Logger.Warn("[AutoSRP] 获取成员失败", zap.String("fleet_id", fleetID), zap.Error(err))
		return
	}

	for _, member := range members {
		s.processOneMember(fleet, member, fittingByShip, itemsByFitting, repByItem)
	}

	global.Logger.Info("[AutoSRP] 处理完毕",
		zap.String("fleet_id", fleetID),
		zap.Int("members", len(members)),
	)
}

// processOneMember 处理单个成员的自动 SRP
func (s *AutoSrpService) processOneMember(
	fleet *model.Fleet,
	member model.FleetMember,
	fittingByShip map[int64]*model.FleetConfigFitting,
	itemsByFitting map[uint][]model.FleetConfigFittingItem,
	repByItem map[uint][]model.FleetConfigFittingItemReplacement,
) {
	// 查找该成员的受害 KM
	var killmails []model.EveCharacterKillmail
	if err := global.DB.Where(
		"character_id = ? AND victim = ?", member.CharacterID, true,
	).Find(&killmails).Error; err != nil {
		return
	}
	if len(killmails) == 0 {
		return
	}

	// 批量获取 KM 详情，同时按时间范围过滤
	killmailIDs := make([]int64, len(killmails))
	for i, ckm := range killmails {
		killmailIDs[i] = ckm.KillmailID
	}
	var kmList []model.EveKillmailList
	if err := global.DB.Where(
		"kill_mail_id IN ? AND kill_mail_time BETWEEN ? AND ?",
		killmailIDs, fleet.StartAt, fleet.EndAt,
	).Find(&kmList).Error; err != nil {
		return
	}
	kmByID := make(map[int64]model.EveKillmailList, len(kmList))
	for _, km := range kmList {
		kmByID[km.KillmailID] = km
	}

	for _, ckm := range killmails {
		km, ok := kmByID[ckm.KillmailID]
		if !ok {
			continue
		}

		// 查找对应的装配配置
		fitting, ok := fittingByShip[km.ShipTypeID]
		if !ok {
			continue
		}

		// 确定 SRP 金额
		baseAmount := s.getBaseAmount(fitting, km.ShipTypeID)

		// 提交 SRP 申请
		fleetID := fleet.ID
		app := &model.SrpApplication{
			UserID:            member.UserID,
			CharacterID:       member.CharacterID,
			CharacterName:     member.CharacterName,
			KillmailID:        ckm.KillmailID,
			FleetID:           &fleetID,
			Note:              "",
			ShipTypeID:        km.ShipTypeID,
			SolarSystemID:     km.SolarSystemID,
			KillmailTime:      km.KillmailTime,
			CorporationID:     km.CorporationID,
			AllianceID:        km.AllianceID,
			RecommendedAmount: baseAmount,
			FinalAmount:       baseAmount,
			ReviewStatus:      model.SrpReviewPending,
			PayoutStatus:      model.SrpPayoutPending,
		}

		if fleet.AutoSrpMode == model.FleetAutoSrpAutoApprove {
			configItemsForFitting := itemsByFitting[fitting.ID]
			finalAmount, note := s.validateFitting(km.KillmailID, configItemsForFitting, repByItem, baseAmount)
			app.FinalAmount = finalAmount
			app.ReviewStatus = model.SrpReviewApproved
			now := time.Now()
			app.ReviewedAt = &now
			if note != "" {
				app.ReviewNote = "[自动审批-不符] " + note
			} else {
				app.ReviewNote = "[自动审批-符合]"
			}
		}

		if err := s.srpRepo.CreateApplication(app); err != nil {
			if isDuplicateSrpApplicationError(err) {
				continue
			}
			global.Logger.Warn("[AutoSRP] 创建申请失败",
				zap.Int64("killmail_id", ckm.KillmailID),
				zap.Int64("character_id", member.CharacterID),
				zap.Error(err),
			)
		}
	}
}

// isDuplicateSrpApplicationError 检查是否为唯一约束冲突（重复 SRP 申请）
func isDuplicateSrpApplicationError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "UNIQUE constraint") || strings.Contains(msg, "Duplicate entry")
}

// getBaseAmount 获取 SRP 基础金额：配置金额 > 0 则用配置金额，否则查全局价格表
func (s *AutoSrpService) getBaseAmount(fitting *model.FleetConfigFitting, shipTypeID int64) float64 {
	if fitting.SrpAmount > 0 {
		return fitting.SrpAmount
	}
	if price, err := s.srpRepo.GetShipPriceByTypeID(shipTypeID); err == nil {
		return price.Amount
	}
	return 0
}

// validateFitting 验证 KM 装配是否符合配置要求，返回最终金额和不符说明
func (s *AutoSrpService) validateFitting(
	killmailID int64,
	configItems []model.FleetConfigFittingItem,
	repByItem map[uint][]model.FleetConfigFittingItemReplacement,
	baseAmount float64,
) (float64, string) {
	if len(configItems) == 0 {
		return baseAmount, ""
	}

	// 获取 KM 物品
	var kmItems []model.EveKillmailItem
	if err := global.DB.Where("kill_mail_id = ?", killmailID).Find(&kmItems).Error; err != nil {
		return baseAmount, ""
	}

	// 获取 KM 物品的 flag 名称
	flagIDSet := make(map[int]struct{})
	for _, item := range kmItems {
		flagIDSet[item.Flag] = struct{}{}
	}
	flagIDs := make([]int, 0, len(flagIDSet))
	for fid := range flagIDSet {
		flagIDs = append(flagIDs, fid)
	}
	flagInfos, err := s.sdeRepo.GetFlags(flagIDs)
	if err != nil {
		return baseAmount, ""
	}
	flagNameMap := make(map[int]string, len(flagInfos))
	for _, fi := range flagInfos {
		flagNameMap[fi.FlagID] = fi.FlagName
	}

	// 不检查的槽位类别
	skipCategories := map[string]bool{
		"DroneBay":   true,
		"FighterBay": true,
		"Cargo":      true,
	}

	// 按槽位类别统计 KM 物品的 type_id → 数量（合并 destroyed + dropped）
	kmByCategory := make(map[string]map[int]int64) // category → type_id → total_quantity
	for _, item := range kmItems {
		flagName := flagNameMap[item.Flag]
		if flagName == "" {
			continue
		}
		cat := slotCategory(flagName)
		if skipCategories[cat] {
			continue
		}
		if kmByCategory[cat] == nil {
			kmByCategory[cat] = make(map[int]int64)
		}
		kmByCategory[cat][item.ItemID] += item.ItemNum
	}

	// 按槽位类别验证配置物品
	hasHalf := false
	hasNone := false
	var mismatches []string

	for _, cfgItem := range configItems {
		if cfgItem.Importance == model.FittingItemOptional {
			continue
		}

		cfgCat := slotCategory(cfgItem.Flag)
		if skipCategories[cfgCat] {
			continue
		}

		expectedQty := int64(cfgItem.Quantity)
		actualQty := int64(0)
		usedReplacement := false

		catItems := kmByCategory[cfgCat]
		if catItems != nil {
			// 检查原始 type_id
			actualQty = catItems[int(cfgItem.TypeID)]

			// 如果是可替换，也检查替代品
			if cfgItem.Importance == model.FittingItemReplaceable && actualQty < expectedQty {
				reps := repByItem[cfgItem.ID]
				for _, rep := range reps {
					repQty := catItems[int(rep.TypeID)]
					if repQty > 0 {
						usedReplacement = true
						actualQty += repQty
					}
					if actualQty >= expectedQty {
						break
					}
				}
			}
		}

		if actualQty < expectedQty {
			// 装备不足
			mismatches = append(mismatches,
				fmt.Sprintf("type_id=%d: 期望%d 实际%d", cfgItem.TypeID, expectedQty, actualQty),
			)
			switch cfgItem.Penalty {
			case model.FittingPenaltyNone:
				hasNone = true
			case model.FittingPenaltyHalf:
				hasHalf = true
			}
		} else if usedReplacement {
			// 装备数量够但使用了替代品
			switch cfgItem.ReplacementPenalty {
			case model.FittingPenaltyNone:
				hasNone = true
				mismatches = append(mismatches,
					fmt.Sprintf("type_id=%d: 使用替代品(不补损)", cfgItem.TypeID),
				)
			case model.FittingPenaltyHalf:
				hasHalf = true
				mismatches = append(mismatches,
					fmt.Sprintf("type_id=%d: 使用替代品(半额)", cfgItem.TypeID),
				)
			}
		}
	}

	if len(mismatches) == 0 {
		return baseAmount, ""
	}

	note := "装配不符: " + strings.Join(mismatches, "; ")

	if hasNone {
		return 0, note
	}
	if hasHalf {
		return baseAmount * 0.5, note
	}
	return baseAmount, note
}
