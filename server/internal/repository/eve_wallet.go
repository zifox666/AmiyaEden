package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"

	"gorm.io/gorm"
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

func (r *EveWalletRepository) getWalletJournalsQuery(db *gorm.DB, characterID int64, refTypes []string) *gorm.DB {
	query := db.Model(&model.EVECharacterWalletJournal{}).Where("character_id = ?", characterID)
	if len(refTypes) > 0 {
		query = query.Where("ref_type IN ?", refTypes)
	}
	return query
}

func (r *EveWalletRepository) getWalletJournalRefTypesQuery(db *gorm.DB, characterID int64) *gorm.DB {
	return db.Model(&model.EVECharacterWalletJournal{}).
		Where("character_id = ?", characterID).
		Distinct().
		Order("ref_type ASC")
}

// GetWalletJournals 分页获取角色钱包流水
func (r *EveWalletRepository) GetWalletJournals(characterID int64, page, pageSize int, refTypes []string) ([]model.EVECharacterWalletJournal, int64, error) {
	var journals []model.EVECharacterWalletJournal
	var total int64

	db := r.getWalletJournalsQuery(global.DB, characterID, refTypes)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.Order("date DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&journals).Error
	if err != nil {
		return nil, 0, err
	}
	return journals, total, nil
}

// ListWalletJournalRefTypes 获取角色钱包流水中出现过的所有交易类型
func (r *EveWalletRepository) ListWalletJournalRefTypes(characterID int64) ([]string, error) {
	var refTypes []string
	err := r.getWalletJournalRefTypesQuery(global.DB, characterID).Pluck("ref_type", &refTypes).Error
	if err != nil {
		return nil, err
	}
	return refTypes, nil
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
