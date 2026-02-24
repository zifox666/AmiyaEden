package middleware

import (
	"net/http"
	"time"

	"amiya-eden/global"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ZapLogger Gin 请求日志中间件（基于 zap）
func ZapLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		cost := time.Since(start)
		status := c.Writer.Status()

		fields := []zap.Field{
			zap.Int("status", status),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.Duration("cost", cost),
			zap.String("request-id", c.GetString("request-id")),
		}

		if len(c.Errors) > 0 {
			for _, e := range c.Errors.Errors() {
				global.Logger.Error(e, fields...)
			}
			return
		}

		if status >= http.StatusInternalServerError {
			global.Logger.Error("服务器错误", fields...)
		} else if status >= http.StatusBadRequest {
			global.Logger.Warn("客户端错误", fields...)
		} else {
			global.Logger.Info("请求完成", fields...)
		}
	}
}
