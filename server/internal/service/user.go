package service

import (
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
)

// UserService 用户业务逻辑层
type UserService struct {
	repo *repository.UserRepository
}

func NewUserService() *UserService {
	return &UserService{repo: repository.NewUserRepository()}
}

// GetUserByID 查询用户详情
func (s *UserService) GetUserByID(id uint) (*model.User, error) {
	return s.repo.GetByID(id)
}

// ListUsers 分页获取用户列表
func (s *UserService) ListUsers(page, pageSize int) ([]model.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	return s.repo.List(page, pageSize)
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(user *model.User) error {
	return s.repo.Update(user)
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(id uint) error {
	return s.repo.Delete(id)
}
