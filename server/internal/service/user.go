package service

import (
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"errors"
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

// ListUsers 分页获取用户列表（支持筛选）
func (s *UserService) ListUsers(page, pageSize int, filter repository.UserFilter) ([]model.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	return s.repo.List(page, pageSize, filter)
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(user *model.User) error {
	return s.repo.Update(user)
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(id uint) error {
	return s.repo.Delete(id)
}

// UpdateUserRole 修改用户角色（只有 super_admin 可同时将其他用户设置为 super_admin）
func (s *UserService) UpdateUserRole(operatorRole string, targetID uint, newRole string) error {
	// 判断新角色是否合法
	validRoles := map[string]bool{
		model.RoleSuperAdmin: true,
		model.RoleAdmin:      true,
		model.RoleUser:       true,
		model.RoleGuest:      true,
	}
	if !validRoles[newRole] {
		return errors.New("无效的角色")
	}
	// 只有 super_admin 可以授予或取消 super_admin
	if newRole == model.RoleSuperAdmin && operatorRole != model.RoleSuperAdmin {
		return errors.New("只有超级管理员可以授予该角色")
	}
	// 只有 super_admin 可以降级已是 super_admin 的用户
	target, err := s.repo.GetByID(targetID)
	if err != nil {
		return errors.New("用户不存在")
	}
	if target.Role == model.RoleSuperAdmin && operatorRole != model.RoleSuperAdmin {
		return errors.New("只有超级管理员可以修改其他超级管理员的角色")
	}
	return s.repo.UpdateRole(targetID, newRole)
}
