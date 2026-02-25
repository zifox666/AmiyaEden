package router

import (
	"amiya-eden/internal/handler"
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/model"

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
		registerSdeRoutes(v1)
		registerESIRefreshRoutes(v1)
		registerMenuRoutes(v1)
		registerOperationRoutes(v1)
		registerSrpRoutes(v1)
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
		// 绑定新角色：发起 EVE SSO 授权（需要登录）
		sso.GET("/bind", middleware.JWTAuth(), ssoHandler.BindLogin)
		// 设置主角色（需要登录）
		sso.PUT("/primary/:character_id", middleware.JWTAuth(), ssoHandler.SetPrimary)
		// 解绑角色（需要登录）
		sso.DELETE("/characters/:character_id", middleware.JWTAuth(), ssoHandler.Unbind)
	}
}

// registerSdeRoutes SDE 数据资产路由
//
//	GET  /sde/version       - 查看当前 SDE 版本
//	POST /sde/update        - 手动触发更新（需要 Admin+）
//	GET  /sde/translation   - 精确查询翻译（需要 API Key）
//	GET  /sde/translations  - 查询 key 所有语言（需要 API Key）
//	GET  /sde/search        - 名称模糊搜索（需要 API Key）
//	GET  /sde/type/:type_id - typeID 详情（需要 API Key）
func registerSdeRoutes(rg *gin.RouterGroup) {
	sdeHandler := handler.NewSdeHandler()
	sde := rg.Group("/sde")
	{
		// 公开接口：查看版本
		sde.GET("/version", sdeHandler.GetVersion)

		// 管理接口：手动触发更新（需要 JWT + Admin）
		sde.POST("/update",
			middleware.JWTAuth(),
			middleware.RequireRole(model.RoleAdmin),
			sdeHandler.TriggerUpdate,
		)

		// 数据查询接口：需要 API Key
		apiRoutes := sde.Group("", middleware.APIKeyAuth())
		{
			apiRoutes.GET("/translation", sdeHandler.GetTranslation)
			apiRoutes.GET("/translations", sdeHandler.GetTranslationsByKey)
			apiRoutes.GET("/search", sdeHandler.SearchByName)
			apiRoutes.GET("/type/:type_id", sdeHandler.GetTypeDetail)
		}
	}
}

// registerUserRoutes 用户管理路由
//
//	GET    /users           - 查看用户列表  需要 Admin 或以上
//	GET    /users/:id       - 查看用户详情  需要 Admin 或以上
//	DELETE /users/:id       - 删除用户      需要 Admin 或以上
//	PATCH  /users/:id/role  - 修改用户角色 需要 Admin 或以上
//	GET    /me              - 获取当前登录用户信息（需要 JWT）
func registerUserRoutes(rg *gin.RouterGroup) {
	userHandler := handler.NewUserHandler()

	// 当前登录用户信息（任意已登录用户可访问）
	rg.GET("/me", middleware.JWTAuth(), userHandler.GetMe)

	// 所有用户管理接口先通过 JWT 鉴权
	users := rg.Group("/users", middleware.JWTAuth())
	{
		// Admin 及以上可访问
		adminRoutes := users.Group("", middleware.RequireRole(model.RoleAdmin))
		{
			adminRoutes.GET("", userHandler.List)
			adminRoutes.GET("/:id", userHandler.Get)
			adminRoutes.DELETE("/:id", userHandler.Delete)
			adminRoutes.PATCH("/:id/role", userHandler.UpdateRole)
		}
	}
}

// registerMenuRoutes 动态菜单路由
//
//	GET /menu - 获取当前登录用户的菜单列表（需要 JWT）
func registerMenuRoutes(rg *gin.RouterGroup) {
	menuHandler := handler.NewMenuHandler()
	rg.GET("/menu", middleware.JWTAuth(), menuHandler.GetMenuList)
}

// registerESIRefreshRoutes ESI 数据刷新队列路由
//
//	GET  /esi/refresh/tasks      - 获取所有任务定义
//	GET  /esi/refresh/statuses   - 获取任务运行状态（分页）
//	POST /esi/refresh/run        - 手动触发指定任务（单角色）
//	POST /esi/refresh/run-task   - 手动触发指定任务（所有角色）
//	POST /esi/refresh/run-all    - 手动触发全量刷新
func registerESIRefreshRoutes(rg *gin.RouterGroup) {
	h := handler.NewESIRefreshHandler()
	esi := rg.Group("/esi/refresh", middleware.JWTAuth(), middleware.RequireRole(model.RoleAdmin))
	{
		esi.GET("/tasks", h.GetTasks)
		esi.GET("/statuses", h.GetStatuses)
		esi.POST("/run", h.RunTask)
		esi.POST("/run-task", h.RunTaskByName)
		esi.POST("/run-all", h.RunAll)
	}
}

