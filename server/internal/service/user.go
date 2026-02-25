package service

import (
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"errors"
)

type UserService struct {
	repo     *repository.UserRepository
	roleRepo *repository.RoleRepository
}

func NewUserService() *UserService {
	return &UserService{
		repo:     repository.NewUserRepository(),
		roleRepo: repository.NewRoleRepository(),
	}
}

func (s *UserService) GetUserByID(id uint) (*model.User, error) {
	return s.repo.GetByID(id)
}

func (s *UserService) ListUsers(page, pageSize int, filter repository.UserFilter) ([]model.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	return s.repo.List(page, pageSize, filter)
}

func (s *UserService) UpdateUser(user *model.User) error {
	return s.repo.Update(user)
}

func (s *UserService) DeleteUser(id uint) error {
	if _, err := s.repo.GetByID(id); err != nil {
		return errors.New("用户不存在")
	}
	return s.repo.Delete(id)
}
