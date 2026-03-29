package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"time"

	"gorm.io/gorm/clause"
)

// CloneRepository 克隆体/植入体数据访问层
type CloneRepository struct{}

func NewCloneRepository() *CloneRepository { return &CloneRepository{} }

// GetCloneBaseInfo 获取人物克隆基础信息
func (r *CloneRepository) GetCloneBaseInfo(characterID int64) (*model.EveCharacterCloneBaseInfo, error) {
	var info model.EveCharacterCloneBaseInfo
	err := global.DB.Where("character_id = ?", characterID).First(&info).Error
	if err != nil {
		return nil, err
	}
	return &info, nil
}

// GetImplants 获取人物所有植入体（含跳跃克隆体植入体）
func (r *CloneRepository) GetImplants(characterID int64) ([]model.EveCharacterImplants, error) {
	var implants []model.EveCharacterImplants
	err := global.DB.Where("character_id = ?", characterID).Find(&implants).Error
	if err != nil {
		return nil, err
	}
	return implants, nil
}

// GetStructureByID 根据建筑 ID 获取建筑信息（15天内更新的）
func (r *CloneRepository) GetStructureByID(structureID int64) (*model.EveStructure, error) {
	var s model.EveStructure
	fifteenDaysAgo := time.Now().Add(-15 * 24 * time.Hour).Unix()
	err := global.DB.Where("structure_id = ? AND update_at > ?", structureID, fifteenDaysAgo).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// UpsertStructure 创建或更新建筑信息
func (r *CloneRepository) UpsertStructure(s *model.EveStructure) error {
	return global.DB.Clauses(clause.OnConflict{UpdateAll: true}).Create(s).Error
}
