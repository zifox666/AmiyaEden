package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"time"
)

// LotteryRepository 抽奖数据访问层
type LotteryRepository struct{}

func NewLotteryRepository() *LotteryRepository {
	return &LotteryRepository{}
}

// ─────────────────────────────────────────────
//  抽奖活动
// ─────────────────────────────────────────────

// CreateActivity 创建抽奖活动
func (r *LotteryRepository) CreateActivity(a *model.ShopLotteryActivity) error {
	return global.DB.Create(a).Error
}

// UpdateActivity 更新抽奖活动
func (r *LotteryRepository) UpdateActivity(a *model.ShopLotteryActivity) error {
	return global.DB.Save(a).Error
}

// DeleteActivity 删除抽奖活动（软删除）
func (r *LotteryRepository) DeleteActivity(id uint) error {
	return global.DB.Delete(&model.ShopLotteryActivity{}, id).Error
}

// GetActivityByID 根据 ID 获取活动（含奖品列表）
func (r *LotteryRepository) GetActivityByID(id uint) (*model.ShopLotteryActivity, error) {
	var a model.ShopLotteryActivity
	if err := global.DB.Preload("Prizes").First(&a, id).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

// ListActivities 分页查询抽奖活动
func (r *LotteryRepository) ListActivities(page, pageSize int, adminMode bool) ([]model.ShopLotteryActivity, int64, error) {
	var list []model.ShopLotteryActivity
	var total int64
	offset := (page - 1) * pageSize

	db := global.DB.Model(&model.ShopLotteryActivity{}).Preload("Prizes")
	if !adminMode {
		now := time.Now()
		db = db.Where("status = ?", model.LotteryStatusActive).
			Where("(start_at IS NULL OR start_at <= ?)", now).
			Where("(end_at IS NULL OR end_at >= ?)", now)
	}
	db = db.Order("sort_order DESC, created_at DESC")

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ─────────────────────────────────────────────
//  抽奖奖品
// ─────────────────────────────────────────────

// CreatePrize 创建奖品
func (r *LotteryRepository) CreatePrize(p *model.ShopLotteryPrize) error {
	return global.DB.Create(p).Error
}

// UpdatePrize 更新奖品
func (r *LotteryRepository) UpdatePrize(p *model.ShopLotteryPrize) error {
	return global.DB.Save(p).Error
}

// DeletePrize 删除奖品
func (r *LotteryRepository) DeletePrize(id uint) error {
	return global.DB.Delete(&model.ShopLotteryPrize{}, id).Error
}

// ListPrizesByActivity 获取某活动的所有奖品
func (r *LotteryRepository) ListPrizesByActivity(activityID uint) ([]model.ShopLotteryPrize, error) {
	var prizes []model.ShopLotteryPrize
	if err := global.DB.Where("activity_id = ?", activityID).Find(&prizes).Error; err != nil {
		return nil, err
	}
	return prizes, nil
}

// GetPrizeByID 根据 ID 获取奖品
func (r *LotteryRepository) GetPrizeByID(id uint) (*model.ShopLotteryPrize, error) {
	var p model.ShopLotteryPrize
	if err := global.DB.First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

// IncrementPrizeDrawnCount 原子递增奖品已抽出数量
func (r *LotteryRepository) IncrementPrizeDrawnCount(prizeID uint) error {
	return global.DB.Model(&model.ShopLotteryPrize{}).
		Where("id = ?", prizeID).
		UpdateColumn("drawn_count", global.DB.Raw("drawn_count + 1")).Error
}

// ─────────────────────────────────────────────
//  抽奖记录
// ─────────────────────────────────────────────

// CreateRecord 记录抽奖结果
func (r *LotteryRepository) CreateRecord(rec *model.ShopLotteryRecord) error {
	return global.DB.Create(rec).Error
}

// UpdateRecordDeliveryStatus 更新抽奖记录发放状态
func (r *LotteryRepository) UpdateRecordDeliveryStatus(id uint, status string) error {
	return global.DB.Model(&model.ShopLotteryRecord{}).Where("id = ?", id).
		UpdateColumn("delivery_status", status).Error
}

// ListRecords 分页查询抽奖记录
func (r *LotteryRepository) ListRecords(page, pageSize int, userID *uint, activityID *uint) ([]model.ShopLotteryRecord, int64, error) {
	var list []model.ShopLotteryRecord
	var total int64
	offset := (page - 1) * pageSize

	db := global.DB.Model(&model.ShopLotteryRecord{}).Order("created_at DESC")
	if userID != nil {
		db = db.Where("user_id = ?", *userID)
	}
	if activityID != nil {
		db = db.Where("activity_id = ?", *activityID)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
