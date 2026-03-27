package service

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"amiya-eden/internal/utils"
	"amiya-eden/pkg/eve/esi"
	"context"
	"errors"

	"go.uber.org/zap"
)

// AutoRoleService ESI 自动权限映射服务
type AutoRoleService struct {
	autoRoleRepo *repository.AutoRoleRepository
	roleRepo     *repository.RoleRepository
	charRepo     *repository.EveCharacterRepository
	roleSvc      *RoleService
}

func NewAutoRoleService() *AutoRoleService {
	return &AutoRoleService{
		autoRoleRepo: repository.NewAutoRoleRepository(),
		roleRepo:     repository.NewRoleRepository(),
		charRepo:     repository.NewEveCharacterRepository(),
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
	fillRoleMappingNames(mappings)
	return mappings, nil
}

// CreateEsiRoleMapping 创建 ESI 角色映射
func (s *AutoRoleService) CreateEsiRoleMapping(esiRole string, roleCode string) (*model.EsiRoleMapping, error) {
	if !isValidEsiRole(esiRole) {
		return nil, errors.New("无效的 ESI 军团角色名")
	}
	if !model.IsValidRoleCode(roleCode) {
		return nil, errors.New("未知的系统角色编码")
	}
	if roleCode == model.RoleSuperAdmin {
		return nil, errors.New("不可映射到超级管理员")
	}

	mapping := &model.EsiRoleMapping{
		EsiRole:  esiRole,
		RoleCode: roleCode,
	}
	if err := s.autoRoleRepo.CreateEsiRoleMapping(mapping); err != nil {
		return nil, err
	}
	if def, ok := model.GetRoleDefinition(roleCode); ok {
		mapping.RoleName = def.Name
	}
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
	fillTitleMappingNames(mappings)
	return mappings, nil
}

// ListCorpTitles 获取数据库中所有去重的军团头衔（用于前端下拉选择）
func (s *AutoRoleService) ListCorpTitles() ([]repository.CorpTitleInfo, error) {
	titles, err := s.autoRoleRepo.ListDistinctCorpTitles()
	if err != nil {
		return nil, err
	}

	corpNames, err := s.fetchCorporationNames(titles)
	if err != nil {
		global.Logger.Warn("[AutoRole] Failed to fetch corporation names from ESI",
			zap.Error(err),
		)
	} else {
		for i := range titles {
			if name, ok := corpNames[titles[i].CorporationID]; ok {
				titles[i].CorporationName = name
			}
		}
	}

	return titles, nil
}

type esiNameEntry struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (s *AutoRoleService) fetchCorporationNames(titles []repository.CorpTitleInfo) (map[int64]string, error) {
	corpIDSet := make(map[int64]struct{})
	for _, t := range titles {
		if t.CorporationID > 0 {
			corpIDSet[t.CorporationID] = struct{}{}
		}
	}
	if len(corpIDSet) == 0 {
		return nil, nil
	}

	corpIDs := make([]int64, 0, len(corpIDSet))
	for id := range corpIDSet {
		corpIDs = append(corpIDs, id)
	}

	client := esi.NewClientWithConfig(global.Config.EveSSO.ESIBaseURL, global.Config.EveSSO.ESIAPIPrefix)
	var esiResults []esiNameEntry
	if err := client.PostJSON(
		context.Background(),
		"/universe/names?datasource=tranquility",
		"",
		corpIDs,
		&esiResults,
	); err != nil {
		return nil, err
	}

	nameMap := make(map[int64]string, len(esiResults))
	for _, entry := range esiResults {
		nameMap[int64(entry.ID)] = entry.Name
	}

	return nameMap, nil
}

// CreateEsiTitleMapping 创建 ESI 头衔映射
func (s *AutoRoleService) CreateEsiTitleMapping(corpID int64, titleID int, titleName string, roleCode string) (*model.EsiTitleMapping, error) {
	if !model.IsValidRoleCode(roleCode) {
		return nil, errors.New("未知的系统角色编码")
	}
	if roleCode == model.RoleSuperAdmin {
		return nil, errors.New("不可映射到超级管理员")
	}

	mapping := &model.EsiTitleMapping{
		CorporationID: corpID,
		TitleID:       titleID,
		TitleName:     titleName,
		RoleCode:      roleCode,
	}
	if err := s.autoRoleRepo.CreateEsiTitleMapping(mapping); err != nil {
		return nil, err
	}
	if def, ok := model.GetRoleDefinition(roleCode); ok {
		mapping.RoleName = def.Name
	}
	return mapping, nil
}

// DeleteEsiTitleMapping 删除 ESI 头衔映射
func (s *AutoRoleService) DeleteEsiTitleMapping(id uint) error {
	return s.autoRoleRepo.DeleteEsiTitleMapping(id)
}

// ─── 自动权限同步 ───

// SyncUserAutoRoles 根据 ESI 军团角色 + 头衔，自动同步用户的系统权限
func (s *AutoRoleService) SyncUserAutoRoles(ctx context.Context, userID uint) error {
	currentCodes, err := s.roleRepo.GetUserRoleCodes(userID)
	if err != nil {
		return err
	}
	if model.ContainsAnyRole(currentCodes, model.RoleSuperAdmin) {
		return nil
	}

	chars, err := s.charRepo.ListByUserID(userID)
	if err != nil {
		return err
	}
	if len(chars) == 0 {
		return nil
	}

	allowCorps := utils.GetAllowCorporations()
	allowCorpSet := make(map[int64]struct{}, len(allowCorps))
	for _, id := range allowCorps {
		allowCorpSet[id] = struct{}{}
	}

	autoRoleCodes := make(map[string]struct{})
	shouldPromoteGuestToUser := shouldAutoPromoteGuestToUser(currentCodes, chars, allowCorpSet)
	if shouldPromoteGuestToUser {
		autoRoleCodes[model.RoleUser] = struct{}{}
	}

	// Collect ESI corp roles from allowed corporations
	allEsiRoles := make(map[string]struct{})
	for _, char := range chars {
		if !isAllowedCorporation(char.CorporationID, allowCorpSet) {
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

	// ESI role mappings
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
				autoRoleCodes[m.RoleCode] = struct{}{}
			}
		}
	}

	// ESI title mappings
	for _, char := range chars {
		if !isAllowedCorporation(char.CorporationID, allowCorpSet) {
			continue
		}
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
			autoRoleCodes[m.RoleCode] = struct{}{}
		}
	}

	if hasDirectorCorpRole(allEsiRoles) {
		autoRoleCodes[model.RoleAdmin] = struct{}{}
	}

	// Merge: keep existing roles, add auto-mapped roles
	existingSet := make(map[string]struct{}, len(currentCodes))
	for _, code := range currentCodes {
		existingSet[code] = struct{}{}
	}

	var toAdd []string
	for code := range autoRoleCodes {
		if _, exists := existingSet[code]; !exists {
			toAdd = append(toAdd, code)
		}
	}

	if len(toAdd) > 0 {
		for _, code := range toAdd {
			if err := s.roleRepo.AddUserRole(userID, code); err != nil {
				global.Logger.Warn("[AutoRole] 添加自动角色失败",
					zap.Uint("user_id", userID),
					zap.String("role_code", code),
					zap.Error(err))
			}
		}
		if shouldPromoteGuestToUser {
			if err := s.roleRepo.RemoveUserRole(userID, model.RoleGuest); err != nil {
				global.Logger.Warn("[AutoRole] 移除 guest 角色失败",
					zap.Uint("user_id", userID),
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

func fillRoleMappingNames(mappings []model.EsiRoleMapping) {
	for i, m := range mappings {
		if def, ok := model.GetRoleDefinition(m.RoleCode); ok {
			mappings[i].RoleName = def.Name
		}
	}
}

func fillTitleMappingNames(mappings []model.EsiTitleMapping) {
	for i, m := range mappings {
		if def, ok := model.GetRoleDefinition(m.RoleCode); ok {
			mappings[i].RoleName = def.Name
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
