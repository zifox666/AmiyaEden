package bootstrap

import (
	"amiya-eden/global"
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/router"

	"github.com/gin-gonic/gin"
)

// InitRouter 初始化并返回 Gin 路由引擎
func InitRouter() *gin.Engine {
	gin.SetMode(global.Config.Server.Mode)

	r := gin.New()

	// 全局中间件（注册顺序即执行顺序，defer 逆序执行）
	// 执行顺序(before): RequestID → OperationLog → ResponseWrapper → ZapLogger → ZapRecovery → Cors → handler
	// 执行顺序(after) : Cors → ZapRecovery → ZapLogger → ResponseWrapper(写biz_code) → OperationLog(读biz_code存DB)
	r.Use(
		middleware.RequestID(),
		middleware.OperationLog(),
		middleware.ResponseWrapper(),
		middleware.ZapLogger(),
		middleware.ZapRecovery(),
		middleware.Cors(),
	)

	// 注册业务路由
	router.RegisterRoutes(r)

	global.Logger.Info("路由注册完成")
	return r
}
