package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
)

type FittingsRepository struct{}

func NewFittingsRepository() *FittingsRepository {
	return &FittingsRepository{}
}

// ListByCharacterIDs 获取多个人物的装配列表
func (r *FittingsRepository) ListByCharacterIDs(characterIDs []int64) ([]model.EveCharacterFitting, error) {
	var fittings []model.EveCharacterFitting
	err := global.DB.Where("character_id IN ?", characterIDs).Find(&fittings).Error
	return fittings, err
}

// GetItemsByFittingAndCharacter 获取指定装配的物品明细
func (r *FittingsRepository) GetItemsByFittingAndCharacter(fittingID, characterID int64) ([]model.EveCharacterFittingItem, error) {
	var items []model.EveCharacterFittingItem
	err := global.DB.Where("fitting_id = ? AND character_id = ?", fittingID, characterID).Find(&items).Error
	return items, err
}

// GetItemsByCharacterIDs 批量获取多个人物的所有装配物品
func (r *FittingsRepository) GetItemsByCharacterIDs(characterIDs []int64) ([]model.EveCharacterFittingItem, error) {
	var items []model.EveCharacterFittingItem
	err := global.DB.Where("character_id IN ?", characterIDs).Find(&items).Error
	return items, err
}

// SaveFitting 保存装配（先删后插）
func (r *FittingsRepository) SaveFitting(fitting *model.EveCharacterFitting, items []model.EveCharacterFittingItem) error {
	tx := global.DB.Begin()

	// 删除旧的 items
	if err := tx.Where("fitting_id = ? AND character_id = ?", fitting.FittingID, fitting.CharacterID).
		Delete(&model.EveCharacterFittingItem{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	// 删除旧的 fitting
	if err := tx.Where("fitting_id = ? AND character_id = ?", fitting.FittingID, fitting.CharacterID).
		Delete(&model.EveCharacterFitting{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 创建 fitting
	if err := tx.Create(fitting).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 创建 items
	if len(items) > 0 {
		if err := tx.Create(&items).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// DeleteFitting 删除指定装配及其物品
func (r *FittingsRepository) DeleteFitting(fittingID, characterID int64) error {
	tx := global.DB.Begin()

	if err := tx.Where("fitting_id = ? AND character_id = ?", fittingID, characterID).
		Delete(&model.EveCharacterFittingItem{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("fitting_id = ? AND character_id = ?", fittingID, characterID).
		Delete(&model.EveCharacterFitting{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
