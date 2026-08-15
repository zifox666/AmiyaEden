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

// GetSelectedPrimaryByUserID returns the only character exposed by AT pages.
// SeAT main_char_id takes precedence over the EVE SSO primary selection.
func (r *EveCharacterRepository) GetSelectedPrimaryByUserID(userID uint) (*model.EveCharacter, error) {
	var char model.EveCharacter
	err := global.DB.Table("eve_character ec").
		Select("ec.*").
		Joins(`JOIN "user" u ON u.id = ec.user_id AND u.deleted_at IS NULL`).
		Joins("LEFT JOIN seat_user su ON su.user_id = ec.user_id AND su.deleted_at IS NULL").
		Where("ec.user_id = ?", userID).
		Where(`ec.character_id = COALESCE(NULLIF(su.main_char_id, 0), NULLIF(u.primary_character_id, 0), (SELECT fallback.character_id FROM eve_character fallback WHERE fallback.user_id = ec.user_id ORDER BY fallback.created_at ASC LIMIT 1))`).
		First(&char).Error
	return &char, err
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

// ListRefreshablePrimaryWithToken returns exactly one refreshable main
// character per user. SeAT main_char_id takes precedence over the EVE SSO
// primary character; the earliest bound character is a legacy fallback.
func (r *EveCharacterRepository) ListRefreshablePrimaryWithToken() ([]model.EveCharacter, error) {
	var chars []model.EveCharacter
	refreshable := "((ec.refresh_token != '' AND ec.refresh_token IS NOT NULL AND ec.token_invalid = false) OR (ec.scopes != '' AND ec.scopes IS NOT NULL AND (ec.refresh_token = '' OR ec.refresh_token IS NULL)))"
	mainCharacter := `ec.character_id = COALESCE(
		NULLIF(su.main_char_id, 0),
		NULLIF(u.primary_character_id, 0),
		(SELECT fallback.character_id FROM eve_character fallback WHERE fallback.user_id = ec.user_id ORDER BY fallback.created_at ASC LIMIT 1)
	)`
	err := global.DB.Table("eve_character ec").
		Select("ec.*").
		Joins(`JOIN "user" u ON u.id = ec.user_id AND u.deleted_at IS NULL`).
		Joins("LEFT JOIN seat_user su ON su.user_id = ec.user_id AND su.deleted_at IS NULL").
		Where(refreshable).
		Where(mainCharacter).
		Scan(&chars).Error
	return chars, err
}

// GetRefreshablePrimaryByCharacterID validates that a requested character is
// the selected main character before an on-demand refresh can run.
func (r *EveCharacterRepository) GetRefreshablePrimaryByCharacterID(characterID int64) (*model.EveCharacter, error) {
	var char model.EveCharacter
	err := global.DB.Table("eve_character ec").
		Select("ec.*").
		Joins(`JOIN "user" u ON u.id = ec.user_id AND u.deleted_at IS NULL`).
		Joins("LEFT JOIN seat_user su ON su.user_id = ec.user_id AND su.deleted_at IS NULL").
		Where("ec.character_id = ?", characterID).
		Where(`ec.character_id = COALESCE(NULLIF(su.main_char_id, 0), NULLIF(u.primary_character_id, 0), (SELECT fallback.character_id FROM eve_character fallback WHERE fallback.user_id = ec.user_id ORDER BY fallback.created_at ASC LIMIT 1))`).
		First(&char).Error
	if err != nil {
		return nil, err
	}
	return &char, nil
}

type CharacterSearchItem struct {
	CharacterID   int64
	CharacterName string
}

func (r *EveCharacterRepository) SearchByName(keyword string, limit int) ([]CharacterSearchItem, error) {
	if keyword == "" || limit <= 0 {
		return nil, nil
	}
	var result []CharacterSearchItem
	err := global.DB.Model(&model.EveCharacter{}).
		Select("character_id, character_name").
		Where("character_name ILIKE ?", "%"+keyword+"%").
		Limit(limit).
		Scan(&result).Error
	return result, err
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
