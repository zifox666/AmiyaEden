package service

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"context"
	"errors"

	"go.uber.org/zap"
)

// AutoRoleService ESI 自动权限映射服务
type AutoRoleService struct {
	autoRoleRepo *repository.AutoRoleRepository
	roleRepo     *repository.RoleRepository
	charRepo     *repository.EveCharacterRepository
	userRepo     *repository.UserRepository
	roleSvc      *RoleService
}

func NewAutoRoleService() *AutoRoleService {
	return &AutoRoleService{
		autoRoleRepo: repository.NewAutoRoleRepository(),
		roleRepo:     repository.NewRoleRepository(),
		charRepo:     repository.NewEveCharacterRepository(),
		userRepo:     repository.NewUserRepository(),
		roleSvc:      NewRoleService(),
	}
}

// ─── ESI Role Mapping CRUD ───

// ListEsiRoleMappings 获取所有 ESI 角色映射（带角色信息）
func (s *AutoRoleService) ListEsiRoleMappings() ([]model.EsiRoleMapping, error) {
	mappings, err := s.autoRoleRepo.ListEsiRoleMappings()
	if err != nil {
		return nil, err
	}
	s.fillRoleInfo(mappings)
	return mappings, nil
}

// CreateEsiRoleMapping 创建 ESI 角色映射
func (s *AutoRoleService) CreateEsiRoleMapping(esiRole string, roleID uint) (*model.EsiRoleMapping, error) {
	// 验证 ESI 角色名合法性
	if !isValidEsiRole(esiRole) {
		return nil, errors.New("无效的 ESI 军团角色名")
	}
	// 验证系统角色存在
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		return nil, errors.New("系统角色不存在")
	}
	// 不允许映射到 super_admin
	if role.Code == model.RoleSuperAdmin {
		return nil, errors.New("不可映射到超级管理员")
	}

	mapping := &model.EsiRoleMapping{
		EsiRole: esiRole,
		RoleID:  roleID,
	}
	if err := s.autoRoleRepo.CreateEsiRoleMapping(mapping); err != nil {
		return nil, err
	}
	mapping.RoleCode = role.Code
	mapping.RoleName = role.Name
	return mapping, nil
}

// DeleteEsiRoleMapping 删除 ESI 角色映射
func (s *AutoRoleService) DeleteEsiRoleMapping(id uint) error {
	return s.autoRoleRepo.DeleteEsiRoleMapping(id)
}

// GetAllEsiRoles 获取所有 ESI 军团角色名列表（供前端选择）
func (s *AutoRoleService) GetAllEsiRoles() []string {
	return model.AllEsiCorpRoles
}

// ─── ESI Title Mapping CRUD ───

// ListEsiTitleMappings 获取所有 ESI 头衔映射（带角色信息）
func (s *AutoRoleService) ListEsiTitleMappings() ([]model.EsiTitleMapping, error) {
	mappings, err := s.autoRoleRepo.ListEsiTitleMappings()
	if err != nil {
		return nil, err
	}
	s.fillTitleRoleInfo(mappings)
	return mappings, nil
}

// ListCorpTitles 获取数据库中所有去重的军团头衔（用于前端下拉选择）
func (s *AutoRoleService) ListCorpTitles() ([]repository.CorpTitleInfo, error) {
	return s.autoRoleRepo.ListDistinctCorpTitles()
}

// CreateEsiTitleMapping 创建 ESI 头衔映射
func (s *AutoRoleService) CreateEsiTitleMapping(corpID int64, titleID int, titleName string, roleID uint) (*model.EsiTitleMapping, error) {
	// 验证系统角色存在
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		return nil, errors.New("系统角色不存在")
	}
	// 不允许映射到 super_admin
	if role.Code == model.RoleSuperAdmin {
		return nil, errors.New("不可映射到超级管理员")
	}

	mapping := &model.EsiTitleMapping{
		CorporationID: corpID,
		TitleID:       titleID,
		TitleName:     titleName,
		RoleID:        roleID,
	}
	if err := s.autoRoleRepo.CreateEsiTitleMapping(mapping); err != nil {
		return nil, err
	}
	mapping.RoleCode = role.Code
	mapping.RoleName = role.Name
	return mapping, nil
}

