package middleware

import (
	"amiya-eden/internal/model"
	"amiya-eden/pkg/jwt"
	"amiya-eden/pkg/response"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	ContextKeyUserID      = "user_id"
	ContextKeyCharacterID = "character_id"
	ContextKeyRole        = "role"
)

// JWTAuth JWT 鉴权中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractToken(c)
		if tokenStr == "" {
			response.Fail(c, response.CodeUnauthorized, "未提供认证令牌")
			c.Abort()
			return
		}

		claims, err := jwt.ParseToken(tokenStr)
		if err != nil {
			response.Fail(c, response.CodeUnauthorized, err.Error())
			c.Abort()
			return
		}

		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyCharacterID, claims.CharacterID)
		c.Set(ContextKeyRole, claims.Role)
		c.Next()
	}
}

// extractToken 从 Authorization header 或 query 参数中提取 Bearer token
func extractToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return c.Query("token")
}

// GetUserID 从 gin.Context 中获取当前用户 ID
func GetUserID(c *gin.Context) uint {
	v, _ := c.Get(ContextKeyUserID)
	id, _ := v.(uint)
	return id
}

// GetCharacterID 从 gin.Context 中获取当前角色 ID
func GetCharacterID(c *gin.Context) int64 {
	v, _ := c.Get(ContextKeyCharacterID)
	id, _ := v.(int64)
	return id
}

// GetUserRole 从 gin.Context 中获取当前用户角色
func GetUserRole(c *gin.Context) string {
	v, _ := c.Get(ContextKeyRole)
	role, _ := v.(string)
	return role
}

// RequireRole 要求用户拥有该角色或更高权限（基于角色继承）。
// 必须在 JWTAuth() 中间件之后使用。
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := GetUserRole(c)
		if !model.HasRole(userRole, role) {
			response.Fail(c, response.CodeForbidden, "权限不足")
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireAnyRole 要求用户具有指定角色列表中任意一个（不考虑继承，精确匹配）。
// 适用于各角色权限相互独立的场景。
func RequireAnyRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := GetUserRole(c)
		for _, r := range roles {
			if userRole == r {
				c.Next()
				return
			}
		}
		response.Fail(c, response.CodeForbidden, "权限不足")
		c.Abort()
	}
}
