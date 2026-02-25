package model

// --- 系统角色编码常量 ---

const (
	RoleSuperAdmin = "super_admin"
	RoleAdmin      = "admin"
	RoleSRP        = "srp"
	RoleFC         = "fc"
	RoleUser       = "user"
	RoleGuest      = "guest"
)

// --- 数据模型 ---

// Role 角色
type Role struct {
	BaseModel
	Code        string `gorm:"size:50;uniqueIndex"  json:"code"`
	Name        string `gorm:"size:100"             json:"name"`
	Description string `gorm:"size:500"             json:"description"`
	IsSystem    bool   `gorm:"default:false"        json:"is_system"`
	Sort        int    `gorm:"default:0"            json:"sort"`
	Status      int8   `gorm:"default:1"            json:"status"`
	MenuIDs     []uint `gorm:"-"                    json:"menu_ids,omitempty"`
}

func (Role) TableName() string { return "role" }

// RoleMenu 角色-菜单关联
type RoleMenu struct {
	RoleID uint `gorm:"primaryKey;autoIncrement:false" json:"role_id"`
	MenuID uint `gorm:"primaryKey;autoIncrement:false" json:"menu_id"`
}

func (RoleMenu) TableName() string { return "role_menu" }

// UserRole 用户-角色关联
type UserRole struct {
	UserID uint `gorm:"primaryKey;autoIncrement:false" json:"user_id"`
	RoleID uint `gorm:"primaryKey;autoIncrement:false" json:"role_id"`
}

func (UserRole) TableName() string { return "user_role" }

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

// --- 系统角色种子数据 ---

var SystemRoleSeeds = []Role{
	{Code: RoleSuperAdmin, Name: "超级管理员", Description: "拥有系统全部权限", IsSystem: true, Sort: 100, Status: 1},
	{Code: RoleAdmin, Name: "管理员", Description: "系统管理权限", IsSystem: true, Sort: 90, Status: 1},
	{Code: RoleSRP, Name: "SRP管理员", Description: "补损审批与舰船价格管理", IsSystem: true, Sort: 80, Status: 1},
	{Code: RoleFC, Name: "FC", Description: "舰队指挥，管理舰队与活动", IsSystem: true, Sort: 70, Status: 1},
	{Code: RoleUser, Name: "用户", Description: "已认证用户，基本访问权限", IsSystem: true, Sort: 10, Status: 1},
	{Code: RoleGuest, Name: "访客", Description: "访客，只读公开信息", IsSystem: true, Sort: 0, Status: 1},
}
