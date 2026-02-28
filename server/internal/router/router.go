package router

import (
	"amiya-eden/internal/handler"
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/model"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册所有业务路由
func RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")

	// ─── 无需认证 ───
	ssoH := handler.NewEveSSOHandler()
	sso := api.Group("/sso/eve")
	{
		sso.GET("/login", ssoH.Login)
		sso.GET("/callback", ssoH.Callback)
	}

	// ─── SDE 公开查询（API Key 鉴权）───
	sdeH := handler.NewSdeHandler()
	sde := api.Group("/sde")
	{
		sde.GET("/version", sdeH.GetVersion)
		sde.POST("/types", sdeH.GetTypes)
		sde.POST("/names", sdeH.GetNames)
	}

	// ─── 需要登录 ───
	auth := api.Group("", middleware.JWTAuth())

	// SSO 角色管理（绑定/解绑/设主角色）
	ssoAuth := auth.Group("/sso/eve")
	{
		ssoAuth.GET("/scopes", ssoH.GetScopes)
		ssoAuth.GET("/characters", ssoH.GetMyCharacters)
		ssoAuth.GET("/bind", ssoH.BindLogin)
		ssoAuth.PUT("/primary/:character_id", ssoH.SetPrimary)
		ssoAuth.DELETE("/characters/:character_id", ssoH.Unbind)
	}

	// ─── 当前用户 ───
	meH := handler.NewMeHandler()
	auth.GET("/me", meH.GetMe)

	// ─── 菜单 ───
	menuH := handler.NewMenuHandler()
	auth.GET("/menu/list", menuH.GetMenuList) // 当前用户可用菜单

	// ─── 舰队 ───
	fleetH := handler.NewFleetHandler()
	operation := auth.Group("/operation")
	fleet := operation.Group("/fleets")
	{
		fleet.POST("", fleetH.CreateFleet)
		fleet.GET("", fleetH.ListFleets)
		fleet.GET("/:id", fleetH.GetFleet)
		fleet.PUT("/:id", fleetH.UpdateFleet)
		fleet.DELETE("/:id", fleetH.DeleteFleet)
		fleet.POST("/:id/refresh-esi", fleetH.RefreshFleetESI)

		// 成员
		fleet.GET("/:id/members", fleetH.GetMembers)
		fleet.POST("/:id/members/sync", fleetH.SyncESIMembers)

		// PAP
		fleet.POST("/:id/pap", fleetH.IssuePap)
		fleet.GET("/:id/pap", fleetH.GetPapLogs)
		fleet.GET("/pap/me", fleetH.GetMyPapLogs)

		// 邀请
		fleet.POST("/:id/invites", fleetH.CreateInvite)
		fleet.GET("/:id/invites", fleetH.GetInvites)
		fleet.DELETE("/invites/:invite_id", fleetH.DeactivateInvite)
		fleet.POST("/join", fleetH.JoinFleet)

		// 查角色所在舰队
		fleet.GET("/esi/:character_id", fleetH.GetCharacterFleetInfo)
	}

	// ─── 系统钱包（用户端）───
	walletH := handler.NewSysWalletHandler()
	wallet := operation.Group("/wallet")
	{
		wallet.POST("/my", walletH.GetMyWallet)
		wallet.POST("/my/transactions", walletH.GetMyTransactions)
	}

	// ─── 商店（用户端）───
	shopH := handler.NewShopHandler()
	shop := auth.Group("/shop")
	{
		shop.POST("/products", shopH.ListProducts)
		shop.POST("/product/detail", shopH.GetProductDetail)
		shop.POST("/buy", shopH.BuyProduct)
		shop.POST("/orders", shopH.GetMyOrders)
		shop.POST("/redeem/list", shopH.GetMyRedeemCodes)
	}

	// ─── SRP 补损 ───
	srpH := handler.NewSrpHandler()
	srp := auth.Group("/srp")
	{
		// 价格表（查看公开，修改需权限）
		srp.GET("/prices", srpH.ListShipPrices)
		srp.POST("/prices", middleware.RequirePermission("srp:price:edit"), srpH.UpsertShipPrice)
		srp.DELETE("/prices/:id", middleware.RequirePermission("srp:price:edit"), srpH.DeleteShipPrice)

		// 个人申请
		srp.POST("/applications", srpH.SubmitApplication)
		srp.GET("/applications/me", srpH.ListMyApplications)
		srp.GET("/killmails/me", srpH.GetMyKillmails)
		srp.GET("/killmails/fleet/:fleet_id", srpH.GetFleetKillmails)

		// 审核（需权限）
		srpAdmin := srp.Group("", middleware.RequirePermission("srp:review"))
		{
			srpAdmin.GET("/applications", srpH.ListApplications)
			srpAdmin.GET("/applications/:id", srpH.GetApplication)
			srpAdmin.PUT("/applications/:id/review", srpH.ReviewApplication)
			srpAdmin.PUT("/applications/:id/payout", srpH.Payout)
		}
	}

	// ─── ESI 刷新队列 ───
	esiH := handler.NewESIRefreshHandler()
	esiRefresh := auth.Group("/esi/refresh", middleware.RequireRole(model.RoleAdmin))
	{
		esiRefresh.GET("/tasks", esiH.GetTasks)
		esiRefresh.GET("/statuses", esiH.GetStatuses)
		esiRefresh.POST("/run", esiH.RunTask)
		esiRefresh.POST("/run-task", esiH.RunTaskByName)
		esiRefresh.POST("/run-all", esiH.RunAll)
	}

	// ─── 系统管理（需要 admin 角色）───
	admin := auth.Group("/system", middleware.RequireRole(model.RoleAdmin))

	// 菜单管理
	adminMenu := admin.Group("/menu")
	{
		adminMenu.GET("/tree", menuH.GetMenuTree)
		adminMenu.POST("", menuH.CreateMenu)
		adminMenu.PUT("/:id", menuH.UpdateMenu)
		adminMenu.DELETE("/:id", menuH.DeleteMenu)
	}

	// 角色管理
	roleH := handler.NewRoleHandler()
	adminRole := admin.Group("/role")
	{
		adminRole.GET("", roleH.ListRoles)
		adminRole.GET("/all", roleH.ListAllRoles)
		adminRole.GET("/:id", roleH.GetRole)
		adminRole.POST("", roleH.CreateRole)
		adminRole.PUT("/:id", roleH.UpdateRole)
		adminRole.DELETE("/:id", roleH.DeleteRole)

		// 角色权限
		adminRole.GET("/:id/menus", roleH.GetRoleMenus)
		adminRole.PUT("/:id/menus", roleH.SetRoleMenus)
	}

	// 用户管理
	userH := handler.NewUserHandler()
	adminUser := admin.Group("/user")
	{
		adminUser.GET("", userH.ListUsers)
		adminUser.GET("/:id", userH.GetUser)
		adminUser.PUT("/:id", userH.UpdateUser)
		adminUser.DELETE("/:id", userH.DeleteUser)

		// 用户角色分配
		adminUser.GET("/:id/roles", roleH.GetUserRoles)
		adminUser.PUT("/:id/roles", roleH.SetUserRoles)
	}

	// 系统钱包管理（管理员）
	adminWalletH := handler.NewSysWalletHandler()
	adminWallet := admin.Group("/wallet")
	{
		adminWallet.POST("/list", adminWalletH.AdminListWallets)
		adminWallet.POST("/detail", adminWalletH.AdminGetWallet)
		adminWallet.POST("/adjust", adminWalletH.AdminAdjust)
		adminWallet.POST("/transactions", adminWalletH.AdminListTransactions)
		adminWallet.POST("/logs", adminWalletH.AdminListLogs)
	}

	// 商店管理（管理员）
	adminShopH := handler.NewShopHandler()
	adminShopProduct := admin.Group("/shop/product")
	{
		adminShopProduct.POST("/list", adminShopH.AdminListProducts)
		adminShopProduct.POST("/add", adminShopH.AdminCreateProduct)
		adminShopProduct.POST("/edit", adminShopH.AdminUpdateProduct)
		adminShopProduct.POST("/delete", adminShopH.AdminDeleteProduct)
	}
	adminShopOrder := admin.Group("/shop/order")
	{
		adminShopOrder.POST("/list", adminShopH.AdminListOrders)
		adminShopOrder.POST("/approve", adminShopH.AdminApproveOrder)
		adminShopOrder.POST("/reject", adminShopH.AdminRejectOrder)
	}
	adminShopRedeem := admin.Group("/shop/redeem")
	{
		adminShopRedeem.POST("/list", adminShopH.AdminListRedeemCodes)
	}
}
