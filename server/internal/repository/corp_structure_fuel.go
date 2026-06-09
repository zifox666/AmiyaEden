package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"

	"gorm.io/gorm"
)

type CorpStructureFuelSettingRepository struct{}

func NewCorpStructureFuelSettingRepository() *CorpStructureFuelSettingRepository {
	return &CorpStructureFuelSettingRepository{}
}

func (r *CorpStructureFuelSettingRepository) GetByCorpID(corpID int64) (*model.CorpStructureFuelSetting, error) {
	var setting model.CorpStructureFuelSetting
	err := global.DB.Where("corporation_id = ?", corpID).First(&setting).Error
	return &setting, err
}

func (r *CorpStructureFuelSettingRepository) Save(setting *model.CorpStructureFuelSetting) error {
	return global.DB.Save(setting).Error
}

type CorpStructureFuelTaskRepository struct{}

func NewCorpStructureFuelTaskRepository() *CorpStructureFuelTaskRepository {
	return &CorpStructureFuelTaskRepository{}
}

func (r *CorpStructureFuelTaskRepository) Create(task *model.CorpStructureFuelTask) error {
	return global.DB.Create(task).Error
}

func (r *CorpStructureFuelTaskRepository) GetByID(taskID uint) (*model.CorpStructureFuelTask, error) {
	var task model.CorpStructureFuelTask
	err := global.DB.Where("id = ?", taskID).First(&task).Error
	return &task, err
}

func (r *CorpStructureFuelTaskRepository) GetActiveByStructureID(structureID int64) (*model.CorpStructureFuelTask, error) {
	var task model.CorpStructureFuelTask
	err := global.DB.Where("structure_id = ? AND status = ?", structureID, model.FuelTaskStatusClaimed).
		Order("created_at DESC").
		First(&task).Error
	return &task, err
}

func (r *CorpStructureFuelTaskRepository) GetClaimedByStructureAndUser(structureID int64, userID uint) (*model.CorpStructureFuelTask, error) {
	var task model.CorpStructureFuelTask
	err := global.DB.Where("structure_id = ? AND claimer_user_id = ? AND status = ?",
		structureID, userID, model.FuelTaskStatusClaimed).
		Order("created_at DESC").
		First(&task).Error
	return &task, err
}

func (r *CorpStructureFuelTaskRepository) UpdateTx(tx *gorm.DB, task *model.CorpStructureFuelTask) error {
	return tx.Save(task).Error
}

func (r *CorpStructureFuelTaskRepository) MarkIskPaid(taskID uint, operatorID uint, note string) error {
	return global.DB.Model(&model.CorpStructureFuelTask{}).
		Where("id = ? AND isk_payout_status = ?", taskID, model.IskPayoutStatusPending).
		Updates(map[string]interface{}{
			"isk_payout_status": model.IskPayoutStatusPaid,
			"isk_paid_by":       operatorID,
			"isk_paid_at":       gorm.Expr("NOW()"),
			"isk_payout_note":   note,
		}).Error
}

func (r *CorpStructureFuelTaskRepository) ListLatestByStructureIDs(structureIDs []int64) ([]model.CorpStructureFuelTask, error) {
	if len(structureIDs) == 0 {
		return []model.CorpStructureFuelTask{}, nil
	}

	sub := global.DB.Model(&model.CorpStructureFuelTask{}).
		Select("structure_id, MAX(created_at) as max_created_at").
		Where("structure_id IN ?", structureIDs).
		Group("structure_id")

	var list []model.CorpStructureFuelTask
	err := global.DB.Table("corp_structure_fuel_task t").
		Select("t.*").
		Joins("JOIN (?) latest ON t.structure_id = latest.structure_id AND t.created_at = latest.max_created_at", sub).
		Scan(&list).Error
	return list, err
}
