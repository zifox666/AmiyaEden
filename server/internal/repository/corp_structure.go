package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"

	"gorm.io/gorm"
)

type CorpStructureRepository struct{}

func NewCorpStructureRepository() *CorpStructureRepository {
	return &CorpStructureRepository{}
}

func (r *CorpStructureRepository) baseListQuery(corpID int64, state string, fuelExpiresSoon bool, keyword string) *gorm.DB {
	db := global.DB.Model(&model.CorpStructureInfo{}).Where("corporation_id = ?", corpID)
	if state != "" {
		db = db.Where("state = ?", state)
	}
	if fuelExpiresSoon {
		db = db.Where("fuel_expires != '' AND fuel_expires::timestamptz < NOW() + INTERVAL '7 days'")
	}
	if keyword != "" {
		db = db.Where("name ILIKE ?", "%"+keyword+"%")
	}
	return db
}

// ListByCorpID 分页查询军团建筑列表，支持按状态、燃料到期、关键词过滤
func (r *CorpStructureRepository) ListByCorpID(corpID int64, page, pageSize int, state string, fuelExpiresSoon bool, keyword string) ([]model.CorpStructureInfo, int64, error) {
	var list []model.CorpStructureInfo
	var total int64

	db := r.baseListQuery(corpID, state, fuelExpiresSoon, keyword)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Offset((page - 1) * pageSize).Limit(pageSize).
		Order("fuel_expires ASC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *CorpStructureRepository) ListAllByCorpID(corpID int64, state string, fuelExpiresSoon bool, keyword string) ([]model.CorpStructureInfo, error) {
	var list []model.CorpStructureInfo
	err := r.baseListQuery(corpID, state, fuelExpiresSoon, keyword).
		Order("fuel_expires ASC").
		Find(&list).Error
	return list, err
}

// GetByStructureID 根据建筑 ID 获取军团建筑详情
func (r *CorpStructureRepository) GetByStructureID(structureID int64) (*model.CorpStructureInfo, error) {
	var info model.CorpStructureInfo
	err := global.DB.Where("structure_id = ?", structureID).First(&info).Error
	return &info, err
}

// GetCorpIDsByUserID 获取用户所有角色关联的军团 ID（去重）
func (r *CorpStructureRepository) GetCorpIDsByUserID(userID uint) ([]int64, error) {
	var corpIDs []int64
	err := global.DB.Model(&model.EveCharacter{}).
		Where("user_id = ? AND corporation_id > 0", userID).
		Distinct("corporation_id").
		Pluck("corporation_id", &corpIDs).Error
	return corpIDs, err
}
