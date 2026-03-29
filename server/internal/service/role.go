package service

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type RoleService struct {
	repo     *repository.RoleRepository
	menuRepo *repository.MenuRepository
	userRepo *repository.UserRepository
}

func NewRoleService() *RoleService {
	return &RoleService{
		repo:     repository.NewRoleRepository(),
		menuRepo: repository.NewMenuRepository(),
		userRepo: repository.NewUserRepository(),
	}
}

// ─── Redis 缓存 ───

const (
	userRolesCachePrefix = "user_roles:"
	userPermsCachePrefix = "user_perms:"
	cacheTTL             = 30 * time.Minute
)

// GetUserRoleNames 获取用户角色编码列表（带缓存）
func (s *RoleService) GetUserRoleNames(ctx context.Context, userID uint) ([]string, error) {
	cacheKey := fmt.Sprintf("%s%d", userRolesCachePrefix, userID)
	val, err := global.Redis.Get(ctx, cacheKey).Result()
	if err == nil && val != "" {
		var roles []string
		if json.Unmarshal([]byte(val), &roles) == nil {
			return roles, nil
		}
	}

	roles, err := s.repo.GetUserRoleCodes(userID)
	if err != nil {
		return nil, err
	}
	if len(roles) == 0 {
		roles = []string{model.RoleGuest}
	}

	if data, err := json.Marshal(roles); err == nil {
		global.Redis.Set(ctx, cacheKey, string(data), cacheTTL)
	}
	return roles, nil
}

// GetUserPermissions 获取用户所有权限标识列表（带缓存）
func (s *RoleService) GetUserPermissions(ctx context.Context, userID uint) ([]string, error) {
	cacheKey := fmt.Sprintf("%s%d", userPermsCachePrefix, userID)
	val, err := global.Redis.Get(ctx, cacheKey).Result()
	if err == nil && val != "" {
		var perms []string
		if json.Unmarshal([]byte(val), &perms) == nil {
			return perms, nil
		}
	}

	// 获取用户角色
	roleCodes, err := s.repo.GetUserRoleCodes(userID)
	if err != nil {
		return nil, err
	}

	// super_admin 拥有所有权限
	if model.IsSuperAdmin(roleCodes) {
		allMenus, err := s.menuRepo.ListAll()
		if err != nil {
			return nil, err
		}
		perms := make([]string, 0)
		for _, m := range allMenus {
			if m.Type == model.MenuTypeButton && m.Permission != "" {
				perms = append(perms, m.Permission)
			}
		}
		if data, err := json.Marshal(perms); err == nil {
			global.Redis.Set(ctx, cacheKey, string(data), cacheTTL)
		}
		return perms, nil
	}

	// 获取角色ID
	roleIDs, err := s.repo.GetUserRoleIDs(userID)
	if err != nil {
		return nil, err
	}
	if len(roleIDs) == 0 {
		return []string{}, nil
	}

	// 获取角色所有菜单ID
	menuIDs, err := s.repo.GetMenuIDsByRoles(roleIDs)
	if err != nil {
		return nil, err
	}
	if len(menuIDs) == 0 {
		return []string{}, nil
	}

	// 获取菜单，过滤出按钮权限
	menus, err := s.menuRepo.ListByIDs(menuIDs)
	if err != nil {
		return nil, err
	}

	perms := make([]string, 0)
	for _, m := range menus {
		if m.Type == model.MenuTypeButton && m.Permission != "" {
			perms = append(perms, m.Permission)
		}
	}

	if data, err := json.Marshal(perms); err == nil {
		global.Redis.Set(ctx, cacheKey, string(data), cacheTTL)
	}
	return perms, nil
}

// InvalidateUserCache 清除用户角色和权限缓存
func (s *RoleService) InvalidateUserCache(ctx context.Context, userID uint) {
	global.Redis.Del(ctx, fmt.Sprintf("%s%d", userRolesCachePrefix, userID))
	global.Redis.Del(ctx, fmt.Sprintf("%s%d", userPermsCachePrefix, userID))
}

