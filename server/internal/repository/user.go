package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
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

// Update 更新用户信息
func (r *UserRepository) Update(user *model.User) error {
	return global.DB.Save(user).Error
}

// Delete 软删除用户
func (r *UserRepository) Delete(id uint) error {
	return global.DB.Delete(&model.User{}, id).Error
}

// UpdateRole 修改用户角色
func (r *UserRepository) UpdateRole(id uint, role string) error {
	return global.DB.Model(&model.User{}).Where("id = ?", id).Update("role", role).Error
}

// UserFilter 用户列表筛选条件
type UserFilter struct {
	Nickname string
	Status   *int
	Role     string
}

// List 分页查询用户列表（支持筛选）
func (r *UserRepository) List(page, pageSize int, filter UserFilter) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	offset := (page - 1) * pageSize
	db := global.DB.Model(&model.User{})

	if filter.Nickname != "" {
		db = db.Where("nickname LIKE ?", "%"+filter.Nickname+"%")
	}
	if filter.Status != nil {
		db = db.Where("status = ?", *filter.Status)
	}
	if filter.Role != "" {
		db = db.Where("role = ?", filter.Role)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Order("id DESC").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}
