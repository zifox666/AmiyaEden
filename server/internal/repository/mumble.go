package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MumbleRepository struct{}

func NewMumbleRepository() *MumbleRepository {
	return &MumbleRepository{}
}

func (r *MumbleRepository) GetByUserID(userID uint) (*model.MumbleAccount, error) {
	var account model.MumbleAccount
	err := global.DB.Where("user_id = ?", userID).First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *MumbleRepository) Create(account *model.MumbleAccount) error {
	return global.DB.Create(account).Error
}

func (r *MumbleRepository) UpdatePasswordAndDisplayName(userID uint, password, displayName string) error {
	return global.DB.Model(&model.MumbleAccount{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"password":     password,
			"display_name": displayName,
		}).Error
}

func (r *MumbleRepository) IsNotFound(err error) bool {
	return err == gorm.ErrRecordNotFound
}

func (r *MumbleRepository) ListRoleGroupMappings(provider string) ([]model.VoiceRoleGroupMapping, error) {
	var mappings []model.VoiceRoleGroupMapping
	err := global.DB.Where("provider = ?", provider).Find(&mappings).Error
	return mappings, err
}

func (r *MumbleRepository) UpsertRoleGroupMappings(mappings []model.VoiceRoleGroupMapping) error {
	if len(mappings) == 0 {
		return nil
	}
	return global.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "provider"},
			{Name: "role_code"},
		},
		DoUpdates: clause.AssignmentColumns([]string{"group_name", "enabled"}),
	}).Create(&mappings).Error
}