// InvalidateUserRolesCache 兼容旧接口
func (s *RoleService) InvalidateUserRolesCache(ctx context.Context, userID uint) {
	s.InvalidateUserCache(ctx, userID)
}

// ─── 角色 CRUD ───

func (s *RoleService) ListRoles(page, pageSize int) ([]model.Role, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.List(page, pageSize)
}

func (s *RoleService) ListAllRoles() ([]model.Role, error) {
	return s.repo.ListAll()
}

func (s *RoleService) GetRole(id uint) (*model.Role, error) {
	role, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	menuIDs, _ := s.repo.GetRoleMenuIDs(id)
	role.MenuIDs = menuIDs
	return role, nil
}

func (s *RoleService) CreateRole(role *model.Role) error {
	if role.Code == "" {
		return errors.New("角色编码不能为空")
	}
	if role.Name == "" {
		return errors.New("角色名称不能为空")
	}
	role.Status = 1
	return s.repo.Create(role)
}

func (s *RoleService) UpdateRole(role *model.Role) error {
	existing, err := s.repo.GetByID(role.ID)
	if err != nil {
		return errors.New("角色不存在")
	}
	if existing.IsSystem {
		// 系统角色只允许修改名称和描述
		existing.Name = role.Name
		existing.Description = role.Description
		return s.repo.Update(existing)
	}
	role.IsSystem = false
	return s.repo.Update(role)
}

func (s *RoleService) DeleteRole(id uint) error {
	role, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("角色不存在")
	}
	if role.IsSystem {
		return errors.New("系统内置角色不可删除")
	}
	return s.repo.Delete(id)
}

// ─── 角色权限（菜单）管理 ───

func (s *RoleService) GetRoleMenuIDs(roleID uint) ([]uint, error) {
	return s.repo.GetRoleMenuIDs(roleID)
}

func (s *RoleService) SetRoleMenus(ctx context.Context, roleID uint, menuIDs []uint) error {
	_, err := s.repo.GetByID(roleID)
	if err != nil {
		return errors.New("角色不存在")
	}
	if err := s.repo.SetRoleMenus(roleID, menuIDs); err != nil {
		return err
	}
	// 清除该角色所有用户的缓存
	userIDs, _ := s.repo.GetRoleUserIDs(roleID)
	for _, uid := range userIDs {
		s.InvalidateUserCache(ctx, uid)
	}
	return nil
}

// ─── 用户角色管理 ───

func (s *RoleService) GetUserRoles(userID uint) ([]model.Role, error) {
	roleIDs, err := s.repo.GetUserRoleIDs(userID)
	if err != nil {
		return nil, err
	}
	if len(roleIDs) == 0 {
		return []model.Role{}, nil
	}
	var roles []model.Role
	err = global.DB.Where("id IN ?", roleIDs).Order("sort DESC, id ASC").Find(&roles).Error
	if roles == nil {
		roles = []model.Role{}
	}
	return roles, err
}

func (s *RoleService) SetUserRoles(ctx context.Context, operatorRoles []string, userID uint, roleIDs []uint) error {
	// 检查是否包含 super_admin 角色
	for _, rid := range roleIDs {
		role, err := s.repo.GetByID(rid)
		if err != nil {
			return fmt.Errorf("角色ID %d 不存在", rid)
		}
		if role.Code == model.RoleSuperAdmin && !model.IsSuperAdmin(operatorRoles) {
			return errors.New("只有超级管理员可以分配该角色")
		}
	}

	if err := s.repo.SetUserRoles(userID, roleIDs); err != nil {
		return err
	}

	// 同步 User.Role 字段（取最高优先级角色）
	s.SyncUserPrimaryRole(userID)
	s.InvalidateUserCache(ctx, userID)
	return nil
}

// ─── 内部辅助 ───

func (s *RoleService) SyncUserPrimaryRole(userID uint) {
	codes, err := s.repo.GetUserRoleCodes(userID)
	if err != nil || len(codes) == 0 {
		_ = s.userRepo.UpdateRole(userID, model.RoleUser)
		return
	}
	// 取第一个（已按 sort 排序）
	_ = s.userRepo.UpdateRole(userID, codes[0])
}

