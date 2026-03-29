package service

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"amiya-eden/pkg/jwt"
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

func (s *UserService) ListUsers(page, pageSize int, filter repository.UserFilter) ([]model.UserListItemDTO, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	users, total, err := s.repo.List(page, pageSize, filter)
	if err != nil {
		return nil, 0, err
	}

	// 批量查询每个用户在 user_role 中的角色 code
	ids := make([]uint, len(users))
	for i, u := range users {
		ids[i] = u.ID
	}
	roleMap, _ := s.repo.GetRoleCodesByUserIDs(ids) // 查询失败时退化为空 map

	dtos := make([]model.UserListItemDTO, len(users))
	for i, u := range users {
		roles := roleMap[u.ID]
		if len(roles) == 0 && u.Role != "" {
			// user_role 尚无记录时，降级使用 user.role 历史字段
			roles = []string{u.Role}
		}
		dtos[i] = model.UserListItemDTO{
			ID:                 u.ID,
			Nickname:           u.Nickname,
			Avatar:             u.Avatar,
			Status:             u.Status,
			Role:               u.Role,
			Roles:              roles,
			PrimaryCharacterID: u.PrimaryCharacterID,
			LastLoginAt:        u.LastLoginAt,
			LastLoginIP:        u.LastLoginIP,
			CreatedAt:          u.CreatedAt,
			UpdatedAt:          u.UpdatedAt,
		}
	}
	return dtos, total, nil
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

// ImpersonateUser 以指定用户身份生成 JWT（仅超级管理员可用）
func (s *UserService) ImpersonateUser(id uint) (string, *model.User, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return "", nil, errors.New("用户不存在")
	}
	token, err := jwt.GenerateToken(user.ID, user.PrimaryCharacterID, user.Role, global.Config.JWT.ExpireDay)
	if err != nil {
		return "", nil, err
	}
	return token, user, nil
}
