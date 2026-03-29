package model

// --- 系统角色编码常量 ---

const (
	RoleSuperAdmin = "super_admin"
	RoleAdmin      = "admin"
	RoleSRP        = "srp"
	RoleFC         = "fc"
	RoleSeniorFC   = "senior_fc"
	RoleCaptain    = "captain"
	RoleWelfare    = "welfare"
	RoleUser       = "user"
	RoleGuest      = "guest"
)

// --- 数据模型 ---

// UserRole 用户-角色关联（直接存储角色编码，不再依赖 role 表）
type UserRole struct {
	UserID   uint   `gorm:"primaryKey;autoIncrement:false" json:"user_id"`
	RoleCode string `gorm:"primaryKey;size:50"             json:"role_code"`
}

func (UserRole) TableName() string { return "user_role" }

// --- 角色定义（纯内存，不入库）---

// RoleDefinition 系统角色定义，供前端展示和管理接口使用
type RoleDefinition struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	I18nKey     string `json:"i18nKey"`
	Description string `json:"description"`
	Sort        int    `json:"sort"`
}

// SystemRoleDefinitions 系统角色定义列表（按 Sort 降序排列）
var SystemRoleDefinitions = []RoleDefinition{
	{Code: RoleSuperAdmin, Name: "Super Admin", I18nKey: "roles.super_admin", Description: "Has full system permissions", Sort: 100},
	{Code: RoleAdmin, Name: "Admin", I18nKey: "roles.admin", Description: "System administration permissions", Sort: 90},
	{Code: RoleSeniorFC, Name: "Senior FC", I18nKey: "roles.senior_fc", Description: "Senior Fleet Commander, manages fleet configurations and skill plans", Sort: 85},
	{Code: RoleFC, Name: "FC", I18nKey: "roles.fc", Description: "Fleet Commander, manages fleets and operations", Sort: 70},
	{Code: RoleSRP, Name: "SRP Officer", I18nKey: "roles.srp", Description: "SRP approval and ship price management", Sort: 60},
	{Code: RoleWelfare, Name: "Welfare Officer", I18nKey: "roles.welfare", Description: "Corporation welfare approval and management", Sort: 50},
	{Code: RoleCaptain, Name: "Captain", I18nKey: "roles.captain", Description: "Newbro mentor captain view permissions", Sort: 30},
	{Code: RoleUser, Name: "Verified User", I18nKey: "roles.user", Description: "Authenticated user with basic access", Sort: 10},
	{Code: RoleGuest, Name: "Guest", I18nKey: "roles.guest", Description: "Guest, read-only access to public information", Sort: 0},
}

// roleDefinitionMap 角色编码到定义的映射（内部使用）
var roleDefinitionMap map[string]RoleDefinition

func init() {
	roleDefinitionMap = make(map[string]RoleDefinition, len(SystemRoleDefinitions))
	for _, def := range SystemRoleDefinitions {
		roleDefinitionMap[def.Code] = def
	}
}

// GetRoleDefinition 根据角色编码获取角色定义
func GetRoleDefinition(code string) (RoleDefinition, bool) {
	def, ok := roleDefinitionMap[code]
	return def, ok
}

// IsValidRoleCode 检查角色编码是否为已知的系统角色
func IsValidRoleCode(code string) bool {
	_, ok := roleDefinitionMap[code]
	return ok
}

// --- 角色检查辅助函数 ---

// IsSuperAdmin 检查角色列表中是否包含超级管理员
func IsSuperAdmin(roleCodes []string) bool {
	for _, code := range roleCodes {
		if code == RoleSuperAdmin {
			return true
		}
	}
	return false
}

// ContainsRole 检查角色列表中是否包含指定角色
func ContainsRole(roleCodes []string, target string) bool {
	for _, code := range roleCodes {
		if code == target {
			return true
		}
	}
	return false
}

// ContainsAnyRole 检查角色列表中是否包含指定角色中的任意一个
func ContainsAnyRole(roleCodes []string, targets ...string) bool {
	set := make(map[string]struct{}, len(roleCodes))
	for _, code := range roleCodes {
		set[code] = struct{}{}
	}
	for _, target := range targets {
		if _, ok := set[target]; ok {
			return true
		}
	}
	return false
}

// HasNonGuestRole 检查角色列表中是否存在任一非 guest 角色
func HasNonGuestRole(roleCodes []string) bool {
	for _, code := range roleCodes {
		if code != RoleGuest {
			return true
		}
	}
	return false
}

// NormalizeRoleCodes returns active role codes ordered by priority, falling back
// to the legacy single-role field when the association table is empty.
func NormalizeRoleCodes(roleCodes []string, fallback string) []string {
	if len(roleCodes) > 0 {
		return roleCodes
	}
	if fallback != "" {
		return []string{fallback}
	}
	return []string{RoleGuest}
}

// HasAnyRoleMatch 检查用户角色列表中是否有满足 requiredRole 的角色
// 超级管理员拥有所有权限
func HasAnyRoleMatch(userRoles []string, requiredRole string) bool {
	if IsSuperAdmin(userRoles) {
		return true
	}
	return ContainsRole(userRoles, requiredRole)
}

// HasRole 兼容接口
func HasRole(userRole, requiredRole string) bool {
	if userRole == RoleSuperAdmin {
		return true
	}
	return userRole == requiredRole
}