// SeedSystemRoles 初始化系统角色种子数据
func (s *RoleService) SeedSystemRoles() {
	// 清理旧版 code 为空的脏数据
	if err := global.DB.Where("code = '' OR code IS NULL").Delete(&model.Role{}).Error; err != nil {
		global.Logger.Warn("清理旧角色数据失败", zap.Error(err))
	}

	for _, seed := range model.SystemRoleSeeds {
		role := seed
		if err := s.repo.UpsertSystemRole(&role); err != nil {
			global.Logger.Error("种子角色同步失败", zap.String("role", seed.Code), zap.Error(err))
		}
	}
	global.Logger.Info("系统角色种子同步完成")
}

// SeedSystemMenus 初始化系统菜单种子数据
func (s *RoleService) SeedSystemMenus() {
	seeds := model.GetSystemMenuSeeds()
	nameToID := make(map[string]uint)

	// 先处理根菜单，再处理子菜单
	for pass := 0; pass < 5; pass++ {
		for _, seed := range seeds {
			// 确定父ID
			parentID := uint(0)
			if seed.ParentName != "" {
				pid, ok := nameToID[seed.ParentName]
				if !ok {
					continue // 父菜单未创建，等下一轮
				}
				parentID = pid
			}

			// 已处理过的跳过
			if _, exists := nameToID[seed.Menu.Name]; exists {
				continue
			}

			menu := seed.Menu
			menu.ParentID = parentID
			if err := s.menuRepo.UpsertByName(&menu); err != nil {
				global.Logger.Error("种子菜单同步失败", zap.String("name", seed.Menu.Name), zap.Error(err))
				continue
			}

			// 获取真实ID
			created, err := s.menuRepo.GetByName(seed.Menu.Name)
			if err != nil {
				global.Logger.Error("查询种子菜单失败", zap.String("name", seed.Menu.Name), zap.Error(err))
				continue
			}
			nameToID[seed.Menu.Name] = created.ID
		}
	}

	global.Logger.Info("系统菜单种子同步完成", zap.Int("count", len(nameToID)))

	// 设置默认角色-菜单映射
	s.seedDefaultRoleMenus(nameToID)
}

func (s *RoleService) seedDefaultRoleMenus(nameToID map[string]uint) {
	roleMenuMap := model.DefaultRoleMenuMap()

	for roleCode, menuNames := range roleMenuMap {
		role, err := s.repo.GetByCode(roleCode)
		if err != nil {
			global.Logger.Warn("默认角色未找到", zap.String("code", roleCode))
			continue
		}

		// admin 角色自动获取所有菜单权限
		if roleCode == model.RoleAdmin {
			var allMenuIDs []uint
			for _, id := range nameToID {
				allMenuIDs = append(allMenuIDs, id)
			}
			if len(allMenuIDs) > 0 {
				existing, _ := s.repo.GetRoleMenuIDs(role.ID)
				existSet := make(map[uint]struct{}, len(existing))
				for _, id := range existing {
					existSet[id] = struct{}{}
				}
				var toAdd []uint
				for _, id := range allMenuIDs {
					if _, ok := existSet[id]; !ok {
						toAdd = append(toAdd, id)
					}
				}
				if len(toAdd) > 0 {
					merged := append(existing, toAdd...)
					if err := s.repo.SetRoleMenus(role.ID, merged); err != nil {
						global.Logger.Error("设置管理员全部菜单失败", zap.String("role", roleCode), zap.Error(err))
					} else {
						global.Logger.Info("管理员菜单已增量更新", zap.String("role", roleCode), zap.Int("added", len(toAdd)))
					}
				}
			}
			continue
		}

		// 计算 seed 中该角色应有的菜单 ID 集合
		var seedMenuIDs []uint
		for _, name := range menuNames {
			if id, ok := nameToID[name]; ok {
				seedMenuIDs = append(seedMenuIDs, id)
			}
		}
		if len(seedMenuIDs) == 0 {
			continue
		}

		existing, _ := s.repo.GetRoleMenuIDs(role.ID)

		if len(existing) == 0 {
			// 角色尚无菜单，直接写入
			if err := s.repo.SetRoleMenus(role.ID, seedMenuIDs); err != nil {
				global.Logger.Error("设置默认角色菜单失败", zap.String("role", roleCode), zap.Error(err))
			}
			continue
		}

		// 角色已有菜单配置：增量补入 seed 中新增但尚未分配的菜单
		existSet := make(map[uint]struct{}, len(existing))
		for _, id := range existing {
			existSet[id] = struct{}{}
		}
		var toAdd []uint
		for _, id := range seedMenuIDs {
			if _, ok := existSet[id]; !ok {
				toAdd = append(toAdd, id)
			}
		}
		if len(toAdd) > 0 {
			merged := append(existing, toAdd...)
			if err := s.repo.SetRoleMenus(role.ID, merged); err != nil {
				global.Logger.Error("增量更新角色菜单失败", zap.String("role", roleCode), zap.Error(err))
			} else {
				global.Logger.Info("角色菜单已增量更新", zap.String("role", roleCode), zap.Int("added", len(toAdd)))
			}
		}
	}
	global.Logger.Info("默认角色菜单映射完成")
}