// registerOperationRoutes 舰队行动管理路由
//
//	POST   /operation/fleets                    - 创建舰队（FC+）
//	GET    /operation/fleets                    - 查看舰队列表（FC+）
//	GET    /operation/fleets/:id                - 舰队详情（FC+）
//	PUT    /operation/fleets/:id                - 更新舰队（FC+）
//	DELETE /operation/fleets/:id                - 删除舰队（FC+）
//	GET    /operation/fleets/:id/members        - 获取成员列表（FC+）
//	POST   /operation/fleets/:id/members/sync   - ESI 成员同步（FC+）
//	POST   /operation/fleets/:id/pap            - 发放 PAP（FC+）
//	GET    /operation/fleets/:id/pap            - PAP 发放记录（FC+）
//	POST   /operation/fleets/:id/invites        - 创建邀请链接（FC+）
//	GET    /operation/fleets/:id/invites        - 获取邀请链接（FC+）
//	DELETE /operation/fleets/invites/:invite_id - 禁用邀请链接（FC+）
//	POST   /operation/fleets/join               - 通过邀请码加入（已登录）
//	GET    /operation/fleets/pap/me             - 我的 PAP 记录（已登录）
//	GET    /operation/fleets/esi/:character_id  - 角色 ESI 舰队信息（已登录）
//	GET    /operation/wallet                    - 我的钱包（已登录）
//	GET    /operation/wallet/transactions       - 我的钱包流水（已登录）
func registerOperationRoutes(rg *gin.RouterGroup) {
	h := handler.NewFleetHandler()
	op := rg.Group("/operation", middleware.JWTAuth())
	{
		// ── 需要 FC 或更高权限 ──
		fleets := op.Group("/fleets", middleware.RequireRole(model.RoleFC))
		{
			fleets.POST("", h.CreateFleet)
			fleets.GET("", h.ListFleets)
			fleets.GET("/:id", h.GetFleet)
			fleets.PUT("/:id", h.UpdateFleet)
			fleets.DELETE("/:id", h.DeleteFleet)
			fleets.POST("/:id/refresh-esi", h.RefreshFleetESI)

			// 成员
			fleets.GET("/:id/members", h.GetMembers)
			fleets.POST("/:id/members/sync", h.SyncESIMembers)

			// PAP
			fleets.POST("/:id/pap", h.IssuePap)
			fleets.GET("/:id/pap", h.GetPapLogs)

			// 邀请链接
			fleets.POST("/:id/invites", h.CreateInvite)
			fleets.GET("/:id/invites", h.GetInvites)
			fleets.DELETE("/invites/:invite_id", h.DeactivateInvite)
		}

		// ── 已登录用户可访问 ──
		op.POST("/fleets/join", h.JoinFleet)
		op.GET("/fleets/pap/me", h.GetMyPapLogs)
		op.GET("/fleets/esi/:character_id", h.GetCharacterFleetInfo)

		// 钱包
		op.GET("/wallet", h.GetWallet)
		op.GET("/wallet/transactions", h.GetWalletTransactions)
	}
}

// registerSrpRoutes 补损管理路由
//
//	GET    /srp/prices                           - 查看舰船价格表（已登录）
//	POST   /srp/prices                           - 新增/更新价格（SRP+）
//	DELETE /srp/prices/:id                       - 删除价格（SRP+）
//	GET    /srp/fleet-killmails                  - 获取舰队范围内我的 KM 列表（已登录）
//	GET    /srp/my-killmails                     - 获取我的全部 KM（已登录，不限舰队）
//	POST   /srp/applications                     - 提交补损申请（已登录）
//	GET    /srp/applications/my                  - 我的补损申请（已登录）
//	GET    /srp/manage/applications              - 查看全部申请（FC+）
//	GET    /srp/manage/applications/:id          - 申请详情（FC+）
//	PATCH  /srp/manage/applications/:id/review   - 审批（FC+）
//	PATCH  /srp/manage/applications/:id/payout   - 发放（SRP+）
func registerSrpRoutes(rg *gin.RouterGroup) {
	h := handler.NewSrpHandler()
	srp := rg.Group("/srp", middleware.JWTAuth())
	{
		// ── 已登录用户可访问 ──
		srp.GET("/prices", h.ListShipPrices)
		srp.GET("/fleet-killmails", h.GetFleetKillmails)
		srp.GET("/my-killmails", h.GetMyKillmails)
		srp.POST("/applications", h.SubmitApplication)
		srp.GET("/applications/my", h.ListMyApplications)

		// ── 舰船价格表管理（SRP+） ──
		srpAdmin := srp.Group("", middleware.RequireRole(model.RoleSRP))
		{
			srpAdmin.POST("/prices", h.UpsertShipPrice)
			srpAdmin.DELETE("/prices/:id", h.DeleteShipPrice)
		}

		// ── 审批（FC+） ──
		manage := srp.Group("/manage", middleware.RequireRole(model.RoleFC))
		{
			manage.GET("/applications", h.ListApplications)
			manage.GET("/applications/:id", h.GetApplication)
			manage.PATCH("/applications/:id/review", h.ReviewApplication)
		}

		// ── 发放（SRP+） ──
		srpPayout := srp.Group("/manage", middleware.RequireRole(model.RoleSRP))
		{
			srpPayout.PATCH("/applications/:id/payout", h.Payout)
		}
	}
}
