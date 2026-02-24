package middleware

import (
	"amiya-eden/pkg/jwt"
	"amiya-eden/pkg/response"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	ContextKeyUserID      = "user_id"
	ContextKeyCharacterID = "character_id"
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
