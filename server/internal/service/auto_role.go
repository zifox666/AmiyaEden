package service

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// AutoRoleService ESI / SeAT 自动权限映射服务
type AutoRoleService struct {
	autoRoleRepo *repository.AutoRoleRepository
	allowRepo    *repository.AllowedEntityRepository
	roleRepo     *repository.RoleRepository
	charRepo     *repository.EveCharacterRepository
	userRepo     *repository.UserRepository
	seatUserRepo *repository.SeatUserRepository
	configRepo   *repository.SysConfigRepository
	roleSvc      *RoleService
}

func NewAutoRoleService() *AutoRoleService {
	return &AutoRoleService{
		autoRoleRepo: repository.NewAutoRoleRepository(),
		allowRepo:    repository.NewAllowedEntityRepository(),
		roleRepo:     repository.NewRoleRepository(),
		charRepo:     repository.NewEveCharacterRepository(),
		userRepo:     repository.NewUserRepository(),
		seatUserRepo: repository.NewSeatUserRepository(),
		configRepo:   repository.NewSysConfigRepository(),
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
func (s *AutoRoleService) CreateEsiRoleMapping(esiRole string, roleID uint, onlyMainChar bool) (*model.EsiRoleMapping, error) {
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
		EsiRole:      esiRole,
		RoleID:       roleID,
		OnlyMainChar: onlyMainChar,
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
// 只返回在 auto_role 准入名单内的军团头衔（名单为空时不限制）
func (s *AutoRoleService) ListCorpTitles() ([]repository.CorpTitleInfo, error) {
	allowCorpIDs, err := s.allowRepo.GetCorporationIDs(model.AllowListAutoRole)
	if err != nil {
		return nil, err
	}
	return s.autoRoleRepo.ListDistinctCorpTitles(allowCorpIDs)
}

// ─── 准入名单管理 ───

// ListAllowedEntities 获取指定名单的所有实体
func (s *AutoRoleService) ListAllowedEntities(listType string) ([]model.AllowedEntity, error) {
	return s.allowRepo.List(listType)
}

// AddAllowedEntity 添加实体到名单
func (s *AutoRoleService) AddAllowedEntity(e *model.AllowedEntity) error {
	if e.ListType != model.AllowListAutoRole && e.ListType != model.AllowListBasicAccess {
		return errors.New("无效的名单类型")
	}
	if e.EntityType != model.AllowEntityTypeAlliance && e.EntityType != model.AllowEntityTypeCorporation {
		return errors.New("无效的实体类型")
	}
	if e.EntityID <= 0 {
		return errors.New("实体ID无效")
	}
	if e.EntityName == "" {
		return errors.New("实体名称不能为空")
	}
	return s.allowRepo.Add(e)
}

// RemoveAllowedEntity 从名单中删除实体
func (s *AutoRoleService) RemoveAllowedEntity(id uint) error {
	return s.allowRepo.Remove(id)
}

// ─── 准入名单"仅主角色"配置 ───

// AllowListOnlyMainCharConfig 准入名单"仅主角色"开关配置
type AllowListOnlyMainCharConfig struct {
	AutoRoleOnlyMainChar    bool `json:"auto_role_only_main_char"`
	BasicAccessOnlyMainChar bool `json:"basic_access_only_main_char"`
}

// GetAllowListOnlyMainCharConfig 读取两个准入名单的"仅主角色"开关（默认 false）
func (s *AutoRoleService) GetAllowListOnlyMainCharConfig() AllowListOnlyMainCharConfig {
	return AllowListOnlyMainCharConfig{
		AutoRoleOnlyMainChar:    s.configRepo.GetBool(model.SysConfigAutoRoleAllowOnlyMainChar, false),
		BasicAccessOnlyMainChar: s.configRepo.GetBool(model.SysConfigBasicAccessAllowOnlyMainChar, false),
	}
}

// SetAllowListOnlyMainCharConfig 更新两个准入名单的"仅主角色"开关
func (s *AutoRoleService) SetAllowListOnlyMainCharConfig(cfg AllowListOnlyMainCharConfig) error {
	autoVal := "false"
	if cfg.AutoRoleOnlyMainChar {
		autoVal = "true"
	}
	basicVal := "false"
	if cfg.BasicAccessOnlyMainChar {
		basicVal = "true"
	}
	if err := s.configRepo.Set(model.SysConfigAutoRoleAllowOnlyMainChar, autoVal, "自动权限准入仅检测主角色"); err != nil {
		return err
	}
	return s.configRepo.Set(model.SysConfigBasicAccessAllowOnlyMainChar, basicVal, "基础访问准入仅检测主角色")
}

// ─── zkillboard 实体搜索 ───

// ZkbSearchResult zkillboard 自动补全结果项
type ZkbSearchResult struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Type  string `json:"type"`
	Image string `json:"image"`
}

// SearchEveEntities 通过 zkillboard 模糊搜索 EVE 联盟/军团
func (s *AutoRoleService) SearchEveEntities(query string) ([]ZkbSearchResult, error) {
	if query == "" {
		return nil, errors.New("搜索词不能为空")
	}

	url := fmt.Sprintf("https://zkillboard.com/autocomplete/%s/", query)
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "AmiyaEden/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("zkillboard 搜索失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil {
		return nil, err
	}

	var all []ZkbSearchResult
	if err := json.Unmarshal(body, &all); err != nil {
		return nil, fmt.Errorf("解析搜索结果失败: %w", err)
	}

	// 只保留联盟和军团
	var filtered []ZkbSearchResult
	for _, item := range all {
		if item.Type == "alliance" || item.Type == "corporation" {
			filtered = append(filtered, item)
		}
	}
	return filtered, nil
}

// CreateEsiTitleMapping 创建 ESI 头衔映射
func (s *AutoRoleService) CreateEsiTitleMapping(corpID int64, titleID int, titleName string, roleID uint, onlyMainChar bool) (*model.EsiTitleMapping, error) {
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
		OnlyMainChar:  onlyMainChar,
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

// ─── SeAT Role Mapping CRUD ───

// ListSeatRoleMappings 获取所有 SeAT 分组映射（带角色信息）
func (s *AutoRoleService) ListSeatRoleMappings() ([]model.SeatRoleMapping, error) {
	mappings, err := s.autoRoleRepo.ListSeatRoleMappings()
	if err != nil {
		return nil, err
	}
	s.fillSeatRoleInfo(mappings)
	return mappings, nil
}

// CreateSeatRoleMapping 创建 SeAT 分组映射
func (s *AutoRoleService) CreateSeatRoleMapping(seatRole string, roleID uint) (*model.SeatRoleMapping, error) {
	if seatRole == "" {
		return nil, errors.New("SeAT 分组名不能为空")
	}
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		return nil, errors.New("系统角色不存在")
	}
	if role.Code == model.RoleSuperAdmin {
		return nil, errors.New("不可映射到超级管理员")
	}
	mapping := &model.SeatRoleMapping{
		SeatRole: seatRole,
		RoleID:   roleID,
	}
	if err := s.autoRoleRepo.CreateSeatRoleMapping(mapping); err != nil {
		return nil, err
	}
	mapping.RoleCode = role.Code
	mapping.RoleName = role.Name
	return mapping, nil
}

// DeleteSeatRoleMapping 删除 SeAT 分组映射
func (s *AutoRoleService) DeleteSeatRoleMapping(id uint) error {
	return s.autoRoleRepo.DeleteSeatRoleMapping(id)
}

// GetAllSeatRoles 获取数据库中所有已录入的 SeAT 分组名列表（供前端选择）
func (s *AutoRoleService) GetAllSeatRoles() ([]string, error) {
	return s.autoRoleRepo.ListDistinctSeatRoles()
}

// ─── 自动权限同步 ───

// SyncUserAutoRoles 根据 ESI 军团角色 + 头衔 + SeAT 分组，自动同步用户的系统权限
// 规则：
//   - Director 始终对应 admin 角色
//   - 根据 esi_role_mapping 表的配置，将 ESI 角色映射到系统角色
//   - 根据 esi_title_mapping 表的配置，将 ESI 头衔映射到系统角色
//   - 根据 seat_role_mapping 表的配置，将 SeAT 分组映射到系统角色
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

	// 获取用户绑定的所有角色
	chars, err := s.charRepo.ListByUserID(userID)
	if err != nil {
		return err
	}
	if len(chars) == 0 {
		return nil
	}

	// 获取用户主角色 ID 和昵称（primaryCharID 供 only_main_char 映射使用，username 供日志冗余）
	primaryCharID := int64(0)
	username := ""
	if u, err := s.userRepo.GetByID(userID); err == nil {
		primaryCharID = u.PrimaryCharacterID
		username = u.Nickname
	}

	// 从 DB 读取 auto_role 准入名单
	allowCorpIDs, allowAllianceIDs, err := s.allowRepo.GetAllIDs(model.AllowListAutoRole)
	if err != nil {
		global.Logger.Warn("[AutoRole] 读取准入名单失败", zap.Error(err))
	}
	allowCorpSet := make(map[int64]struct{}, len(allowCorpIDs))
	for _, id := range allowCorpIDs {
		allowCorpSet[id] = struct{}{}
	}
	allowAllianceSet := make(map[int64]struct{}, len(allowAllianceIDs))
	for _, id := range allowAllianceIDs {
		allowAllianceSet[id] = struct{}{}
	}
	allowFiltered := len(allowCorpSet)+len(allowAllianceSet) > 0

	// 若开启了"准入仅主角色"，只看主角色的军团/联盟是否在准入名单内
	if allowFiltered && s.configRepo.GetBool(model.SysConfigAutoRoleAllowOnlyMainChar, false) {
		primaryQualified := false
		for _, char := range chars {
			if char.CharacterID != primaryCharID {
				continue
			}
			if _, ok := allowCorpSet[char.CorporationID]; ok {
				primaryQualified = true
				break
			}
			if char.AllianceID != nil {
				if _, ok := allowAllianceSet[*char.AllianceID]; ok {
					primaryQualified = true
					break
				}
			}
			break
		}
		if !primaryQualified {
			return nil
		}
		// 主角色通过准入，后续不再逐角色过滤
		allowFiltered = false
	}

	// 收集所有角色的 ESI 军团角色（仅限允许军团/联盟）
	// allEsiRoles: 所有允许角色的军团职位集合
	// primaryEsiRoles: 仅主角色（primary_character_id）的军团职位集合，供 only_main_char 映射使用
	allEsiRoles := make(map[string]struct{})
	primaryEsiRoles := make(map[string]struct{})
	hasDirector := false

	for _, char := range chars {
		// 跳过不在允许名单中的角色
		if allowFiltered {
			inCorpList := false
			if _, ok := allowCorpSet[char.CorporationID]; ok {
				inCorpList = true
			}
			if !inCorpList && char.AllianceID != nil {
				if _, ok := allowAllianceSet[*char.AllianceID]; ok {
					inCorpList = true
				}
			}
			if !inCorpList {
				continue
			}
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
			if char.CharacterID == primaryCharID {
				primaryEsiRoles[r] = struct{}{}
			}
			if r == "Director" {
				hasDirector = true
			}
		}
	}

	// 根据所有 ESI 角色查找映射
	autoRoleIDs := make(map[uint]struct{})

	// Director → admin
	if hasDirector {
		adminRole, err := s.roleRepo.GetByCode(model.RoleAdmin)
		if err == nil {
			autoRoleIDs[adminRole.ID] = struct{}{}
		}
	}

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
				// 若该映射仅限主角色，则只有主角色拥有该职位时才匹配
				if m.OnlyMainChar {
					if _, ok := primaryEsiRoles[m.EsiRole]; !ok {
						continue
					}
				}
				autoRoleIDs[m.RoleID] = struct{}{}
			}
		}
	}

	// 查找 ESI 头衔映射（仅限允许军团/联盟）
	for _, char := range chars {
		if char.CorporationID == 0 {
			continue
		}
		// 跳过不在允许名单中的角色
		if allowFiltered {
			inList := false
			if _, ok := allowCorpSet[char.CorporationID]; ok {
				inList = true
			}
			if !inList && char.AllianceID != nil {
				if _, ok := allowAllianceSet[*char.AllianceID]; ok {
					inList = true
				}
			}
			if !inList {
				continue
			}
		}
		// 查询角色头衔
		var titles []model.EveCharacterTitle
		if err := global.DB.Where("character_id = ?", char.CharacterID).Find(&titles).Error; err != nil {
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
			// 若该映射仅限主角色，则只有主角色所在军团头衔匹配时才生效
			if m.OnlyMainChar && char.CharacterID != primaryCharID {
				continue
			}
			autoRoleIDs[m.RoleID] = struct{}{}
		}
	}

	// 查找 SeAT 分组映射
	seatUser, seatErr := s.seatUserRepo.GetByUserID(userID)
	if seatErr == nil && seatUser.Groups != "" {
		var seatGroups []string
		if jsonErr := json.Unmarshal([]byte(seatUser.Groups), &seatGroups); jsonErr == nil && len(seatGroups) > 0 {
			seatMappings, mapErr := s.autoRoleRepo.GetSeatRoleMappingsBySeatRoles(seatGroups)
			if mapErr != nil {
				global.Logger.Warn("[AutoRole] 查询 SeAT 分组映射失败", zap.Error(mapErr))
			} else {
				for _, m := range seatMappings {
					autoRoleIDs[m.RoleID] = struct{}{}
				}
			}
		}
	}

	// 获取用户所有当前角色 ID（用于判断是否需要新增）
	currentRoleIDs, err := s.roleRepo.GetUserRoleIDs(userID)
	if err != nil {
		return err
	}
	existingSet := make(map[uint]struct{}, len(currentRoleIDs))
	for _, id := range currentRoleIDs {
		existingSet[id] = struct{}{}
	}

	// 获取用户当前由自动系统分配的角色 ID（用于判断是否需要移除）
	currentAutoRoleIDs, err := s.roleRepo.GetUserAutoRoleIDs(userID)
	if err != nil {
		return err
	}
	currentAutoSet := make(map[uint]struct{}, len(currentAutoRoleIDs))
	for _, id := range currentAutoRoleIDs {
		currentAutoSet[id] = struct{}{}
	}

	// 计算需要新增的角色（当前没有的，无论手动还是自动）
	var toAdd []uint
	for rid := range autoRoleIDs {
		if _, exists := existingSet[rid]; !exists {
			toAdd = append(toAdd, rid)
		}
	}

	// 计算需要移除的角色（自动分配但不再符合条件的）
	var toRemove []uint
	for rid := range currentAutoSet {
		if _, shouldHave := autoRoleIDs[rid]; !shouldHave {
			toRemove = append(toRemove, rid)
		}
	}

	changed := false

	for _, rid := range toRemove {
		if err := s.roleRepo.RemoveUserRole(userID, rid); err != nil {
			global.Logger.Warn("[AutoRole] 移除过期自动角色失败",
				zap.Uint("user_id", userID),
				zap.Uint("role_id", rid),
				zap.Error(err))
		} else {
			changed = true
			s.writeLog(userID, username, rid, "remove")
		}
	}

	for _, rid := range toAdd {
		if err := s.roleRepo.AddAutoUserRole(userID, rid); err != nil {
			global.Logger.Warn("[AutoRole] 添加自动角色失败",
				zap.Uint("user_id", userID),
				zap.Uint("role_id", rid),
				zap.Error(err))
		} else {
			changed = true
			s.writeLog(userID, username, rid, "add")
		}
	}

	if changed {
		s.roleSvc.InvalidateUserCache(ctx, userID)
		s.roleSvc.SyncUserPrimaryRole(userID)
		global.Logger.Info("[AutoRole] 用户自动角色已更新",
			zap.Uint("user_id", userID),
			zap.Int("added", len(toAdd)),
			zap.Int("removed", len(toRemove)))
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

// SyncAllUsersBasicAccess 同步所有用户的基础准入权限（basic_access 名单）
// 与 roleCheckTask 逻辑相同，供手动触发调用
func (s *AutoRoleService) SyncAllUsersBasicAccess(ctx context.Context) {
	nonEmpty, _ := s.allowRepo.IsNonEmpty(model.AllowListBasicAccess)
	if !nonEmpty {
		return
	}

	ids, err := s.userRepo.ListAllIDs()
	if err != nil {
		global.Logger.Error("[AutoRole] 查询用户 ID 列表失败（basic_access 同步）", zap.Error(err))
		return
	}

	global.Logger.Info("[AutoRole] 开始基础准入同步", zap.Int("users", len(ids)))
	for _, uid := range ids {
		if err := s.roleSvc.CheckCorpAccessAndAdjustRole(ctx, uid); err != nil {
			global.Logger.Warn("[AutoRole] basic_access 同步失败",
				zap.Uint("user_id", uid),
				zap.Error(err))
		}
	}
	global.Logger.Info("[AutoRole] 基础准入同步完成")
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

func (s *AutoRoleService) fillSeatRoleInfo(mappings []model.SeatRoleMapping) {
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

// writeLog 写入一条自动权限操作日志（失败仅打 warn，不影响主流程）
func (s *AutoRoleService) writeLog(userID uint, username string, roleID uint, action string) {
	roleName, roleCode := "", ""
	if role, err := s.roleRepo.GetByID(roleID); err == nil {
		roleName = role.Name
		roleCode = role.Code
	}
	log := &model.AutoRoleLog{
		UserID:   userID,
		Username: username,
		RoleID:   roleID,
		RoleName: roleName,
		RoleCode: roleCode,
		Action:   action,
	}
	if err := s.autoRoleRepo.CreateAutoRoleLog(log); err != nil {
		global.Logger.Warn("[AutoRole] 写入日志失败", zap.Error(err))
	}
}

// ListAutoRoleLogs 分页查询自动权限操作日志
func (s *AutoRoleService) ListAutoRoleLogs(page, pageSize int) ([]model.AutoRoleLog, int64, error) {
	return s.autoRoleRepo.ListAutoRoleLogs(page, pageSize)
}
