package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
)

type EveWalletRepository struct{}

func NewEveWalletRepository() *EveWalletRepository {
	return &EveWalletRepository{}
}

func (r *EveWalletRepository) GetWallet(characterID int) (*model.EVECharacterWallet, error) {
	var wallet model.EVECharacterWallet
	err := global.DB.Where("character_id = ?", characterID).First(&wallet).Error
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

// GetWalletJournals 分页获取角色钱包流水
func (r *EveWalletRepository) GetWalletJournals(characterID int64, page, pageSize int) ([]model.EVECharacterWalletJournal, int64, error) {
	var journals []model.EVECharacterWalletJournal
	var total int64

	db := global.DB.Model(&model.EVECharacterWalletJournal{}).Where("character_id = ?", characterID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.Order("date DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&journals).Error
	if err != nil {
		return nil, 0, err
	}
	return journals, total, nil
}

// SumBalanceByCharacterIDs 汇总多个角色的钱包余额
func (r *EveWalletRepository) SumBalanceByCharacterIDs(characterIDs []int64) (float64, error) {
	if len(characterIDs) == 0 {
		return 0, nil
	}
	var total float64
	err := global.DB.Model(&model.EVECharacterWallet{}).
		Where("character_id IN ?", characterIDs).
		Select("COALESCE(SUM(balance), 0)").
		Scan(&total).Error
	return total, err
}
