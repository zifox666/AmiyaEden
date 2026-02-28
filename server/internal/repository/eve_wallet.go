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
