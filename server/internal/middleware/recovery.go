package middleware

import (
	"amiya-eden/global"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ZapRecovery panic 恢复中间件，记录堆栈到 zap
func ZapRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 判断是否是连接断开
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						errStr := strings.ToLower(se.Error())
						if strings.Contains(errStr, "broken pipe") || strings.Contains(errStr, "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					global.Logger.Error(c.Request.URL.Path,
						zap.Any("error", err),
						zap.ByteString("request", httpRequest),
					)
					c.Error(err.(error))
					c.Abort()
					return
				}

				global.Logger.Error("[Recovery] panic 已恢复",
					zap.Any("error", err),
					zap.ByteString("request", httpRequest),
					zap.ByteString("stack", debug.Stack()),
				)
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}