// DeleteEsiTitleMapping 删除 ESI 头衔映射
func (s *AutoRoleService) DeleteEsiTitleMapping(id uint) error {
	return s.autoRoleRepo.DeleteEsiTitleMapping(id)
}

// ─── 自动权限同步 ───

// SyncUserAutoRoles 根据 ESI 军团角色 + 头衔，自动同步用户的系统权限
// 规则：
//   - 主角色在 allow_corporations 内时自动补 user 角色
//   - 任一允许军团角色拥有 ESI corp role `Director` 时自动补 admin 角色
//   - 根据 esi_role_mapping 表的配置，将 ESI 角色映射到系统角色
//   - 根据 esi_title_mapping 表的配置，将 ESI 头衔映射到系统角色
//   - super_admin 不受影响
//   - 保留用户手动分配的角色，仅补充自动映射的角色
func (s *AutoRoleService) SyncUserAutoRoles(ctx context.Context, userID uint) error {
	// admin / super_admin 不受自动权限影响
	currentCodes, err := s.roleRepo.GetUserRoleCodes(userID)
	if err != nil {
		return err
	}
	if model.ContainsAnyRole(currentCodes, model.RoleSuperAdmin) {
		return nil
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	// 获取用户绑定的所有角色
	chars, err := s.charRepo.ListByUserID(userID)
	if err != nil {
		return err
	}
	if len(chars) == 0 {
		return nil
	}

	// 构建允许军团白名单（为空表示不限制）
	allowCorps := global.Config.App.AllowCorporations
	allowCorpSet := make(map[int64]struct{}, len(allowCorps))
	for _, id := range allowCorps {
		allowCorpSet[id] = struct{}{}
	}

	autoRoleIDs := make(map[uint]struct{})
	if hasAllowedPrimaryCharacter(user.PrimaryCharacterID, chars, allowCorpSet) {
		userRole, err := s.roleRepo.GetByCode(model.RoleUser)
		if err != nil {
			global.Logger.Warn("[AutoRole] 查询 user 角色失败", zap.Error(err))
		} else {
			autoRoleIDs[userRole.ID] = struct{}{}
		}
	}

	// 收集所有角色的 ESI 军团角色。
	// 当配置了 allow_corporations 时，仅允许名单内军团参与自动映射；
	// 基于主角色/Director corp role 的内置快捷规则则始终要求命中 allow_corporations。
	allEsiRoles := make(map[string]struct{})

	for _, char := range chars {
		inAllowedCorp := true
		if len(allowCorpSet) > 0 {
			_, inAllowedCorp = allowCorpSet[char.CorporationID]
		}
		if !inAllowedCorp {
			continue
		}

		corpRoles, err := s.autoRoleRepo.ListCharacterCorpRoles(char.CharacterID)
		if err != nil {
			global.Logger.Warn("[AutoRole] 查询角色军团角色失败",
				zap.Int64("character_id", char.CharacterID),
				zap.Error(err))
			continue
		}
		for _, r := range corpRoles {
			allEsiRoles[r] = struct{}{}
		}
	}

	// 根据所有 ESI 角色查找映射
	// 查找 ESI 角色映射
	if len(allEsiRoles) > 0 {
		esiRoleNames := make([]string, 0, len(allEsiRoles))
		for r := range allEsiRoles {
			esiRoleNames = append(esiRoleNames, r)
		}
		mappings, err := s.autoRoleRepo.GetEsiRoleMappingsByEsiRoles(esiRoleNames)
		if err != nil {
			global.Logger.Warn("[AutoRole] 查询 ESI 角色映射失败", zap.Error(err))
		} else {
			for _, m := range mappings {
				autoRoleIDs[m.RoleID] = struct{}{}
			}
		}
	}

	// 查找 ESI 头衔映射（仅限允许军团）
	for _, char := range chars {
		if char.CorporationID == 0 {
			continue
		}
		// 跳过不在允许军团中的角色
		if len(allowCorpSet) > 0 {
			if _, ok := allowCorpSet[char.CorporationID]; !ok {
				continue
			}
		}
		// 查询角色头衔
		titles, err := s.autoRoleRepo.ListCharacterTitles(char.CharacterID)
		if err != nil {
			global.Logger.Warn("[AutoRole] 查询角色头衔失败",
				zap.Int64("character_id", char.CharacterID),
				zap.Error(err))
			continue
		}
		if len(titles) == 0 {
			continue
		}
		titleIDs := make([]int, 0, len(titles))
		for _, t := range titles {
			titleIDs = append(titleIDs, t.TitleID)
		}
		titleMappings, err := s.autoRoleRepo.GetEsiTitleMappingsByCorpAndTitles(char.CorporationID, titleIDs)
		if err != nil {
			continue
		}
		for _, m := range titleMappings {
			autoRoleIDs[m.RoleID] = struct{}{}
		}
	}

	if hasDirectorCorpRole(allEsiRoles) {
		adminRole, err := s.roleRepo.GetByCode(model.RoleAdmin)
		if err != nil {
			global.Logger.Warn("[AutoRole] 查询 admin 角色失败", zap.Error(err))
		} else {
			autoRoleIDs[adminRole.ID] = struct{}{}
		}
	}

	// 获取用户当前角色 ID
	currentRoleIDs, err := s.roleRepo.GetUserRoleIDs(userID)
	if err != nil {
		return err
	}

	// 合并：保留现有角色，补充自动映射的角色
	existingSet := make(map[uint]struct{}, len(currentRoleIDs))
	for _, id := range currentRoleIDs {
		existingSet[id] = struct{}{}
	}

	var toAdd []uint
	for rid := range autoRoleIDs {
		if _, exists := existingSet[rid]; !exists {
			toAdd = append(toAdd, rid)
		}
	}

	if len(toAdd) > 0 {
		for _, rid := range toAdd {
			if err := s.roleRepo.AddUserRole(userID, rid); err != nil {
				global.Logger.Warn("[AutoRole] 添加自动角色失败",
					zap.Uint("user_id", userID),
					zap.Uint("role_id", rid),
					zap.Error(err))
			}
		}
		s.roleSvc.InvalidateUserCache(ctx, userID)
		s.roleSvc.SyncUserPrimaryRole(userID)
		global.Logger.Info("[AutoRole] 用户自动角色已更新",
			zap.Uint("user_id", userID),
			zap.Int("added", len(toAdd)))
	}

	return nil
}

// SyncAllUsersAutoRoles 同步所有用户的自动权限（供定时任务调用）
func (s *AutoRoleService) SyncAllUsersAutoRoles(ctx context.Context) {
	userRepo := repository.NewUserRepository()
	ids, err := userRepo.ListAllIDs()
	if err != nil {
		global.Logger.Error("[AutoRole] 查询用户 ID 列表失败", zap.Error(err))
		return
	}

	global.Logger.Info("[AutoRole] 开始自动权限同步", zap.Int("users", len(ids)))
	for _, uid := range ids {
		if err := s.SyncUserAutoRoles(ctx, uid); err != nil {
			global.Logger.Warn("[AutoRole] 同步失败",
				zap.Uint("user_id", uid),
				zap.Error(err))
		}
	}
	global.Logger.Info("[AutoRole] 自动权限同步完成")
}

// ─── 内部辅助 ───

func (s *AutoRoleService) fillRoleInfo(mappings []model.EsiRoleMapping) {
	for i, m := range mappings {
		role, err := s.roleRepo.GetByID(m.RoleID)
		if err == nil {
			mappings[i].RoleCode = role.Code
			mappings[i].RoleName = role.Name
		}
	}
}

func (s *AutoRoleService) fillTitleRoleInfo(mappings []model.EsiTitleMapping) {
	for i, m := range mappings {
		role, err := s.roleRepo.GetByID(m.RoleID)
		if err == nil {
			mappings[i].RoleCode = role.Code
			mappings[i].RoleName = role.Name
		}
	}
}

func isValidEsiRole(name string) bool {
	for _, r := range model.AllEsiCorpRoles {
		if r == name {
			return true
		}
	}
	return false
}
