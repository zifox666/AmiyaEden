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
