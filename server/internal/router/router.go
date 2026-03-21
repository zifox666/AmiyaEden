package router

import (
	"amiya-eden/internal/handler"
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/model"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册所有业务路由
func RegisterRoutes(r *gin.Engine) {
	// ─── 上传文件静态目录 ───
	r.Static("/uploads", "./uploads")

	api := r.Group("/api/v1")

	// ─── 无需认证 ───
	ssoH := handler.NewEveSSOHandler()
	sso := api.Group("/sso/eve")
	{
		sso.GET("/login", ssoH.Login)
		sso.GET("/callback", ssoH.Callback)
		sso.GET("/scopes", ssoH.GetScopes)
	}

	// ─── SDE 公开查询（API Key 鉴权）───
	sdeH := handler.NewSdeHandler()
	sde := api.Group("/sde")
	{
		sde.GET("/version", sdeH.GetVersion)
		sde.POST("/types", sdeH.GetTypes)
		sde.POST("/names", sdeH.GetNames)
		sde.POST("/search", sdeH.FuzzySearch)
	}

	// ─── 需要登录 ───
	auth := api.Group("", middleware.JWTAuth())
	login := auth.Group("", middleware.RequireLoginUser())

	// SSO 角色管理（绑定/解绑/设主角色）
	// guest 也应可访问，用于完成初次登录后的角色管理与补充授权。
	ssoAuth := auth.Group("/sso/eve")
	{
		// ssoAuth.GET("/scopes", ssoH.GetScopes)
		ssoAuth.GET("/characters", ssoH.GetMyCharacters)
		ssoAuth.GET("/bind", ssoH.BindLogin)
		ssoAuth.PUT("/primary/:character_id", ssoH.SetPrimary)
		ssoAuth.DELETE("/characters/:character_id", ssoH.Unbind)
	}

	// ─── 当前用户 ───
	meH := handler.NewMeHandler()
	auth.GET("/me", meH.GetMe)
	auth.PUT("/me", meH.UpdateMe)

	dashboardH := handler.NewDashboardHandler()
	auth.POST("/dashboard", dashboardH.GetDashboard)

	// ─── 通知 ───
	notifH := handler.NewNotificationHandler()
	notification := auth.Group("/notification")
	{
		notification.POST("/list", notifH.ListNotifications)
		notification.POST("/unread-count", notifH.GetUnreadCount)
	}
	notificationWrite := login.Group("/notification")
	{
		notificationWrite.POST("/read", notifH.MarkAsRead)
		notificationWrite.POST("/read-all", notifH.MarkAllAsRead)
	}

	// ─── 菜单 ───
	menuH := handler.NewMenuHandler()
	auth.GET("/menu/list", menuH.GetMenuList) // 当前用户可用菜单

	// ─── 舰队 ───
	fleetH := handler.NewFleetHandler()
	operation := login.Group("/operation")
	fleet := operation.Group("/fleets")
	{
		manageFleets := middleware.RequireRole(model.RoleAdmin, model.RoleFC)
		deleteFleets := middleware.RequireRole(model.RoleAdmin)

		fleet.POST("", manageFleets, fleetH.CreateFleet)
		fleet.GET("", manageFleets, fleetH.ListFleets)
		fleet.GET("/me", fleetH.GetMyFleets)
		fleet.GET("/:id", manageFleets, fleetH.GetFleet)
		fleet.PUT("/:id", manageFleets, fleetH.UpdateFleet)
		fleet.DELETE("/:id", deleteFleets, fleetH.DeleteFleet)
		fleet.POST("/:id/refresh-esi", manageFleets, fleetH.RefreshFleetESI)

		// 成员
		fleet.GET("/:id/members", manageFleets, fleetH.GetMembers)
		fleet.GET("/:id/members-pap", manageFleets, fleetH.GetMembersWithPap)
		fleet.POST("/:id/members/manual", manageFleets, fleetH.ManualAddMembers)
		fleet.POST("/:id/members/sync", manageFleets, fleetH.SyncESIMembers)

		// ――― PAP
		fleet.POST("/:id/pap", manageFleets, fleetH.IssuePap)
		fleet.GET("/:id/pap", manageFleets, fleetH.GetPapLogs)
		fleet.GET("/pap/me", fleetH.GetMyPapLogs)
		fleet.GET("/pap/corporation", fleetH.GetCorporationPapSummary)

		// ――― 联盟 PAP
		alliancePAPH := handler.NewAlliancePAPHandler()
		fleet.GET("/pap/alliance", alliancePAPH.GetMyAlliancePAP)

		// 邀请
		fleet.POST("/:id/invites", manageFleets, fleetH.CreateInvite)
		fleet.GET("/:id/invites", manageFleets, fleetH.GetInvites)
		fleet.DELETE("/invites/:invite_id", manageFleets, fleetH.DeactivateInvite)
		fleet.POST("/join", fleetH.JoinFleet)

		// 查角色所在舰队
		fleet.GET("/esi/:character_id", fleetH.GetCharacterFleetInfo)

		// Webhook Ping（FC 或管理员手动触发）
		fleet.POST("/:id/ping", manageFleets, fleetH.PingFleet)
	}

	// ─── 舰队配置 ───
	fleetConfigH := handler.NewFleetConfigHandler()
	fleetConfig := operation.Group("/fleet-configs")
	{
		viewFleetConfigs := middleware.RequireLoginUser()
		manageFleetConfigs := middleware.RequireRole(model.RoleAdmin, model.RoleFC)

		fleetConfig.GET("", viewFleetConfigs, fleetConfigH.ListFleetConfigs)
		fleetConfig.GET("/:id", viewFleetConfigs, fleetConfigH.GetFleetConfig)
		fleetConfig.GET("/:id/eft", viewFleetConfigs, fleetConfigH.GetFittingEFT)
		fleetConfig.POST("", manageFleetConfigs, fleetConfigH.CreateFleetConfig)
		fleetConfig.PUT("/:id", manageFleetConfigs, fleetConfigH.UpdateFleetConfig)
		fleetConfig.DELETE("/:id", manageFleetConfigs, fleetConfigH.DeleteFleetConfig)
		fleetConfig.POST("/import-fitting", manageFleetConfigs, fleetConfigH.ImportFromUserFitting)
		fleetConfig.POST("/export-esi", viewFleetConfigs, fleetConfigH.ExportToESI)
		fleetConfig.GET("/:id/fittings/:fitting_id/items", viewFleetConfigs, fleetConfigH.GetFittingItems)
		fleetConfig.PUT("/:id/fittings/:fitting_id/items/settings", manageFleetConfigs, fleetConfigH.UpdateFittingItemsSettings)
	}

	// ─── 军团技能计划 ───
	skillPlanH := handler.NewSkillPlanHandler()
	skillPlanning := login.Group("/skill-planning")
	skillPlan := skillPlanning.Group("/skill-plans")
	{
		manageSkillPlans := middleware.RequireRole(model.RoleAdmin, model.RoleFC)
		viewSkillPlanChecks := middleware.RequireLoginUser()

		skillPlan.GET("/check/selection", viewSkillPlanChecks, skillPlanH.GetCheckSelection)
		skillPlan.PUT("/check/selection", viewSkillPlanChecks, skillPlanH.SaveCheckSelection)
		skillPlan.POST("/check/run", viewSkillPlanChecks, skillPlanH.RunCompletionCheck)
		skillPlan.GET("", manageSkillPlans, skillPlanH.ListSkillPlans)
		skillPlan.GET("/:id", manageSkillPlans, skillPlanH.GetSkillPlan)
		skillPlan.POST("", manageSkillPlans, skillPlanH.CreateSkillPlan)
		skillPlan.PUT("/:id", manageSkillPlans, skillPlanH.UpdateSkillPlan)
		skillPlan.DELETE("/:id", manageSkillPlans, skillPlanH.DeleteSkillPlan)
	}

	// ─── EVE 角色信息 ───
	infoH := handler.NewEveInfoHandler()
	info := auth.Group("/info")
	{
		info.POST("/wallet", infoH.GetWalletJournal)
		info.POST("/skills", infoH.GetCharacterSkills)
		info.POST("/ships", infoH.GetCharacterShips)
		info.POST("/implants", infoH.GetCharacterImplants)
		info.POST("/assets", infoH.GetAssets)
		info.POST("/contracts", infoH.GetContracts)
		info.POST("/contracts/detail", infoH.GetContractDetail)
	}

	// ─── 装配 ───
	fittingsH := handler.NewFittingsHandler()
	info.POST("/fittings", fittingsH.GetFittings)
	info.POST("/fittings/save", fittingsH.SaveFitting)

	// ─── NPC 刷怪报表 ───
	npcKillH := handler.NewNpcKillHandler()
	info.POST("/npc-kills", npcKillH.GetNpcKills)
	info.POST("/npc-kills/all", npcKillH.GetAllNpcKills)

	// ─── 商店（用户端）───
	shopH := handler.NewShopHandler()
	shop := login.Group("/shop")
	walletH := handler.NewSysWalletHandler()
	shopWallet := shop.Group("/wallet")
	{
		shop.POST("/products", shopH.ListProducts)
		shop.POST("/product/detail", shopH.GetProductDetail)
		shop.POST("/buy", shopH.BuyProduct)
		shop.POST("/orders", shopH.GetMyOrders)
		shop.POST("/redeem/list", shopH.GetMyRedeemCodes)

		shopWallet.POST("/my", walletH.GetMyWallet)
		shopWallet.POST("/my/transactions", walletH.GetMyTransactions)
	}

	// ─── 文件上传（需要登录）───
	uploadH := handler.NewUploadHandler()
	login.POST("/upload/image", uploadH.UploadImage)

	// ─── SRP 补损 ───
	srpH := handler.NewSrpHandler()
	srp := login.Group("/srp")
	{
		// 价格表（查看公开，修改需权限）
		srp.GET("/prices", srpH.ListShipPrices)
		srp.POST("/prices", middleware.RequirePermission("srp:price:add"), srpH.UpsertShipPrice)
		srp.DELETE("/prices/:id", middleware.RequirePermission("srp:price:delete"), srpH.DeleteShipPrice)

		// 个人申请
		srp.POST("/applications", srpH.SubmitApplication)
		srp.GET("/applications/me", srpH.ListMyApplications)
		srp.GET("/killmails/me", srpH.GetMyKillmails)
		srp.GET("/killmails/fleet/:fleet_id", srpH.GetFleetKillmails)
		srp.POST("/killmails/detail", srpH.GetKillmailDetail)
		srp.POST("/open-info-window", srpH.OpenInfoWindow)

		// 审核（需权限）
		srpAdmin := srp.Group("", middleware.RequirePermission("srp:review"))
		{
			srpAdmin.GET("/applications", srpH.ListApplications)
			srpAdmin.GET("/applications/batch-payout-summary", srpH.ListBatchPayoutSummary)
			srpAdmin.GET("/applications/:id", srpH.GetApplication)
			srpAdmin.PUT("/applications/:id/review", srpH.ReviewApplication)
			srpAdmin.PUT("/applications/:id/payout", srpH.Payout)
			srpAdmin.PUT("/applications/users/:user_id/payout", srpH.BatchPayoutByUser)
		}
	}

	// ─── ESI 刷新队列 ───
	esiH := handler.NewESIRefreshHandler()
	esiRefresh := login.Group("/esi/refresh", middleware.RequireRole(model.RoleAdmin))
	{
		esiRefresh.GET("/tasks", esiH.GetTasks)
		esiRefresh.GET("/statuses", esiH.GetStatuses)
		esiRefresh.POST("/run", esiH.RunTask)
		esiRefresh.POST("/run-task", esiH.RunTaskByName)
		esiRefresh.POST("/run-all", esiH.RunAll)
	}

	// ─── 系统管理（需要 admin 角色）───
	admin := login.Group("/system", middleware.RequireRole(model.RoleAdmin))

	// 系统基础配置
	sysConfigH := handler.NewSysConfigHandler()
	admin.GET("/basic-config", sysConfigH.GetBasicConfig)
	admin.PUT("/basic-config", sysConfigH.UpdateBasicConfig)

	// NPC 刷怪报表（管理员 — 公司级）
	admin.POST("/npc-kills", npcKillH.GetCorpNpcKills)

	// 联盟 PAP 管理（管理员）
	alliancePAPAdminH := handler.NewAlliancePAPHandler()
	alliancePAPAdmin := admin.Group("/pap")
	{
		alliancePAPAdmin.GET("", alliancePAPAdminH.GetAllAlliancePAP)
		alliancePAPAdmin.POST("/fetch", alliancePAPAdminH.TriggerFetch)
		alliancePAPAdmin.POST("/import", alliancePAPAdminH.ImportAlliancePAP)
		// PAP 兑换配置
		alliancePAPAdmin.GET("/config", alliancePAPAdminH.GetExchangeConfig)
		alliancePAPAdmin.PUT("/config", alliancePAPAdminH.SetExchangeConfig)
		// 月度归档 + 兑换系统钱包
		alliancePAPAdmin.POST("/settle", alliancePAPAdminH.SettleMonth)
	}

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

		// 模拟登录（仅超级管理员）
		adminUser.POST("/:id/impersonate", middleware.RequireRole(model.RoleSuperAdmin), userH.ImpersonateUser)
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

	// 自动权限映射管理（管理员）
	autoRoleH := handler.NewAutoRoleHandler()
	adminAutoRole := admin.Group("/auto-role")
	{
		// ESI 军团角色映射
		adminAutoRole.GET("/esi-roles", autoRoleH.GetAllEsiRoles)
		adminAutoRole.GET("/esi-role-mappings", autoRoleH.ListEsiRoleMappings)
		adminAutoRole.POST("/esi-role-mappings", autoRoleH.CreateEsiRoleMapping)
		adminAutoRole.DELETE("/esi-role-mappings/:id", autoRoleH.DeleteEsiRoleMapping)

		// ESI 头衔映射
		adminAutoRole.GET("/corp-titles", autoRoleH.ListCorpTitles)
		adminAutoRole.GET("/esi-title-mappings", autoRoleH.ListEsiTitleMappings)
		adminAutoRole.POST("/esi-title-mappings", autoRoleH.CreateEsiTitleMapping)
		adminAutoRole.DELETE("/esi-title-mappings/:id", autoRoleH.DeleteEsiTitleMapping)

		// 手动触发同步
		adminAutoRole.POST("/sync", autoRoleH.TriggerSync)
	}

	// Webhook 配置（管理员）
	webhookH := handler.NewWebhookHandler()
	adminWebhook := admin.Group("/webhook")
	{
		adminWebhook.GET("/config", webhookH.GetConfig)
		adminWebhook.PUT("/config", webhookH.SetConfig)
		adminWebhook.POST("/test", webhookH.TestWebhook)
	}
}