// CheckCorpAccessAndAdjustRole 检查用户名下所有角色的军团/联盟归属是否在准入列表内
// 规则：
//   - basic_access 名单为空 → 不限制，直接返回
//   - admin / super_admin → 不受影响
//   - 至少有一个角色的 CorporationID 或 AllianceID 在允许列表内 → 确保拥有 user 角色（从 guest 升级）
//   - 没有符合条件的角色 → 降级为 guest（清除所有非高级角色）
func (s *RoleService) CheckCorpAccessAndAdjustRole(ctx context.Context, userID uint) error {
	allowRepo := repository.NewAllowedEntityRepository()
	allowCorpIDs, allowAllianceIDs, err := allowRepo.GetAllIDs(model.AllowListBasicAccess)
	if err != nil {
		return err
	}
	if len(allowCorpIDs)+len(allowAllianceIDs) == 0 {
		return nil
	}

	allowCorpSet := make(map[int64]struct{}, len(allowCorpIDs))
	for _, id := range allowCorpIDs {
		allowCorpSet[id] = struct{}{}
	}
	allowAllianceSet := make(map[int64]struct{}, len(allowAllianceIDs))
	for _, id := range allowAllianceIDs {
		allowAllianceSet[id] = struct{}{}
	}

	// 查询该用户绑定的所有 EVE 角色
	charRepo := repository.NewEveCharacterRepository()
	chars, err := charRepo.ListByUserID(userID)
	if err != nil {
		return err
	}

	// 检查是否有角色属于允许军团或允许联盟
	hasAccess := false
	for _, c := range chars {
		if c.CorporationID != 0 {
			if _, ok := allowCorpSet[c.CorporationID]; ok {
				hasAccess = true
				break
			}
		}
		if c.AllianceID != nil && *c.AllianceID != 0 {
			if _, ok := allowAllianceSet[*c.AllianceID]; ok {
				hasAccess = true
				break
			}
		}
	}

	// 获取用户当前拥有的角色
	rollCodes, err := s.repo.GetUserRoleCodes(userID)
	if err != nil {
		return err
	}

	// admin / super_admin 不受军团限制影响
	if model.ContainsAnyRole(rollCodes, model.RoleAdmin, model.RoleSuperAdmin) {
		return nil
	}

	if hasAccess {
		// 已有 user 或更高普通权限则无需变更
		if model.ContainsRole(rollCodes, model.RoleUser) {
			return nil
		}
		// 从 guest 升级为 user：先移除 guest，再添加 user
		userRole, err := s.repo.GetByCode(model.RoleUser)
		if err != nil {
			return err
		}
		if guestRole, err := s.repo.GetByCode(model.RoleGuest); err == nil {
			_ = s.repo.RemoveUserRole(userID, guestRole.ID)
		}
		if err := s.repo.AddUserRole(userID, userRole.ID); err != nil {
			return err
		}
		s.InvalidateUserCache(ctx, userID)
		global.Logger.Info("[CorpCheck] 用户升级为 user",
			zap.Uint("user_id", userID))
		// 写入同步日志
		s.writeBasicAccessLog(userID, userRole.ID, userRole.Name, model.RoleUser, "add")
	} else {
		// 已经是纯 guest 则无需变更
		if len(rollCodes) == 1 && rollCodes[0] == model.RoleGuest {
			return nil
		}
		// 清除所有角色，降级为 guest
		guestRole, err := s.repo.GetByCode(model.RoleGuest)
		if err != nil {
			return err
		}
		roleIDs, _ := s.repo.GetUserRoleIDs(userID)
		for _, rid := range roleIDs {
			_ = s.repo.RemoveUserRole(userID, rid)
		}
		if err := s.repo.AddUserRole(userID, guestRole.ID); err != nil {
			return err
		}
		s.InvalidateUserCache(ctx, userID)
		global.Logger.Info("[CorpCheck] 用户降级为 guest",
			zap.Uint("user_id", userID))
		// 写入同步日志（记录 user 角色被移除）
		if userRole, err := s.repo.GetByCode(model.RoleUser); err == nil {
			s.writeBasicAccessLog(userID, userRole.ID, userRole.Name, model.RoleUser, "remove")
		}
	}
	return nil
}

