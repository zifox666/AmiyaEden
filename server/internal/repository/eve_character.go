package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
)

// EveCharacterRepository EVE 角色数据访问层
type EveCharacterRepository struct{}

func NewEveCharacterRepository() *EveCharacterRepository {
	return &EveCharacterRepository{}
}

// Create 创建角色记录
func (r *EveCharacterRepository) Create(char *model.EveCharacter) error {
	return global.DB.Create(char).Error
}

// Save 保存（create or update）角色记录
func (r *EveCharacterRepository) Save(char *model.EveCharacter) error {
	return global.DB.Save(char).Error
}

// GetByCharacterID 根据 EVE 角色 ID 查询
func (r *EveCharacterRepository) GetByCharacterID(characterID int64) (*model.EveCharacter, error) {
	var char model.EveCharacter
	err := global.DB.Where("character_id = ?", characterID).First(&char).Error
	return &char, err
}

// ListByUserID 查询某用户绑定的所有角色
func (r *EveCharacterRepository) ListByUserID(userID uint) ([]model.EveCharacter, error) {
	var chars []model.EveCharacter
	err := global.DB.Where("user_id = ?", userID).Find(&chars).Error
	return chars, err
}

// GetMainCharByUserID 获取用户最早绑定的角色（主角色）
func (r *EveCharacterRepository) GetMainCharByUserID(userID uint) (*model.EveCharacter, error) {
	var char model.EveCharacter
	err := global.DB.Where("user_id = ?", userID).Order("created_at ASC").First(&char).Error
	if err != nil {
		return nil, err
	}
	return &char, nil
}

// Update 更新角色信息
func (r *EveCharacterRepository) Update(char *model.EveCharacter) error {
	return global.DB.Save(char).Error
}

// ListAllWithToken 查询所有有 refresh_token 且 token 未失效的角色，
// 以及通过 SeAT passthrough 已获取过 scopes 的 SeAT-only 角色（用于 ESI 数据刷新队列）
func (r *EveCharacterRepository) ListAllWithToken() ([]model.EveCharacter, error) {
	var chars []model.EveCharacter
	err := global.DB.Where(
		"(refresh_token != '' AND refresh_token IS NOT NULL AND token_invalid = false) OR (scopes != '' AND scopes IS NOT NULL AND (refresh_token = '' OR refresh_token IS NULL))",
	).Find(&chars).Error
	return chars, err
}

// Delete 删除角色记录（硬删除）
func (r *EveCharacterRepository) Delete(id uint) error {
	return global.DB.Unscoped().Delete(&model.EveCharacter{}, id).Error
}

// GetByCharacterName 根据角色名称查询
func (r *EveCharacterRepository) GetByCharacterName(name string) (*model.EveCharacter, error) {
	var char model.EveCharacter
	err := global.DB.Where("character_name = ?", name).First(&char).Error
	return &char, err
}
