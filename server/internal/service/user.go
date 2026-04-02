package service

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"amiya-eden/pkg/jwt"
	"errors"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"gorm.io/gorm"
)

type UserService struct {
	repo      *repository.UserRepository
	roleRepo  *repository.RoleRepository
	charRepo  *repository.EveCharacterRepository
	skillRepo *repository.EveSkillRepository
}

const (
	maxUserNicknameLength = 20
	maxUserContactLength  = 20
)

type UserPatch struct {
	Nickname  *string
	QQ        *string
	DiscordID *string
	Status    *int8
}

func NewUserService() *UserService {
	return &UserService{
		repo:      repository.NewUserRepository(),
		roleRepo:  repository.NewRoleRepository(),
		charRepo:  repository.NewEveCharacterRepository(),
		skillRepo: repository.NewEveSkillRepository(),
	}
}

func (s *UserService) GetUserByID(id uint) (*model.User, error) {
	return s.repo.GetByID(id)
}

func (s *UserService) ValidateCurrentUserBootstrap(user *model.User, characters []model.EveCharacter) error {
	if user == nil {
		return errors.New("用户不存在")
	}
	return validatePrimaryCharacterTokenHealth(*user, characters)
}

func (s *UserService) ListUsers(page, pageSize int, filter repository.UserFilter) ([]model.UserListItem, int64, error) {
	page = normalizePage(page)
	pageSize = normalizeLedgerPageSize(pageSize)
	users, total, err := s.repo.List(page, pageSize, filter)
	if err != nil {
		return nil, 0, err
	}
	return s.buildUserListItems(users), total, nil
}

func (s *UserService) UpdateUserByAdmin(id uint, operatorRoles []string, patch UserPatch) error {
	current, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("用户不存在")
	}
	targetRoles, err := s.roleRepo.GetUserRoleCodes(id)
	if err != nil {
		return err
	}
	if err := validateManageUserPermission(operatorRoles, targetRoles); err != nil {
		return err
	}
	next, updates, err := buildUserPatchUpdates(current, patch, false)
	if err != nil {
		return err
	}
	if patch.QQ != nil || patch.DiscordID != nil {
		if err := s.validateUniqueContacts(next); err != nil {
			return err
		}
	}
	return s.repo.UpdateFields(id, updates)
}

func (s *UserService) UpdateCurrentProfile(id uint, patch UserPatch) (*model.User, error) {
	current, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	next, updates, err := buildUserPatchUpdates(current, patch, true)
	if err != nil {
		return nil, err
	}
	if err := s.validateUniqueContacts(next); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateFields(id, updates); err != nil {
		return nil, err
	}
	return s.repo.GetByID(id)
}

func (s *UserService) DeleteUser(id uint, operatorRoles []string) error {
	if _, err := s.repo.GetByID(id); err != nil {
		return errors.New("用户不存在")
	}
	targetRoles, err := s.roleRepo.GetUserRoleCodes(id)
	if err != nil {
		return err
	}
	if model.ContainsAnyRole(targetRoles, model.RoleSuperAdmin) {
		return errors.New("超级管理员仅通过配置文件管理，不可删除")
	}
	if err := validateManageUserPermission(operatorRoles, targetRoles); err != nil {
		return err
	}
	return s.repo.Delete(id)
}

// ImpersonateUser 以指定用户身份生成 JWT（仅超级管理员可用）
func (s *UserService) ImpersonateUser(id uint) (string, *model.User, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return "", nil, errors.New("用户不存在")
	}
	characters, err := s.charRepo.ListByUserID(user.ID)
	if err != nil {
		return "", nil, err
	}
	if err := validateImpersonationTargetPrimaryCharacterHealth(*user, characters); err != nil {
		return "", nil, err
	}
	token, err := jwt.GenerateToken(user.ID, user.PrimaryCharacterID, user.Role, global.Config.JWT.ExpireDay)
	if err != nil {
		return "", nil, err
	}
	return token, user, nil
}