// EnsureUserHasDefaultRole 确保用户拥有默认角色
func (s *RoleService) EnsureUserHasDefaultRole(ctx context.Context, userID uint) {
	codes, err := s.repo.GetUserRoleCodes(userID)
	if err != nil || len(codes) == 0 {
		role, err := s.repo.GetByCode(model.RoleGuest)
		if err != nil {
			global.Logger.Error("查找默认角色失败", zap.Error(err))
			return
		}
		if err := s.repo.AddUserRole(userID, role.ID); err != nil {
			global.Logger.Error("分配默认角色失败", zap.Uint("userID", userID), zap.Error(err))
		}
		s.InvalidateUserCache(ctx, userID)
	}
}

// writeBasicAccessLog 写入基础访问权限变更日志（失败仅打 warn，不影响主流程）
func (s *RoleService) writeBasicAccessLog(userID uint, roleID uint, roleName, roleCode, action string) {
	username := ""
	if u, err := s.userRepo.GetByID(userID); err == nil {
		username = u.Nickname
	}
	logEntry := &model.AutoRoleLog{
		UserID:   userID,
		Username: username,
		RoleID:   roleID,
		RoleName: roleName,
		RoleCode: roleCode,
		Action:   action,
		Reason:   "basic_access",
	}
	if err := repository.NewAutoRoleRepository().CreateAutoRoleLog(logEntry); err != nil {
		global.Logger.Warn("[CorpCheck] 写入日志失败", zap.Error(err))
	}
}

// MigrateExistingUsers 将旧 User.Role 字段迁移到 user_role 表
func (s *RoleService) MigrateExistingUsers() {
	var users []model.User
	if err := global.DB.Find(&users).Error; err != nil {
		global.Logger.Error("迁移用户角色失败：查询用户", zap.Error(err))
		return
	}
	for _, u := range users {
		existing, _ := s.repo.GetUserRoleCodes(u.ID)
		if len(existing) > 0 {
			continue
		}
		roleName := u.Role
		if roleName == "" {
			roleName = model.RoleGuest
		}
		role, err := s.repo.GetByCode(roleName)
		if err != nil {
			global.Logger.Warn("迁移角色未找到", zap.String("role", roleName), zap.Uint("userID", u.ID))
			continue
		}
		if err := s.repo.AddUserRole(u.ID, role.ID); err != nil {
			global.Logger.Error("迁移用户角色失败", zap.Uint("userID", u.ID), zap.Error(err))
		}
	}
	global.Logger.Info("现有用户角色迁移完成")
}
