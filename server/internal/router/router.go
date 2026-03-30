package router

import (
	"amiya-eden/internal/handler"
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/model"

	"github.com/gin-gonic/gin"
)

var (
	srpManageRoles = []string{model.RoleSRP, model.RoleFC, model.RoleAdmin}
	srpPayoutRoles = []string{model.RoleSRP, model.RoleAdmin}
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

	// ─── SDE 公开查询 ──
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

	// 福利（用户端 + 管理端共用 handler）
	welfareH := handler.NewWelfareHandler()

	// SSO 人物管理（绑定/解绑/设主人物）
	// guest 也应可访问，用于完成初次登录后的人物管理与补充授权。
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
	badgeH := handler.NewBadgeHandler()
	login.GET("/badge-counts", badgeH.GetBadgeCounts)

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

	// ─── 舰队 ───
	fleetH := handler.NewFleetHandler()
	operation := login.Group("/operation")
	fleet := operation.Group("/fleets")
	{
		manageFleets := middleware.RequireRole(model.RoleAdmin, model.RoleFC, model.RoleSeniorFC)
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

		// 查人物所在舰队
		fleet.GET("/esi/:character_id", fleetH.GetCharacterFleetInfo)

		// Webhook Ping（FC 或管理员手动触发）
		fleet.POST("/:id/ping", manageFleets, fleetH.PingFleet)
	}

	// ─── 舰队配置 ───
	fleetConfigH := handler.NewFleetConfigHandler()
	fleetConfig := operation.Group("/fleet-configs")
	{
		viewFleetConfigs := middleware.RequireLoginUser()
		manageFleetConfigs := middleware.RequireRole(model.RoleAdmin, model.RoleSeniorFC)

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
		viewSkillPlans := middleware.RequireRole(model.RoleAdmin, model.RoleSeniorFC, model.RoleFC)
		manageSkillPlans := middleware.RequireRole(model.RoleAdmin, model.RoleSeniorFC)
		viewSkillPlanChecks := middleware.RequireLoginUser()

		skillPlan.GET("/check/selection", viewSkillPlanChecks, skillPlanH.GetCheckSelection)
		skillPlan.PUT("/check/selection", viewSkillPlanChecks, skillPlanH.SaveCheckSelection)
		skillPlan.GET("/check/plan-selection", viewSkillPlanChecks, skillPlanH.GetCheckPlanSelection)
		skillPlan.PUT("/check/plan-selection", viewSkillPlanChecks, skillPlanH.SaveCheckPlanSelection)
		skillPlan.POST("/check/run", viewSkillPlanChecks, skillPlanH.RunCompletionCheck)
		skillPlan.GET("", viewSkillPlans, skillPlanH.ListSkillPlans)
		skillPlan.GET("/:id", viewSkillPlans, skillPlanH.GetSkillPlan)
		skillPlan.POST("", manageSkillPlans, skillPlanH.CreateSkillPlan)
		skillPlan.PUT("/reorder", manageSkillPlans, skillPlanH.ReorderSkillPlans)
		skillPlan.PUT("/:id", manageSkillPlans, skillPlanH.UpdateSkillPlan)
		skillPlan.DELETE("/:id", manageSkillPlans, skillPlanH.DeleteSkillPlan)
	}

	// ─── EVE 人物信息 ───
	infoH := handler.NewEveInfoHandler()
	info := login.Group("/info")
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

	// ─── 新人帮扶（用户/队长） ───
	newbroUserH := handler.NewNewbroUserHandler()
	newbro := login.Group("/newbro")
	{
		newbro.GET("/captains", newbroUserH.ListCaptains)
		newbro.GET("/affiliation/me", newbroUserH.GetMyAffiliation)
		newbro.GET("/affiliations/history", newbroUserH.ListMyAffiliationHistory)
		newbro.POST("/affiliation/select", newbroUserH.SelectCaptain)
		newbro.POST("/affiliation/end", newbroUserH.EndAffiliation)
	}

	newbroCaptainH := handler.NewNewbroCaptainHandler()
	newbroCaptain := login.Group("/newbro/captain", middleware.RequireRole(model.RoleCaptain))
	{
		newbroCaptain.GET("/overview", newbroCaptainH.GetOverview)
		newbroCaptain.GET("/players", newbroCaptainH.GetPlayers)
		newbroCaptain.GET("/attributions", newbroCaptainH.GetAttributions)
		newbroCaptain.GET("/rewards", newbroCaptainH.GetRewardSettlements)
		newbroCaptain.GET("/eligible-players", newbroCaptainH.ListEligiblePlayers)
		newbroCaptain.POST("/enroll", newbroCaptainH.EnrollPlayer)
		newbroCaptain.POST("/affiliation/end", newbroCaptainH.EndAffiliation)
	}

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
		srp.POST("/prices", middleware.RequireRole(model.RoleSRP), srpH.UpsertShipPrice)
		srp.DELETE("/prices/:id", middleware.RequireRole(model.RoleSRP), srpH.DeleteShipPrice)

		// 个人申请
		srp.POST("/applications", srpH.SubmitApplication)
		srp.GET("/applications/me", srpH.ListMyApplications)
		srp.GET("/killmails/me", srpH.GetMyKillmails)
		srp.GET("/killmails/fleet/:fleet_id", srpH.GetFleetKillmails)
		srp.POST("/killmails/detail", srpH.GetKillmailDetail)
		srp.POST("/open-info-window", srpH.OpenInfoWindow)

		// 审核（srp / fc / admin 可查看列表 / 详情 / 审批；发放和自动审批允许 srp / admin）
		reviewSRP := middleware.RequireRole(srpManageRoles...)
		payoutSRP := middleware.RequireRole(srpPayoutRoles...)
		srp.GET("/applications", reviewSRP, srpH.ListApplications)
		srp.GET("/applications/:id", reviewSRP, srpH.GetApplication)
		srp.PUT("/applications/:id/review", reviewSRP, srpH.ReviewApplication)
		srp.PUT("/applications/auto-approve", payoutSRP, srpH.RunFleetAutoApproval)
		srp.GET("/applications/batch-payout-summary", payoutSRP, srpH.ListBatchPayoutSummary)
		srp.PUT("/applications/fuxi-payout", payoutSRP, srpH.BatchPayoutAsFuxiCoin)
		srp.PUT("/applications/:id/payout", payoutSRP, srpH.Payout)
		srp.PUT("/applications/users/:user_id/payout", payoutSRP, srpH.BatchPayoutByUser)
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

	// ─── 系统管理（需要 admin 职权）───
	admin := login.Group("/system", middleware.RequireRole(model.RoleAdmin))

	// 系统基础配置
	sysConfigH := handler.NewSysConfigHandler()
	admin.GET("/basic-config", sysConfigH.GetBasicConfig)

	// SDE 配置管理
	admin.GET("/sde-config", sysConfigH.GetSDEConfig)
	admin.PUT("/sde-config", sysConfigH.UpdateSDEConfig)

	// 允许访问的军团列表
	admin.GET("/basic-config/allow-corporations", sysConfigH.GetAllowCorporations)
	admin.PUT("/basic-config/allow-corporations", sysConfigH.UpdateAllowCorporations)

	// NPC 刷怪报表（管理员 — 公司级）
	admin.POST("/npc-kills", npcKillH.GetCorpNpcKills)

	// 联盟 PAP 管理（管理员）
	alliancePAPAdminH := handler.NewAlliancePAPHandler()
	alliancePAPAdmin := admin.Group("/pap")
	{
		alliancePAPAdmin.GET("", alliancePAPAdminH.GetAllAlliancePAP)
		alliancePAPAdmin.POST("/fetch", alliancePAPAdminH.TriggerFetch)
		alliancePAPAdmin.POST("/import", alliancePAPAdminH.ImportAlliancePAP)
		// 月度归档
		alliancePAPAdmin.POST("/settle", alliancePAPAdminH.SettleMonth)
	}

	// PAP 兑换汇率管理（管理员）
	papExchangeH := handler.NewPAPExchangeHandler()
	admin.GET("/pap-exchange/rates", papExchangeH.GetRates)
	admin.PUT("/pap-exchange/rates", papExchangeH.SetRates)

	// 职权定义（只读）
	roleH := handler.NewRoleHandler()
	admin.GET("/role/definitions", roleH.ListRoleDefinitions)

	// 用户管理
	userH := handler.NewUserHandler()
	adminUser := admin.Group("/user")
	{
		adminUser.GET("", userH.ListUsers)
		adminUser.GET("/:id", userH.GetUser)
		adminUser.PUT("/:id", userH.UpdateUser)
		adminUser.DELETE("/:id", userH.DeleteUser)

		// 用户职权分配
		adminUser.GET("/:id/roles", roleH.GetUserRoles)
		adminUser.PUT("/:id/roles", roleH.SetUserRoles)

		// 模拟登录（仅超级管理员）
		adminUser.POST("/:id/impersonate", middleware.RequireRole(model.RoleSuperAdmin), userH.ImpersonateUser)
	}

	newbroAdminH := handler.NewNewbroAdminHandler()
	adminNewbro := admin.Group("/newbro")
	{
		adminNewbro.GET("/settings", newbroAdminH.GetSettings)
		adminNewbro.PUT("/settings", newbroAdminH.UpdateSettings)
		adminNewbro.GET("/captains", newbroAdminH.ListCaptains)
		adminNewbro.GET("/captains/:user_id", newbroAdminH.GetCaptainDetail)
		adminNewbro.GET("/affiliations/history", newbroAdminH.ListAffiliationHistory)
		adminNewbro.GET("/rewards", newbroAdminH.ListRewardSettlements)
		adminNewbro.POST("/attribution/sync", newbroAdminH.RunAttributionSync)
		adminNewbro.POST("/reward/process", newbroAdminH.RunRewardProcessing)
	}

	// 伏羲币管理（管理员）
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
	adminShopRedeem := admin.Group("/shop/redeem")
	{
		adminShopRedeem.POST("/list", adminShopH.AdminListRedeemCodes)
	}

	// 商店订单（管理员 / 福利官）
	shopOrder := login.Group("/system/shop/order", middleware.RequireRole(model.RoleAdmin, model.RoleWelfare))
	{
		shopOrder.POST("/list", adminShopH.AdminListOrders)
		shopOrder.POST("/deliver", adminShopH.AdminDeliverOrder)
		shopOrder.POST("/reject", adminShopH.AdminRejectOrder)
	}

	// 福利管理（列表：admin + welfare 可读；写操作仅 admin）
	welfareListGroup := login.Group("/system/welfare", middleware.RequireRole(model.RoleAdmin, model.RoleWelfare))
	welfareListGroup.POST("/list", welfareH.AdminListWelfares)

	adminWelfare := admin.Group("/welfare")
	{
		adminWelfare.POST("/add", welfareH.AdminCreateWelfare)
		adminWelfare.POST("/edit", welfareH.AdminUpdateWelfare)
		adminWelfare.POST("/delete", welfareH.AdminDeleteWelfare)
		adminWelfare.POST("/applications", welfareH.AdminListApplications)
		adminWelfare.POST("/applications/delete", welfareH.AdminDeleteApplication)
		adminWelfare.POST("/review", welfareH.AdminReviewApplication)
		adminWelfare.POST("/import", welfareH.AdminImportRecords)
		adminWelfare.POST("/reorder", welfareH.AdminReorderWelfares)
	}

	// ─── 用户端福利 ───
	welfareUser := login.Group("/welfare")
	{
		welfareUser.POST("/eligible", welfareH.GetEligibleWelfares)
		welfareUser.POST("/apply", welfareH.ApplyForWelfare)
		welfareUser.POST("/my-applications", welfareH.ListMyApplications)
		welfareUser.POST("/upload-evidence", welfareH.UploadEvidence)
	}

	// 自动权限映射管理（管理员）
	autoRoleH := handler.NewAutoRoleHandler()
	adminAutoRole := admin.Group("/auto-role")
	{
		// ESI 军团职权映射
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
