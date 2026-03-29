package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"time"
)

// EveCharacterRepository EVE 人物数据访问层
type EveCharacterRepository struct{}

func NewEveCharacterRepository() *EveCharacterRepository {
	return &EveCharacterRepository{}
}

// Create 创建人物记录
func (r *EveCharacterRepository) Create(char *model.EveCharacter) error {
	return global.DB.Create(char).Error
}

// Save 保存（create or update）人物记录
func (r *EveCharacterRepository) Save(char *model.EveCharacter) error {
	return global.DB.Save(char).Error
}

// GetByCharacterID 根据 EVE 人物 ID 查询
func (r *EveCharacterRepository) GetByCharacterID(characterID int64) (*model.EveCharacter, error) {
	var char model.EveCharacter
	err := global.DB.Where("character_id = ?", characterID).First(&char).Error
	return &char, err
}

// ListByUserID 查询某用户绑定的所有人物
func (r *EveCharacterRepository) ListByUserID(userID uint) ([]model.EveCharacter, error) {
	var chars []model.EveCharacter
	err := global.DB.Where("user_id = ?", userID).Find(&chars).Error
	return chars, err
}

// ListByUserIDs 查询多个用户绑定的人物
func (r *EveCharacterRepository) ListByUserIDs(userIDs []uint) ([]model.EveCharacter, error) {
	var chars []model.EveCharacter
	if len(userIDs) == 0 {
		return chars, nil
	}
	err := global.DB.Where("user_id IN ?", userIDs).Find(&chars).Error
	return chars, err
}

// ListByCharacterIDs 查询多个人物
func (r *EveCharacterRepository) ListByCharacterIDs(characterIDs []int64) ([]model.EveCharacter, error) {
	var chars []model.EveCharacter
	if len(characterIDs) == 0 {
		return chars, nil
	}
	err := global.DB.Where("character_id IN ?", characterIDs).Find(&chars).Error
	return chars, err
}

// Update 更新人物信息
func (r *EveCharacterRepository) Update(char *model.EveCharacter) error {
	return global.DB.Save(char).Error
}

// ListAllWithToken 查询所有有 refresh_token 且 token 未失效的人物（用于 ESI 数据刷新队列）
func (r *EveCharacterRepository) ListAllWithToken() ([]model.EveCharacter, error) {
	var chars []model.EveCharacter
	err := global.DB.Where("refresh_token != '' AND refresh_token IS NOT NULL AND token_invalid = false").Find(&chars).Error
	return chars, err
}

// Delete 删除人物记录（硬删除）
func (r *EveCharacterRepository) Delete(id uint) error {
	return global.DB.Unscoped().Delete(&model.EveCharacter{}, id).Error
}

// GetByCharacterName 根据人物名称查询
func (r *EveCharacterRepository) GetByCharacterName(name string) (*model.EveCharacter, error) {
	var char model.EveCharacter
	err := global.DB.Where("character_name = ?", name).First(&char).Error
	return &char, err
}

func (r *EveCharacterRepository) GetLatestUpdatedAtByUserID(userID uint) (*time.Time, error) {
	var value *time.Time
	err := global.DB.Model(&model.EveCharacter{}).
		Where("user_id = ?", userID).
		Select("MAX(updated_at)").
		Scan(&value).Error
	if err != nil {
		return nil, err
	}
	return value, nil
}
