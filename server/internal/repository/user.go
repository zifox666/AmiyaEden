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

// List 分页查询用户列表
func (r *UserRepository) List(page, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	offset := (page - 1) * pageSize
	db := global.DB.Model(&model.User{})

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}
