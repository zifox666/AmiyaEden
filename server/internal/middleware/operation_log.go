package middleware

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 跳过操作日志的路径（精确匹配）
var opLogSkipPaths = map[string]bool{
	"/health": true,
}

// CtxKeyBizCode context 中存储业务码的 key（由 ResponseWrapper 写入）
const CtxKeyBizCode = "biz_code"

// OperationLog API 操作日志中间件
// 注意：需注册在 ResponseWrapper 之前，使其 defer 在 ResponseWrapper 之后执行，
// 从而能读取到 biz_code。
func OperationLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// 跳过白名单路径
		if opLogSkipPaths[path] {
			c.Next()
			return
		}

		start := time.Now()

		c.Next()

		// ---- 请求结束后记录 ----
		latencyMs := time.Since(start).Milliseconds()

		// 操作人信息（由 JWT 中间件写入 context，无鉴权时为零值）
		userID, _ := c.Get("user_id")
		username, _ := c.Get("username")

		uid, _ := userID.(uint)
		uname, _ := username.(string)

		// biz_code 由 ResponseWrapper 在其 defer 中写入 context
		bizCode, _ := c.Get(CtxKeyBizCode)
		biz, _ := bizCode.(int)

		entry := model.OperationLog{
			RequestID:  c.GetString("request-id"),
			UserID:     uid,
			Username:   uname,
			IP:         c.ClientIP(),
			Method:     c.Request.Method,
			Path:       path,
			Query:      c.Request.URL.RawQuery,
			StatusCode: c.Writer.Status(),
			BizCode:    biz,
			LatencyMs:  latencyMs,
			UserAgent:  c.Request.UserAgent(),
		}

		// 异步写入 DB，不阻塞响应
		go func(log model.OperationLog) {
			if err := global.DB.Create(&log).Error; err != nil {
				global.Logger.Warn("操作日志写入失败",
					zap.String("path", log.Path),
					zap.Error(err),
				)
			}
		}(entry)
	}
}
