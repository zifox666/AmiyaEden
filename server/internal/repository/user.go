package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UserRepository 用户数据访问层
type UserRepository struct{}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

// Create 创建用户
func (r *UserRepository) Create(user *model.User) error {
	return global.DB.Create(user).Error
}

// GetByID 根据 ID 查询用户
func (r *UserRepository) GetByID(id uint) (*model.User, error) {
	var user model.User
	err := global.DB.First(&user, id).Error
	return &user, err
}

func buildGetByIDForUpdateQuery(db *gorm.DB, id uint) *gorm.DB {
	return db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&model.User{}, id)
}

func (r *UserRepository) GetByIDForUpdateTx(tx *gorm.DB, id uint) (*model.User, error) {
	var user model.User
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, id).Error
	return &user, err
}

// UpdateFields 按字段更新用户信息
func (r *UserRepository) UpdateFields(id uint, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	return global.DB.Model(&model.User{}).Where("id = ?", id).Updates(updates).Error
}

// Update 使用完整模型更新用户信息
func (r *UserRepository) Update(user *model.User) error {
	return global.DB.Save(user).Error
}

// ListAllIDs 返回所有用户 ID（供定时任务批量处理使用）
func (r *UserRepository) ListAllIDs() ([]uint, error) {
	var ids []uint
	err := global.DB.Model(&model.User{}).Pluck("id", &ids).Error
	return ids, err
}

// ListByIDs 根据 ID 列表查询用户
func (r *UserRepository) ListByIDs(ids []uint) ([]model.User, error) {
	var users []model.User
	if len(ids) == 0 {
		return users, nil
	}
	err := global.DB.Where("id IN ?", ids).Find(&users).Error
	return users, err
}

// Delete 软删除用户
func (r *UserRepository) Delete(id uint) error {
	return global.DB.Delete(&model.User{}, id).Error
}

// UpdateRole 修改用户职权（兼容字段）
func (r *UserRepository) UpdateRole(id uint, role string) error {
	return global.DB.Model(&model.User{}).Where("id = ?", id).Update("role", role).Error
}

// UserFilter 用户列表筛选条件
type UserFilter struct {
	Keyword           string // 匹配昵称、QQ 或已绑定人物名（不区分大小写）
	Status            *int
	Role              string  // 单职权筛选，仅匹配当前 user_role 关联
	AllowCorporations []int64 // 非空时只返回属于这些军团的用户
}

// GetByPrimaryCharacterID 根据主人物 ID 查询用户
func (r *UserRepository) GetByPrimaryCharacterID(characterID int64) (*model.User, error) {
	var user model.User
	err := global.DB.Where("primary_character_id = ?", characterID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByQQ 根据 QQ 号码查询用户
func (r *UserRepository) GetByQQ(qq string) (*model.User, error) {
	var user model.User
	err := global.DB.Where("qq = ?", qq).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByDiscordID 根据 Discord ID 查询用户
func (r *UserRepository) GetByDiscordID(discordID string) (*model.User, error) {
	var user model.User
	err := global.DB.Where("discord_id = ?", discordID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func buildUserListQuery(db *gorm.DB, filter UserFilter) *gorm.DB {
	query := db.Model(&model.User{})

	query = applyKeywordLikeFilter(
		query,
		filter.Keyword,
		"LOWER(nickname) LIKE ?",
		"LOWER(qq) LIKE ?",
		`id IN (SELECT DISTINCT user_id FROM eve_character WHERE LOWER(character_name) LIKE ?)`)
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.Role != "" {
		query = query.Where(
			`EXISTS (SELECT 1 FROM user_role WHERE user_role.user_id = "user".id AND user_role.role_code = ?)`,
			filter.Role,
		)
	}
	if len(filter.AllowCorporations) > 0 {
		query = query.Where("id IN (SELECT DISTINCT user_id FROM eve_character WHERE corporation_id IN ?)", filter.AllowCorporations)
	}

	return query
}

func buildUserListSelectQuery(db *gorm.DB, filter UserFilter, page, pageSize int) *gorm.DB {
	offset := (page - 1) * pageSize

	return buildUserListQuery(db, filter).
		Order("last_login_at DESC NULLS LAST").
		Order("id DESC").
		Offset(offset).
		Limit(pageSize)
}

// List 分页查询用户列表（支持筛选）
func (r *UserRepository) List(page, pageSize int, filter UserFilter) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	if err := buildUserListQuery(global.DB, filter).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := buildUserListSelectQuery(global.DB, filter, page, pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}
