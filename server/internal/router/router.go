package router

import (
	"amiya-eden/internal/handler"
	"amiya-eden/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册所有业务路由
func RegisterRoutes(r *gin.Engine) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		registerSSORoutes(v1)
		registerUserRoutes(v1)
	}
}

// registerSSORoutes EVE SSO 登录相关路由
func registerSSORoutes(rg *gin.RouterGroup) {
	ssoHandler := handler.NewEveSSOHandler()
	sso := rg.Group("/sso/eve")
	{
		// 登录入口：重定向到 EVE 授权页
		sso.GET("/login", ssoHandler.Login)
		// EVE SSO 回调地址
		sso.GET("/callback", ssoHandler.Callback)
		// 获取已注册的 ESI Scope 列表（公开）
		sso.GET("/scopes", ssoHandler.GetScopes)
		// 获取当前用户绑定的所有角色（需要登录）
		sso.GET("/characters", middleware.JWTAuth(), ssoHandler.GetMyCharacters)
	}
}

// registerUserRoutes 用户管理路由（需要登录和相应权限）
func registerUserRoutes(rg *gin.RouterGroup) {
	userHandler := handler.NewUserHandler()
	users := rg.Group("/users", middleware.JWTAuth())
	{
		users.GET("", userHandler.List)
		users.GET("/:id", userHandler.Get)
		users.DELETE("/:id", userHandler.Delete)
	}
}
