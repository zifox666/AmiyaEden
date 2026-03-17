package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
)

// FittingWithItems 装配条目及其物品明细
type FittingWithItems struct {
	Fitting model.FleetConfigFitting
	Items   []model.FleetConfigFittingItem
}

// FleetConfigRepository 舰队配置数据访问层
type FleetConfigRepository struct{}

func NewFleetConfigRepository() *FleetConfigRepository {
	return &FleetConfigRepository{}
}

// Create 创建舰队配置（含装配条目及物品）
func (r *FleetConfigRepository) Create(config *model.FleetConfig, fittings []FittingWithItems) error {
	tx := global.DB.Begin()

	if err := tx.Create(config).Error; err != nil {
		tx.Rollback()
		return err
	}

	for i := range fittings {
		fittings[i].Fitting.FleetConfigID = config.ID
		if err := tx.Create(&fittings[i].Fitting).Error; err != nil {
			tx.Rollback()
			return err
		}
		for j := range fittings[i].Items {
			fittings[i].Items[j].FleetConfigFittingID = fittings[i].Fitting.ID
		}
		if len(fittings[i].Items) > 0 {
			if err := tx.Create(&fittings[i].Items).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit().Error
}

// GetByID 根据 ID 查询舰队配置
func (r *FleetConfigRepository) GetByID(id uint) (*model.FleetConfig, error) {
	var config model.FleetConfig
	err := global.DB.First(&config, id).Error
	return &config, err
}

// GetFittingByID 根据 ID 查询装配条目
func (r *FleetConfigRepository) GetFittingByID(id uint) (*model.FleetConfigFitting, error) {
	var fitting model.FleetConfigFitting
	err := global.DB.First(&fitting, id).Error
	return &fitting, err
}

// List 分页查询舰队配置列表
func (r *FleetConfigRepository) List(page, pageSize int) ([]model.FleetConfig, int64, error) {
	var configs []model.FleetConfig
	var total int64

	offset := (page - 1) * pageSize
	db := global.DB.Model(&model.FleetConfig{})

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&configs).Error; err != nil {
		return nil, 0, err
	}
	return configs, total, nil
}

// Update 更新舰队配置（替换所有装配条目及物品）
func (r *FleetConfigRepository) Update(config *model.FleetConfig, fittings []FittingWithItems) error {
	tx := global.DB.Begin()

	if err := tx.Save(config).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 获取旧装配 ID 列表
	var oldFittingIDs []uint
	if err := tx.Model(&model.FleetConfigFitting{}).
		Where("fleet_config_id = ?", config.ID).
		Pluck("id", &oldFittingIDs).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 删除旧替代品 + 物品
	if len(oldFittingIDs) > 0 {
		var oldItemIDs []uint
		if err := tx.Model(&model.FleetConfigFittingItem{}).
			Where("fleet_config_fitting_id IN ?", oldFittingIDs).
			Pluck("id", &oldItemIDs).Error; err != nil {
			tx.Rollback()
			return err
		}
		if len(oldItemIDs) > 0 {
			if err := tx.Where("fleet_config_fitting_item_id IN ?", oldItemIDs).
				Delete(&model.FleetConfigFittingItemReplacement{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		if err := tx.Where("fleet_config_fitting_id IN ?", oldFittingIDs).
			Delete(&model.FleetConfigFittingItem{}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 删除旧装配条目
	if err := tx.Where("fleet_config_id = ?", config.ID).
		Delete(&model.FleetConfigFitting{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 创建新装配条目及物品
	for i := range fittings {
		fittings[i].Fitting.FleetConfigID = config.ID
		fittings[i].Fitting.ID = 0
		if err := tx.Create(&fittings[i].Fitting).Error; err != nil {
			tx.Rollback()
			return err
		}
		for j := range fittings[i].Items {
			fittings[i].Items[j].FleetConfigFittingID = fittings[i].Fitting.ID
			fittings[i].Items[j].ID = 0
		}
		if len(fittings[i].Items) > 0 {
			if err := tx.Create(&fittings[i].Items).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit().Error
}

// Delete 删除舰队配置及其装配条目和物品
func (r *FleetConfigRepository) Delete(id uint) error {
	tx := global.DB.Begin()

	var fittingIDs []uint
	if err := tx.Model(&model.FleetConfigFitting{}).
		Where("fleet_config_id = ?", id).
		Pluck("id", &fittingIDs).Error; err != nil {
		tx.Rollback()
		return err
	}

	if len(fittingIDs) > 0 {
		var itemIDs []uint
		if err := tx.Model(&model.FleetConfigFittingItem{}).
			Where("fleet_config_fitting_id IN ?", fittingIDs).
			Pluck("id", &itemIDs).Error; err != nil {
			tx.Rollback()
			return err
		}
		if len(itemIDs) > 0 {
			if err := tx.Where("fleet_config_fitting_item_id IN ?", itemIDs).
				Delete(&model.FleetConfigFittingItemReplacement{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		if err := tx.Where("fleet_config_fitting_id IN ?", fittingIDs).
			Delete(&model.FleetConfigFittingItem{}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Where("fleet_config_id = ?", id).Delete(&model.FleetConfigFitting{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Delete(&model.FleetConfig{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// ListFittingsByConfigID 查询配置下的所有装配条目
func (r *FleetConfigRepository) ListFittingsByConfigID(configID uint) ([]model.FleetConfigFitting, error) {
	var fittings []model.FleetConfigFitting
	err := global.DB.Where("fleet_config_id = ?", configID).Order("id ASC").Find(&fittings).Error
	return fittings, err
}

// ListFittingsByConfigIDs 批量查询多个配置的装配条目
func (r *FleetConfigRepository) ListFittingsByConfigIDs(configIDs []uint) ([]model.FleetConfigFitting, error) {
	var fittings []model.FleetConfigFitting
	if len(configIDs) == 0 {
		return fittings, nil
	}
	err := global.DB.Where("fleet_config_id IN ?", configIDs).Order("id ASC").Find(&fittings).Error
	return fittings, err
}

// ListItemsByFittingIDs 批量查询装配物品
func (r *FleetConfigRepository) ListItemsByFittingIDs(fittingIDs []uint) ([]model.FleetConfigFittingItem, error) {
	var items []model.FleetConfigFittingItem
	if len(fittingIDs) == 0 {
		return items, nil
	}
	err := global.DB.Where("fleet_config_fitting_id IN ?", fittingIDs).Order("id ASC").Find(&items).Error
	return items, err
}

// ListReplacementsByItemIDs 批量查询装备替代品
func (r *FleetConfigRepository) ListReplacementsByItemIDs(itemIDs []uint) ([]model.FleetConfigFittingItemReplacement, error) {
	var reps []model.FleetConfigFittingItemReplacement
	if len(itemIDs) == 0 {
		return reps, nil
	}
	err := global.DB.Where("fleet_config_fitting_item_id IN ?", itemIDs).Order("id ASC").Find(&reps).Error
	return reps, err
}

// UpdateItemSettings 批量更新装配物品的重要性、惩罚及替代品
func (r *FleetConfigRepository) UpdateItemSettings(fittingID uint, updates []ItemSettingUpdate) error {
	tx := global.DB.Begin()

	for _, u := range updates {
		// 更新 importance、penalty 和 replacement_penalty
		if err := tx.Model(&model.FleetConfigFittingItem{}).
			Where("id = ? AND fleet_config_fitting_id = ?", u.ID, fittingID).
			Updates(map[string]interface{}{
				"importance":          u.Importance,
				"penalty":             u.Penalty,
				"replacement_penalty": u.ReplacementPenalty,
			}).Error; err != nil {
			tx.Rollback()
			return err
		}

		// 重建替代品
		if err := tx.Where("fleet_config_fitting_item_id = ?", u.ID).
			Delete(&model.FleetConfigFittingItemReplacement{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		if len(u.Replacements) > 0 {
			reps := make([]model.FleetConfigFittingItemReplacement, len(u.Replacements))
			for i, typeID := range u.Replacements {
				reps[i] = model.FleetConfigFittingItemReplacement{
					FleetConfigFittingItemID: u.ID,
					TypeID:                   typeID,
				}
			}
			if err := tx.Create(&reps).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit().Error
}

// ItemSettingUpdate 单个物品设置更新
type ItemSettingUpdate struct {
	ID                 uint
	Importance         string
	Penalty            string
	ReplacementPenalty string
	Replacements       []int64
}