func (s *UserService) validateUniqueContacts(user *model.User) error {
	if user.QQ != "" {
		owner, err := s.repo.GetByQQ(user.QQ)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err == nil {
			if err := validateContactOwner(user.ID, owner, "QQ 号码"); err != nil {
				return err
			}
		}
	}

	if user.DiscordID != "" {
		owner, err := s.repo.GetByDiscordID(user.DiscordID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err == nil {
			if err := validateContactOwner(user.ID, owner, "Discord ID"); err != nil {
				return err
			}
		}
	}

	return nil
}

func buildUserPatchUpdates(current *model.User, patch UserPatch, requireComplete bool) (*model.User, map[string]any, error) {
	next := *current
	updates := map[string]any{}

	if patch.Nickname != nil {
		nickname := strings.TrimSpace(*patch.Nickname)
		if nickname == "" {
			return nil, nil, errors.New("昵称不能为空")
		}
		if utf8.RuneCountInString(nickname) > maxUserNicknameLength {
			return nil, nil, errors.New("昵称最多 20 个字符")
		}
		next.Nickname = nickname
		updates["nickname"] = nickname
	}

	if patch.QQ != nil {
		qq := strings.TrimSpace(*patch.QQ)
		if utf8.RuneCountInString(qq) > maxUserContactLength {
			return nil, nil, errors.New("QQ 号码最多 20 个字符")
		}
		if qq != "" && !isDigitsOnly(qq) {
			return nil, nil, errors.New("QQ 号码只能包含数字")
		}
		next.QQ = qq
		updates["qq"] = qq
	}

	if patch.DiscordID != nil {
		discordID := strings.TrimSpace(*patch.DiscordID)
		if utf8.RuneCountInString(discordID) > maxUserContactLength {
			return nil, nil, errors.New("discord ID 最多 20 个字符")
		}
		next.DiscordID = discordID
		updates["discord_id"] = discordID
	}

	if patch.Status != nil {
		next.Status = *patch.Status
		updates["status"] = *patch.Status
	}

	if requireComplete {
		if !next.HasNickname() {
			return nil, nil, errors.New("昵称不能为空")
		}
		if !next.HasRequiredContact() {
			return nil, nil, errors.New("请至少填写 QQ 号码或 Discord ID")
		}
	}

	return &next, updates, nil
}

func validateContactOwner(currentUserID uint, owner *model.User, label string) error {
	if owner != nil && owner.ID != currentUserID {
		return errors.New("该" + label + "已被其他用户使用")
	}
	return nil
}

func isDigitsOnly(value string) bool {
	for _, r := range value {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return value != ""
}

func validateManageUserPermission(operatorRoles, targetRoles []string) error {
	if model.IsSuperAdmin(operatorRoles) {
		return nil
	}
	if model.ContainsAnyRole(targetRoles, model.RoleSuperAdmin, model.RoleAdmin) {
		return errors.New("管理员不能编辑或删除超级管理员或其他管理员")
	}
	return nil
}

func validatePrimaryCharacterTokenHealth(user model.User, characters []model.EveCharacter) error {
	if user.PrimaryCharacterID == 0 {
		return nil
	}

	for _, character := range characters {
		if character.CharacterID == user.PrimaryCharacterID {
			if character.TokenInvalid {
				return errors.New("主人物 ESI 已过期，请重新授权后再登录")
			}
			return nil
		}
	}

	return nil
}

func validateImpersonationTargetPrimaryCharacterHealth(user model.User, characters []model.EveCharacter) error {
	if err := validatePrimaryCharacterTokenHealth(user, characters); err != nil {
		return errors.New("该用户主人物 ESI 已过期，无法模拟登录")
	}
	return nil
}

func (s *UserService) buildUserListItems(users []model.User) []model.UserListItem {
	if len(users) == 0 {
		return []model.UserListItem{}
	}

	userIDs := make([]uint, 0, len(users))
	for _, user := range users {
		userIDs = append(userIDs, user.ID)
	}

	roleCodesByUserID, err := s.roleRepo.GetUserRoleCodesByUserIDs(userIDs)
	if err != nil {
		roleCodesByUserID = map[uint][]string{}
	}

	userCharactersByUserID := map[uint][]model.UserListCharacter{}
	chars, err := s.charRepo.ListByUserIDs(userIDs)
	if err == nil {
		characterIDs := make([]int64, 0, len(chars))
		for _, char := range chars {
			characterIDs = append(characterIDs, char.CharacterID)
		}

		totalSPByCharacterID, err := s.skillRepo.GetSkillTotalsByCharacterIDs(characterIDs)
		if err != nil {
			totalSPByCharacterID = map[int64]int64{}
		}

		for _, char := range chars {
			userCharactersByUserID[char.UserID] = append(
				userCharactersByUserID[char.UserID],
				model.NewUserListCharacter(char, totalSPByCharacterID[char.CharacterID]),
			)
		}
	}

	items := make([]model.UserListItem, 0, len(users))
	for _, user := range users {
		userCharacters := append([]model.UserListCharacter(nil), userCharactersByUserID[user.ID]...)
		sort.Slice(userCharacters, func(i, j int) bool {
			if userCharacters[i].CharacterID == user.PrimaryCharacterID {
				return true
			}
			if userCharacters[j].CharacterID == user.PrimaryCharacterID {
				return false
			}
			if userCharacters[i].CharacterName != userCharacters[j].CharacterName {
				return userCharacters[i].CharacterName < userCharacters[j].CharacterName
			}
			return userCharacters[i].CharacterID < userCharacters[j].CharacterID
		})

		items = append(items, model.NewUserListItem(user, roleCodesByUserID[user.ID], userCharacters))
	}
	return items
}
