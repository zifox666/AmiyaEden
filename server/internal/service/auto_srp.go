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
	kmRepo          *repository.KillmailRepository
}

type autoSRPFleetContext struct {
	fleet          *model.Fleet
	fittingByShip  map[int64]*model.FleetConfigFitting
	itemsByFitting map[uint][]model.FleetConfigFittingItem
	repByItem      map[uint][]model.FleetConfigFittingItemReplacement
}

// NewAutoSrpService 创建自动 SRP 服务
func NewAutoSrpService() *AutoSrpService {
	return &AutoSrpService{
		fleetRepo:       repository.NewFleetRepository(),
		fleetConfigRepo: repository.NewFleetConfigRepository(),
		srpRepo:         repository.NewSrpRepository(),
		charRepo:        repository.NewEveCharacterRepository(),
		sdeRepo:         repository.NewSdeRepository(),
		kmRepo:          repository.NewKillmailRepository(),
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

	ctx, err := s.buildFleetContext(fleetID)
	if err != nil {
		global.Logger.Warn("[AutoSRP] 构建舰队上下文失败", zap.String("fleet_id", fleetID), zap.Error(err))
		return
	}

	// 获取舰队成员
	members, err := s.fleetRepo.ListMembers(fleetID)
	if err != nil {
		global.Logger.Warn("[AutoSRP] 获取成员失败", zap.String("fleet_id", fleetID), zap.Error(err))
		return
	}

	for _, member := range members {
		s.processOneMember(fleet, member, ctx.fittingByShip, ctx.itemsByFitting, ctx.repByItem)
	}

	global.Logger.Info("[AutoSRP] 处理完毕",
		zap.String("fleet_id", fleetID),
		zap.Int("members", len(members)),
	)
}

func (s *AutoSrpService) buildFleetContext(fleetID string) (*autoSRPFleetContext, error) {
	fleet, err := s.fleetRepo.GetByID(fleetID)
	if err != nil {
		return nil, err
	}
	if fleet.FleetConfigID == nil || *fleet.FleetConfigID == 0 {
		return nil, fmt.Errorf("fleet %s has no fleet config", fleetID)
	}

	fittings, err := s.fleetConfigRepo.ListFittingsByConfigID(*fleet.FleetConfigID)
	if err != nil {
		return nil, err
	}
	if len(fittings) == 0 {
		return nil, fmt.Errorf("fleet %s has no fleet config fittings", fleetID)
	}

	fittingByShip := make(map[int64]*model.FleetConfigFitting, len(fittings))
	fittingIDs := make([]uint, len(fittings))
	for i := range fittings {
		fittingByShip[fittings[i].ShipTypeID] = &fittings[i]
		fittingIDs[i] = fittings[i].ID
	}

	configItems, err := s.fleetConfigRepo.ListItemsByFittingIDs(fittingIDs)
	if err != nil {
		return nil, err
	}
	itemsByFitting := make(map[uint][]model.FleetConfigFittingItem)
	allItemIDs := make([]uint, 0, len(configItems))
	for _, item := range configItems {
		itemsByFitting[item.FleetConfigFittingID] = append(itemsByFitting[item.FleetConfigFittingID], item)
		allItemIDs = append(allItemIDs, item.ID)
	}

	allReplacements, err := s.fleetConfigRepo.ListReplacementsByItemIDs(allItemIDs)
	if err != nil {
		return nil, err
	}
	repByItem := make(map[uint][]model.FleetConfigFittingItemReplacement)
	for _, replacement := range allReplacements {
		repByItem[replacement.FleetConfigFittingItemID] = append(
			repByItem[replacement.FleetConfigFittingItemID],
			replacement,
		)
	}

	return &autoSRPFleetContext{
		fleet:          fleet,
		fittingByShip:  fittingByShip,
		itemsByFitting: itemsByFitting,
		repByItem:      repByItem,
	}, nil
}

func autoApproveReviewNote() string {
	return "补损根据舰队的自动补损设置，已由系统自动批准。"
}

// RecommendSrpAmount 计算 SRP 推荐金额（手动 SRP 与自动 SRP 共用）。
// 若舰队满足装配验证前置条件（mode != disabled、有配置、有匹配装配），
// 则使用配置金额+装配验证；否则回退到全局舰船价格表。
func (s *AutoSrpService) RecommendSrpAmount(shipTypeID int64, killmailID int64, fleetID *string) (float64, string) {
	if fleetID != nil && *fleetID != "" {
		ctx, err := s.buildFleetContext(*fleetID)
		if err == nil && ctx.fleet.AutoSrpMode != model.FleetAutoSrpDisabled {
			if fitting, ok := ctx.fittingByShip[shipTypeID]; ok {
				baseAmount := s.getBaseAmount(fitting, shipTypeID)
				finalAmount, note, _ := s.validateFitting(
					killmailID,
					ctx.itemsByFitting[fitting.ID],
					ctx.repByItem,
					baseAmount,
				)
				return finalAmount, note
			}
		}
	}

	// 回退：全局舰船价格表
	if price, err := s.srpRepo.GetShipPriceByTypeID(shipTypeID); err == nil {
		return price.Amount, ""
	}
	return 0, ""
}

func (s *AutoSrpService) evaluateApplicationWithContext(
	ctx *autoSRPFleetContext,
	app *model.SrpApplication,
) (float64, float64, string, bool) {
	fitting, ok := ctx.fittingByShip[app.ShipTypeID]
	if !ok {
		return 0, 0, "", false
	}

	baseAmount := s.getBaseAmount(fitting, app.ShipTypeID)
	finalAmount, validationNote, skip := s.validateFitting(
		app.KillmailID,
		ctx.itemsByFitting[fitting.ID],
		ctx.repByItem,
		baseAmount,
	)
	return baseAmount, finalAmount, validationNote, !skip && finalAmount > 0
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
	killmails, err := s.kmRepo.ListVictimKillmailsByCharacterID(member.CharacterID)
	if err != nil || len(killmails) == 0 {
		return
	}

	// 批量获取 KM 详情，同时按时间范围过滤
	killmailIDs := make([]int64, len(killmails))
	for i, ckm := range killmails {
		killmailIDs[i] = ckm.KillmailID
	}
	kmList, err := s.kmRepo.ListKillmailsByIDsInTimeRange(killmailIDs, fleet.StartAt, fleet.EndAt)
	if err != nil {
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

		// 确定 SRP 推荐金额（两种模式都执行装配验证）
		baseAmount := s.getBaseAmount(fitting, km.ShipTypeID)
		configItemsForFitting := itemsByFitting[fitting.ID]
		recommendedAmount, _, _ := s.validateFitting(km.KillmailID, configItemsForFitting, repByItem, baseAmount)

		// 推荐金额为 0 时跳过，不产生申请
		if recommendedAmount == 0 {
			continue
		}

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
			RecommendedAmount: recommendedAmount,
			FinalAmount:       recommendedAmount,
			ReviewStatus:      model.SrpReviewSubmitted,
			PayoutStatus:      model.SrpPayoutNotPaid,
		}

		if fleet.AutoSrpMode == model.FleetAutoSrpAutoApprove {
			app.ReviewStatus = model.SrpReviewApproved
			now := time.Now()
			app.ReviewedAt = &now
			app.ReviewNote = autoApproveReviewNote()
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

// validateFitting 验证 KM 装配是否符合配置要求。
// 返回 (最终金额, 不符说明, 是否应跳过该 KM)。
// skip=true 表示存在 penalty=none 的不符项，调用方不应创建/批准该申请。
func (s *AutoSrpService) validateFitting(
	killmailID int64,
	configItems []model.FleetConfigFittingItem,
	repByItem map[uint][]model.FleetConfigFittingItemReplacement,
	baseAmount float64,
) (float64, string, bool) {
	if len(configItems) == 0 {
		return baseAmount, "", false
	}

	// 获取 KM 物品
	kmItems, err := s.kmRepo.ListKillmailItemsByKillmailID(killmailID)
	if err != nil {
		return baseAmount, "", false
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
		return baseAmount, "", false
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
		return baseAmount, "", false
	}

	note := "装配不符: " + strings.Join(mismatches, "; ")

	if hasNone {
		return 0, note, true
	}
	if hasHalf {
		return baseAmount * 0.5, note, false
	}
	return baseAmount, note, false
}
