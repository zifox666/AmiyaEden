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

func (r *CorpStructureFuelTaskRepository) GetClaimedByStructureID(structureID int64) (*model.CorpStructureFuelTask, error) {
	var task model.CorpStructureFuelTask
	err := global.DB.Where("structure_id = ? AND status = ?",
		structureID, model.FuelTaskStatusClaimed).
		Order("created_at DESC").
		First(&task).Error
	return &task, err
}

func (r *CorpStructureFuelTaskRepository) UpdateTx(tx *gorm.DB, task *model.CorpStructureFuelTask) error {
	return tx.Save(task).Error
}

func (r *CorpStructureFuelTaskRepository) Update(task *model.CorpStructureFuelTask) error {
	return global.DB.Save(task).Error
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

type FuelTaskListFilter struct {
	CorporationID int64
	ClaimerUserID *uint
	Status        string
	OnlyPending   bool
}

func (r *CorpStructureFuelTaskRepository) ListTasks(page, pageSize int, filter FuelTaskListFilter) ([]model.CorpStructureFuelTaskListItem, int64, error) {
	var (
		list  []model.CorpStructureFuelTaskListItem
		total int64
	)

	db := global.DB.Table("corp_structure_fuel_task AS t").
		Joins("LEFT JOIN corp_structure_info AS s ON s.structure_id = t.structure_id").
		Joins(`LEFT JOIN "user" AS u ON u.id = t.claimer_user_id`)

	if filter.CorporationID > 0 {
		db = db.Where("t.corporation_id = ?", filter.CorporationID)
	}
	if filter.ClaimerUserID != nil {
		db = db.Where("t.claimer_user_id = ?", *filter.ClaimerUserID)
	}
	if filter.Status != "" {
		db = db.Where("t.status = ?", filter.Status)
	}
	if filter.OnlyPending {
		db = db.Where("t.isk_payout_status = ?", model.IskPayoutStatusPending)
	}
	db = db.Where("t.isk_amount > 0")

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.Select(`
		t.id,
		t.corporation_id,
		t.structure_id,
		COALESCE(s.name, '') AS structure_name,
		t.claimer_user_id,
		COALESCE(u.nickname, '') AS claimer_name,
		t.added_hours,
		t.wallet_amount,
		t.isk_amount,
		t.isk_payout_status,
		t.claimed_at,
		t.completed_at,
		t.isk_paid_at
	`).
		Order("t.completed_at DESC NULLS LAST, t.id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&list).Error
	if err != nil {
		return nil, 0, err
	}
	if list == nil {
		list = []model.CorpStructureFuelTaskListItem{}
	}
	return list, total, nil
}
