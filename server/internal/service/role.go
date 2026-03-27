package service

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"amiya-eden/internal/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type RoleService struct {
	repo     *repository.RoleRepository
	userRepo *repository.UserRepository
}

func NewRoleService() *RoleService {
	return &RoleService{
		repo:     repository.NewRoleRepository(),
		userRepo: repository.NewUserRepository(),
	}
}

// ─── Redis 缓存 ───

const (
	userRolesCachePrefix = "user_roles:"
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

// InvalidateUserCache 清除用户角色缓存
func (s *RoleService) InvalidateUserCache(ctx context.Context, userID uint) {
	global.Redis.Del(ctx, fmt.Sprintf("%s%d", userRolesCachePrefix, userID))
}

// InvalidateUserRolesCache 兼容旧接口
func (s *RoleService) InvalidateUserRolesCache(ctx context.Context, userID uint) {
	s.InvalidateUserCache(ctx, userID)
}

// ─── 角色定义查询 ───

// ListRoleDefinitions 返回系统角色定义列表
func (s *RoleService) ListRoleDefinitions() []model.RoleDefinition {
	return model.SystemRoleDefinitions
}

// ─── 用户角色管理 ───

// GetUserRoles 获取用户的角色定义列表
func (s *RoleService) GetUserRoles(userID uint) ([]model.RoleDefinition, error) {
	codes, err := s.repo.GetUserRoleCodes(userID)
	if err != nil {
		return nil, err
	}
	defs := make([]model.RoleDefinition, 0, len(codes))
	for _, code := range codes {
		if def, ok := model.GetRoleDefinition(code); ok {
			defs = append(defs, def)
		}
	}
	return defs, nil
}

func (s *RoleService) SetUserRoles(ctx context.Context, operatorID uint, operatorRoles []string, userID uint, roleCodes []string) error {
	currentCodes, err := s.repo.GetUserRoleCodes(userID)
	if err != nil {
		return err
	}

	// Validate all requested role codes
	for _, code := range roleCodes {
		if !model.IsValidRoleCode(code) {
			return fmt.Errorf("未知的角色编码: %s", code)
		}
		if code == model.RoleSuperAdmin && !model.IsSuperAdmin(operatorRoles) {
			return errors.New("只有超级管理员可以分配该角色")
		}
	}

	requestedCodes := normalizeAssignedRoleCodes(roleCodes)
	if err := validateSetUserRolesPermission(operatorID, userID, operatorRoles, currentCodes, requestedCodes); err != nil {
		return err
	}

	if err := s.repo.SetUserRoles(userID, requestedCodes); err != nil {
		return err
	}

	// 同步 User.Role 字段（取最高优先级角色）
	s.SyncUserPrimaryRole(userID)
	s.InvalidateUserCache(ctx, userID)
	return nil
}

func validateSetUserRolesPermission(operatorID, targetUserID uint, operatorRoles, currentCodes, requestedCodes []string) error {
	isSelfAdminEdit := operatorID == targetUserID && model.ContainsRole(operatorRoles, model.RoleAdmin)

	if !isSelfAdminEdit {
		if err := validateManageUserPermission(operatorRoles, currentCodes); err != nil {
			return err
		}
	}

	if model.IsSuperAdmin(operatorRoles) {
		return nil
	}

	if model.ContainsAnyRole(requestedCodes, model.RoleSuperAdmin) {
		return errors.New("只有超级管理员可以分配该角色")
	}
	if model.ContainsAnyRole(requestedCodes, model.RoleAdmin) && !isSelfAdminEdit {
		return errors.New("只有超级管理员可以分配管理员角色")
	}
	return nil
}

func normalizeAssignedRoleCodes(codes []string) []string {
	if len(codes) == 0 {
		return nil
	}

	hasNonGuestRole := false
	for _, code := range codes {
		if code != model.RoleGuest {
			hasNonGuestRole = true
			break
		}
	}

	// Deduplicate and filter
	seen := make(map[string]struct{}, len(codes))
	result := make([]string, 0, len(codes))
	for _, code := range codes {
		if _, dup := seen[code]; dup {
			continue
		}
		seen[code] = struct{}{}
		if hasNonGuestRole && code == model.RoleGuest {
			continue
		}
		result = append(result, code)
	}
	return result
}

// ─── 内部辅助 ───

func (s *RoleService) SyncUserPrimaryRole(userID uint) {
	codes, err := s.repo.GetUserRoleCodes(userID)
	if err != nil || len(codes) == 0 {
		_ = s.userRepo.UpdateRole(userID, model.RoleGuest)
		return
	}
	// 取第一个（已按 sort 排序）
	_ = s.userRepo.UpdateRole(userID, codes[0])
}

// CheckCorpAccessAndAdjustRole 检查用户名下所有角色的军团归属是否在准入列表内
func (s *RoleService) CheckCorpAccessAndAdjustRole(ctx context.Context, userID uint) error {
	allowCorps := utils.GetAllowCorporations()
	allowSet := make(map[int64]struct{}, len(allowCorps))
	for _, id := range allowCorps {
		allowSet[id] = struct{}{}
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	charRepo := repository.NewEveCharacterRepository()
	chars, err := charRepo.ListByUserID(userID)
	if err != nil {
		return err
	}

	hasAccess := hasAllowedPrimaryCharacter(user.PrimaryCharacterID, chars, allowSet)

	rollCodes, err := s.repo.GetUserRoleCodes(userID)
	if err != nil {
		return err
	}

	if model.ContainsAnyRole(rollCodes, model.RoleSuperAdmin) {
		return nil
	}

	if hasAccess {
		if model.ContainsRole(rollCodes, model.RoleUser) {
			return nil
		}
		_ = s.repo.RemoveUserRole(userID, model.RoleGuest)
		if err := s.repo.AddUserRole(userID, model.RoleUser); err != nil {
			return err
		}
		s.InvalidateUserCache(ctx, userID)
		s.SyncUserPrimaryRole(userID)
		global.Logger.Info("[CorpCheck] 用户升级为 user",
			zap.Uint("user_id", userID))
	} else {
		if len(rollCodes) == 1 && rollCodes[0] == model.RoleGuest {
			return nil
		}
		// 清除所有角色，降级为 guest
		if err := s.repo.SetUserRoles(userID, []string{model.RoleGuest}); err != nil {
			return err
		}
		s.InvalidateUserCache(ctx, userID)
		s.SyncUserPrimaryRole(userID)
		global.Logger.Info("[CorpCheck] 用户降级为 guest",
			zap.Uint("user_id", userID))
	}
	return nil
}

// EnsureUserHasRole 确保用户至少拥有指定角色（当用户还没有任何 user_role 记录时）
func (s *RoleService) EnsureUserHasRole(ctx context.Context, userID uint, roleCode string) {
	codes, err := s.repo.GetUserRoleCodes(userID)
	if err != nil || len(codes) == 0 {
		if err := s.repo.AddUserRole(userID, roleCode); err != nil {
			global.Logger.Error("分配角色失败", zap.Uint("userID", userID), zap.String("role", roleCode), zap.Error(err))
		}
		s.InvalidateUserCache(ctx, userID)
	}
}

// EnsureUserHasDefaultRole 兼容旧接口，默认补 guest
func (s *RoleService) EnsureUserHasDefaultRole(ctx context.Context, userID uint) {
	s.EnsureUserHasRole(ctx, userID, model.RoleGuest)
}

// MigrateUserRoleTableToCode 将旧的 user_role(user_id, role_id) 迁移为 (user_id, role_code)
// 在 bootstrap 阶段调用一次
func (s *RoleService) MigrateUserRoleTableToCode() {
	migrator := global.DB.Migrator()

	// If old role_id column still exists, perform migration
	if !migrator.HasColumn("user_role", "role_id") {
		return
	}

	global.Logger.Info("[Migration] 开始迁移 user_role 表: role_id → role_code")

	// Check if role_code column already exists (partial migration)
	hasRoleCode := migrator.HasColumn("user_role", "role_code")

	if !hasRoleCode {
		// Add role_code column
		if err := global.DB.Exec(`ALTER TABLE user_role ADD COLUMN role_code VARCHAR(50)`).Error; err != nil {
			global.Logger.Error("[Migration] 添加 role_code 列失败", zap.Error(err))
			return
		}
	}

	// Populate role_code from role table (if role table still exists)
	if migrator.HasTable("role") {
		if err := global.DB.Exec(`
			UPDATE user_role
			SET role_code = (SELECT code FROM role WHERE role.id = user_role.role_id)
			WHERE role_code IS NULL OR role_code = ''
		`).Error; err != nil {
			global.Logger.Error("[Migration] 填充 role_code 失败", zap.Error(err))
			return
		}
	}

	// Delete rows where role_code is still empty (orphaned references)
	global.DB.Exec(`DELETE FROM user_role WHERE role_code IS NULL OR role_code = ''`)

	// Drop old primary key and role_id column, set new primary key
	// Use raw SQL for atomic DDL
	ddlStatements := []string{
		`ALTER TABLE user_role DROP CONSTRAINT IF EXISTS user_role_pkey`,
		`ALTER TABLE user_role DROP COLUMN IF EXISTS role_id`,
		`ALTER TABLE user_role ALTER COLUMN role_code SET NOT NULL`,
		`ALTER TABLE user_role ADD PRIMARY KEY (user_id, role_code)`,
	}
	for _, stmt := range ddlStatements {
		if err := global.DB.Exec(stmt).Error; err != nil {
			global.Logger.Error("[Migration] DDL 执行失败", zap.String("sql", stmt), zap.Error(err))
			return
		}
	}

	global.Logger.Info("[Migration] user_role 表迁移完成")
}

// MigrateEsiMappingsToCode 将 esi_role_mapping / esi_title_mapping 的 role_id 迁移为 role_code
func (s *RoleService) MigrateEsiMappingsToCode() {
	migrator := global.DB.Migrator()

	for _, table := range []string{"esi_role_mapping", "esi_title_mapping"} {
		if !migrator.HasColumn(table, "role_id") {
			continue
		}

		global.Logger.Info("[Migration] 迁移 ESI 映射表", zap.String("table", table))

		hasRoleCode := migrator.HasColumn(table, "role_code")
		if !hasRoleCode {
			if err := global.DB.Exec(fmt.Sprintf(`ALTER TABLE %s ADD COLUMN role_code VARCHAR(50)`, table)).Error; err != nil {
				global.Logger.Error("[Migration] 添加 role_code 列失败", zap.String("table", table), zap.Error(err))
				continue
			}
		}

		// Populate from role table
		if migrator.HasTable("role") {
			if err := global.DB.Exec(fmt.Sprintf(`
				UPDATE %s
				SET role_code = (SELECT code FROM role WHERE role.id = %s.role_id)
				WHERE role_code IS NULL OR role_code = ''
			`, table, table)).Error; err != nil {
				global.Logger.Error("[Migration] 填充 role_code 失败", zap.String("table", table), zap.Error(err))
				continue
			}
		}

		// Delete orphaned rows
		global.DB.Exec(fmt.Sprintf(`DELETE FROM %s WHERE role_code IS NULL OR role_code = ''`, table))

		// Drop role_id column and update constraints
		ddl := []string{
			fmt.Sprintf(`ALTER TABLE %s DROP COLUMN IF EXISTS role_id`, table),
			fmt.Sprintf(`ALTER TABLE %s ALTER COLUMN role_code SET NOT NULL`, table),
		}

		// Recreate unique indexes without role_id
		if table == "esi_role_mapping" {
			ddl = append(ddl,
				`DROP INDEX IF EXISTS idx_esi_role_mapping`,
				`CREATE UNIQUE INDEX idx_esi_role_mapping ON esi_role_mapping (esi_role, role_code)`,
			)
		} else {
			ddl = append(ddl,
				`DROP INDEX IF EXISTS idx_esi_title_mapping`,
				`CREATE UNIQUE INDEX idx_esi_title_mapping ON esi_title_mapping (corporation_id, title_id, role_code)`,
			)
		}

		for _, stmt := range ddl {
			if err := global.DB.Exec(stmt).Error; err != nil {
				global.Logger.Error("[Migration] DDL 执行失败", zap.String("table", table), zap.String("sql", stmt), zap.Error(err))
			}
		}

		global.Logger.Info("[Migration] ESI 映射表迁移完成", zap.String("table", table))
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
		roleName := u.Role
		if roleName == "" {
			roleName = model.RoleGuest
		}
		if model.ContainsRole(existing, roleName) {
			continue
		}
		if !model.IsValidRoleCode(roleName) {
			global.Logger.Warn("迁移角色未找到", zap.String("role", roleName), zap.Uint("userID", u.ID))
			continue
		}
		if err := s.repo.AddUserRole(u.ID, roleName); err != nil {
			global.Logger.Error("迁移用户角色失败", zap.Uint("userID", u.ID), zap.Error(err))
		}
	}
	global.Logger.Info("现有用户角色迁移完成")
}
