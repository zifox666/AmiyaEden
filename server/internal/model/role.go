package model

// 系统角色常量
const (
	RoleSuperAdmin = "super_admin" // 超级管理员，拥有全部权限
	RoleAdmin      = "admin"       // 管理员，拥有非技术性管理权限
	RoleSRP        = "srp"         // 补损管理员，可审批/发放补损及编辑舰船价格表
	RoleFC         = "fc"          // 舰队指挥，可创建/管理舰队、审批补损
	RoleUser       = "user"        // 已认证用户，基本访问权限
	RoleGuest      = "guest"       // 访客，只读公开信息
)

// rolePriority 角色优先级，数值越高代表权限越高
var rolePriority = map[string]int{
	RoleSuperAdmin: 100,
	RoleAdmin:      50,
	RoleSRP:        40,
	RoleFC:         30,
	RoleUser:       10,
	RoleGuest:      0,
}

// HasRole 检查 userRole 是否满足 requiredRole 所需的权限等级（支持继承）
func HasRole(userRole, requiredRole string) bool {
	up, ok1 := rolePriority[userRole]
	rp, ok2 := rolePriority[requiredRole]
	if !ok1 || !ok2 {
		return false
	}
	return up >= rp
}
