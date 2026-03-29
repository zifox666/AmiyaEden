package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"time"

	"gorm.io/gorm/clause"
)

// AssetRepository 人物资产数据访问层
type AssetRepository struct{}

func NewAssetRepository() *AssetRepository { return &AssetRepository{} }

// GetAssetsByCharacterID 获取人物的所有资产
func (r *AssetRepository) GetAssetsByCharacterID(characterID int64) ([]model.EveCharacterAsset, error) {
	var assets []model.EveCharacterAsset
	err := global.DB.Where("character_id = ?", characterID).Find(&assets).Error
	return assets, err
}

// GetAssetsByCharacterIDs 批量获取多个人物的资产
func (r *AssetRepository) GetAssetsByCharacterIDs(characterIDs []int64) ([]model.EveCharacterAsset, error) {
	var assets []model.EveCharacterAsset
	err := global.DB.Where("character_id IN ?", characterIDs).Find(&assets).Error
	return assets, err
}

// GetStructureByID 根据建筑 ID 获取建筑信息（15天内更新的）
func (r *AssetRepository) GetStructureByID(structureID int64) (*model.EveStructure, error) {
	var s model.EveStructure
	fifteenDaysAgo := time.Now().Add(-15 * 24 * time.Hour).Unix()
	err := global.DB.Where("structure_id = ? AND update_at > ?", structureID, fifteenDaysAgo).First(&s).Error
	return &s, err
}

// GetStationByID 根据空间站 ID 获取空间站信息
func (r *AssetRepository) GetStationByID(stationID int64) (*model.EveStation, error) {
	var s model.EveStation
	err := global.DB.Where("station_id = ?", stationID).First(&s).Error
	return &s, err
}

// UpsertStation 创建或更新空间站信息
func (r *AssetRepository) UpsertStation(s *model.EveStation) error {
	return global.DB.Clauses(clause.OnConflict{UpdateAll: true}).Create(s).Error
}
