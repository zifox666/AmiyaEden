package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
)

// ContractRepository 角色合同数据访问层
type ContractRepository struct{}

func NewContractRepository() *ContractRepository { return &ContractRepository{} }

// ContractFilter 合同过滤条件
type ContractFilter struct {
	Type   string
	Status string
}

// GetContractsByCharacterIDs 批量获取多个角色的合同，按 date_issued 倒序
func (r *ContractRepository) GetContractsByCharacterIDs(characterIDs []int64) ([]model.EveCharacterContract, error) {
	var contracts []model.EveCharacterContract
	err := global.DB.
		Where("character_id IN ?", characterIDs).
		Order("date_issued DESC").
		Find(&contracts).Error
	return contracts, err
}

// ListContracts 分页查询合同列表
func (r *ContractRepository) ListContracts(page, pageSize int, characterIDs []int64, filter ContractFilter) ([]model.EveCharacterContract, int64, error) {
	db := global.DB.Model(&model.EveCharacterContract{}).
		Where("character_id IN ?", characterIDs)

	if filter.Type != "" {
		db = db.Where("type = ?", filter.Type)
	}
	if filter.Status != "" {
		db = db.Where("status = ?", filter.Status)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var contracts []model.EveCharacterContract
	offset := (page - 1) * pageSize
	err := db.Order("date_issued DESC").
		Offset(offset).Limit(pageSize).
		Find(&contracts).Error
	return contracts, total, err
}

// GetContractByCharacterAndID 按角色ID和合同ID查询单条合同
func (r *ContractRepository) GetContractByCharacterAndID(characterID, contractID int64) (*model.EveCharacterContract, error) {
	var contract model.EveCharacterContract
	err := global.DB.
		Where("character_id = ? AND contract_id = ?", characterID, contractID).
		First(&contract).Error
	if err != nil {
		return nil, err
	}
	return &contract, nil
}

// GetContractItems 查询合同物品列表
func (r *ContractRepository) GetContractItems(contractID int64) ([]model.EveCharacterContractItem, error) {
	var items []model.EveCharacterContractItem
	err := global.DB.Where("contract_id = ?", contractID).Find(&items).Error
	return items, err
}

// GetContractBids 查询合同竞标列表，按金额倒序
func (r *ContractRepository) GetContractBids(contractID int64) ([]model.EveCharacterContractBid, error) {
	var bids []model.EveCharacterContractBid
	err := global.DB.Where("contract_id = ?", contractID).Order("amount DESC").Find(&bids).Error
	return bids, err
}
