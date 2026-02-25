package middleware

import (
	"amiya-eden/internal/model"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/jwt"
	"amiya-eden/pkg/response"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	ctxKeyUserID      = "userID"
	ctxKeyCharacterID = "characterID"
	ctxKeyUserRole    = "userRole" // JWT 中的单角色字段（向后兼容）
	ctxKeyRoles       = "roles"
	ctxKeyPermissions = "permissions"
)

// JWTAuth JWT 认证中间件
// 解析 Token，加载用户角色与权限到 context
func JWTAuth() gin.HandlerFunc {
	roleSvc := service.NewRoleService()

	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			response.Fail(c, response.CodeUnauthorized, "未提供认证令牌")
			c.Abort()
			return
		}

		claims, err := jwt.ParseToken(token)
		if err != nil {
			response.Fail(c, response.CodeUnauthorized, "令牌无效或已过期")
			c.Abort()
			return
		}

		c.Set(ctxKeyUserID, claims.UserID)
		c.Set(ctxKeyCharacterID, claims.CharacterID)
		c.Set(ctxKeyUserRole, claims.Role) // JWT 单角色（向后兼容）

		// 加载用户角色（带 Redis 缓存）
		roles, err := roleSvc.GetUserRoleNames(c.Request.Context(), claims.UserID)
		if err != nil {
			roles = []string{model.RoleGuest}
		}
		c.Set(ctxKeyRoles, roles)

		// 加载用户权限（带 Redis 缓存）
		perms, err := roleSvc.GetUserPermissions(c.Request.Context(), claims.UserID)
		if err != nil {
			perms = []string{}
		}
		c.Set(ctxKeyPermissions, perms)

		c.Next()
	}
}

// RequireRole 要求用户拥有指定角色之一（super_admin 自动通过）
func RequireRole(codes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles := GetUserRoles(c)
		for _, code := range codes {
			if model.HasAnyRoleMatch(roles, code) {
				c.Next()
				return
			}
		}
		response.Fail(c, response.CodeForbidden, "权限不足，需要角色: "+strings.Join(codes, "/"))
		c.Abort()
	}
}

// RequirePermission 要求用户拥有指定权限标识之一（super_admin 自动通过）
func RequirePermission(perms ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles := GetUserRoles(c)
		if model.IsSuperAdmin(roles) {
			c.Next()
			return
		}
		userPerms := GetUserPermissions(c)
		for _, required := range perms {
			for _, have := range userPerms {
				if have == required {
					c.Next()
					return
				}
			}
		}
		response.Fail(c, response.CodeForbidden, "权限不足，需要权限: "+strings.Join(perms, "/"))
		c.Abort()
	}
}

// ─── Context 辅助函数 ───

func GetUserID(c *gin.Context) uint {
	return c.GetUint(ctxKeyUserID)
}

func GetCharacterID(c *gin.Context) int64 {
	v, _ := c.Get(ctxKeyCharacterID)
	if id, ok := v.(int64); ok {
		return id
	}
	return 0
}

// GetUserRole 获取 JWT 中的单角色字段（向后兼容 fleet 等模块）
func GetUserRole(c *gin.Context) string {
	return c.GetString(ctxKeyUserRole)
}

func GetUserRoles(c *gin.Context) []string {
	v, exists := c.Get(ctxKeyRoles)
	if !exists {
		return nil
	}
	roles, _ := v.([]string)
	return roles
}

func GetUserPermissions(c *gin.Context) []string {
	v, exists := c.Get(ctxKeyPermissions)
	if !exists {
		return nil
	}
	perms, _ := v.([]string)
	return perms
}

func extractToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return auth[7:]
	}
	return c.Query("token")
}
