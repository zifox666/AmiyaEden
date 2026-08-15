package service

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"amiya-eden/pkg/jwt"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"gorm.io/gorm"
)

type UserService struct {
	repo     *repository.UserRepository
	roleRepo *repository.RoleRepository
	seatRepo *repository.SeatUserRepository
	charRepo *repository.EveCharacterRepository
}

func NewUserService() *UserService {
	return &UserService{
		repo:     repository.NewUserRepository(),
		roleRepo: repository.NewRoleRepository(),
		seatRepo: repository.NewSeatUserRepository(),
		charRepo: repository.NewEveCharacterRepository(),
	}
}

func (s *UserService) GetProfile(userID uint) (*model.ProfileDTO, error) {
	user, err := s.repo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	mainID := user.PrimaryCharacterID
	source := "eve"
	if seatUser, seatErr := s.seatRepo.GetByUserID(userID); seatErr == nil && seatUser.MainCharID > 0 {
		mainID = seatUser.MainCharID
		source = "seat"
	} else if seatErr != nil && !errors.Is(seatErr, gorm.ErrRecordNotFound) {
		return nil, seatErr
	}
	if mainID == 0 {
		source = "none"
	}
	mainName := ""
	if mainID > 0 {
		if character, charErr := s.charRepo.GetByCharacterID(mainID); charErr == nil {
			mainName = character.CharacterName
		} else if !errors.Is(charErr, gorm.ErrRecordNotFound) {
			return nil, charErr
		}
	}
	return &model.ProfileDTO{
		QQ:                      user.QQ,
		Nickname:                user.Nickname,
		TournamentCharacterName: user.TournamentCharName,
		TournamentCharacterID:   user.TournamentCharID,
		UILanguage:              languageOrDefault(user.UILanguage, "zh"),
		GameLanguage:            languageOrDefault(user.GameLanguage, "en"),
		MainCharacterSource:     source,
		MainCharacterID:         mainID,
		MainCharacterName:       mainName,
	}, nil
}

var qqPattern = regexp.MustCompile(`^[0-9]{5,20}$`)

func (s *UserService) UpdateProfile(userID uint, req model.UpdateProfileRequest) (*model.ProfileDTO, error) {
	req.QQ = strings.TrimSpace(req.QQ)
	req.Nickname = strings.TrimSpace(req.Nickname)
	req.TournamentCharacterName = strings.TrimSpace(req.TournamentCharacterName)
	if !qqPattern.MatchString(req.QQ) {
		return nil, errors.New("QQ 必须为 5 至 20 位数字")
	}
	if req.Nickname == "" || len([]rune(req.Nickname)) > 128 {
		return nil, errors.New("昵称不能为空且不能超过 128 个字符")
	}
	if req.TournamentCharacterName == "" || len([]rune(req.TournamentCharacterName)) > 128 {
		return nil, errors.New("比赛服角色名不能为空且不能超过 128 个字符")
	}
	if req.TournamentCharacterID == nil || *req.TournamentCharacterID <= 0 {
		return nil, errors.New("比赛服角色 ID 必须为正整数")
	}
	if !validLanguage(req.UILanguage) || !validLanguage(req.GameLanguage) {
		return nil, errors.New("语言仅支持 zh 或 en")
	}
	if err := s.repo.UpdateProfile(userID, req); err != nil {
		return nil, fmt.Errorf("保存个人资料失败: %w", err)
	}
	return s.GetProfile(userID)
}

func validLanguage(language string) bool { return language == "zh" || language == "en" }

func languageOrDefault(language, fallback string) string {
	if validLanguage(language) {
		return language
	}
	return fallback
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
