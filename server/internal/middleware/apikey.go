package middleware

import (
	"amiya-eden/global"
	"amiya-eden/pkg/response"

	"github.com/gin-gonic/gin"
)

// APIKeyAuth API Key 鉴权中间件
// 从 Header X-API-Key 或 Query 参数 api_key 中读取。
func APIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-API-Key")
		if key == "" {
			key = c.Query("api_key")
		}

		expected := global.Config.SDE.APIKey
		if expected == "" || key != expected {
			response.Fail(c, response.CodeUnauthorized, "无效的 API Key")
			c.Abort()
			return
		}
		c.Next()
	}
}
