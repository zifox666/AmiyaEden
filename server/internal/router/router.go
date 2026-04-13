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

	// ─── SeAT SSO（无需认证）───
	seatH := handler.NewSeatSSOHandler()
	seatSSO := api.Group("/sso/seat")
	{
		seatSSO.GET("/enabled", seatH.Enabled)
		seatSSO.GET("/login", seatH.Login)
		seatSSO.GET("/callback", seatH.Callback)
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

	// SSO 角色管理（绑定/解绑/设主角色）
	ssoAuth := auth.Group("/sso/eve")
	{
		// ssoAuth.GET("/scopes", ssoH.GetScopes)
		ssoAuth.GET("/characters", ssoH.GetMyCharacters)
		ssoAuth.GET("/bind", ssoH.BindLogin)
		ssoAuth.POST("/transfer-confirm", ssoH.TransferConfirm)
		ssoAuth.PUT("/primary/:character_id", ssoH.SetPrimary)
		ssoAuth.DELETE("/characters/:character_id", ssoH.Unbind)
	}

	// SeAT 账号管理（绑定/解绑/查看）
	seatAuth := auth.Group("/sso/seat")
	{
		seatAuth.GET("/bind", seatH.Bind)
		seatAuth.GET("/binding", seatH.GetSeatBinding)
		seatAuth.DELETE("/binding", seatH.Unbind)
	}

	// ─── 当前用户 ───
	meH := handler.NewMeHandler()
	auth.GET("/me", meH.GetMe)

	dashboardH := handler.NewDashboardHandler()
	auth.POST("/dashboard", dashboardH.GetDashboard)

	// ─── 通知 ───
	notifH := handler.NewNotificationHandler()
	notification := auth.Group("/notification")
	{
		notification.POST("/list", notifH.ListNotifications)
		notification.POST("/unread-count", notifH.GetUnreadCount)
		notification.POST("/read", notifH.MarkAsRead)
		notification.POST("/read-all", notifH.MarkAllAsRead)
	}

	// ─── 菜单 ───
	menuH := handler.NewMenuHandler()
	auth.GET("/menu/list", menuH.GetMenuList) // 当前用户可用菜单

	// ─── 舰队 ───
	fleetH := handler.NewFleetHandler()
	operation := auth.Group("/operation")
	fleet := operation.Group("/fleets")
	{
		// ─── 所有已认证用户可访问 ───
		fleet.GET("", fleetH.ListFleets)
		fleet.GET("/me", fleetH.GetMyFleets)
		fleet.GET("/pap/me", fleetH.GetMyPapLogs)
		fleet.GET("/:id", fleetH.GetFleet)
		fleet.GET("/:id/members", fleetH.GetMembers)
		fleet.GET("/:id/members-pap", fleetH.GetMembersWithPap)
		fleet.GET("/:id/pap", fleetH.GetPapLogs)
		fleet.POST("/join", fleetH.JoinFleet)
		fleet.GET("/esi/:character_id", fleetH.GetCharacterFleetInfo)

		// ――― 联盟 PAP（所有用户可查）
		alliancePAPH := handler.NewAlliancePAPHandler()
		fleet.GET("/pap/alliance", alliancePAPH.GetMyAlliancePAP)

		// ─── 仅 FC / 管理员可操作 ───
		fleetFC := fleet.Group("", middleware.RequireRole(model.RoleFC, model.RoleAdmin))
		{
			fleetFC.POST("", fleetH.CreateFleet)
			fleetFC.PUT("/:id", fleetH.UpdateFleet)
			fleetFC.DELETE("/:id", fleetH.DeleteFleet)
			fleetFC.POST("/:id/refresh-esi", fleetH.RefreshFleetESI)
			fleetFC.POST("/:id/members/sync", fleetH.SyncESIMembers)
			fleetFC.POST("/:id/pap", fleetH.IssuePap)
			fleetFC.POST("/:id/manual-pap", fleetH.ManualPap)
			fleetFC.POST("/:id/br", fleetH.GenerateBattleReport)
			fleetFC.POST("/:id/invites", fleetH.CreateInvite)
			fleetFC.GET("/:id/invites", fleetH.GetInvites)
			fleetFC.DELETE("/invites/:invite_id", fleetH.DeactivateInvite)
			fleetFC.POST("/:id/ping", fleetH.PingFleet)
		}
	}

	// ─── 舰队配置 ───
	fleetConfigH := handler.NewFleetConfigHandler()
	fleetConfig := operation.Group("/fleet-configs")
	{
		fleetConfig.GET("", fleetConfigH.ListFleetConfigs)
		fleetConfig.GET("/:id", fleetConfigH.GetFleetConfig)
		fleetConfig.GET("/:id/eft", fleetConfigH.GetFittingEFT)
		fleetConfig.POST("", middleware.RequireRole(model.RoleFC, model.RoleSRP), fleetConfigH.CreateFleetConfig)
		fleetConfig.PUT("/:id", middleware.RequireRole(model.RoleFC, model.RoleSRP), fleetConfigH.UpdateFleetConfig)
		fleetConfig.DELETE("/:id", middleware.RequireRole(model.RoleFC, model.RoleSRP), fleetConfigH.DeleteFleetConfig)
		fleetConfig.POST("/import-fitting", fleetConfigH.ImportFromUserFitting)
		fleetConfig.POST("/export-esi", fleetConfigH.ExportToESI)
		fleetConfig.GET("/:id/fittings/:fitting_id/items", fleetConfigH.GetFittingItems)
		fleetConfig.PUT("/:id/fittings/:fitting_id/items/settings", middleware.RequireRole(model.RoleFC, model.RoleSRP), fleetConfigH.UpdateFittingItemsSettings)
	}

	// ─── 技能规划 ───
	skillPlanH := handler.NewSkillPlanHandler()
	skillPlan := operation.Group("/skill-plans")

	// ─── 军团建筑管理 ───
	corpStructureH := handler.NewCorpStructureHandler()
	corpStructure := operation.Group("/corp-structures")
	{
		corpStructure.POST("/list", corpStructureH.ListStructures)
		corpStructure.GET("/corps", corpStructureH.GetCorpIDs)
	}
	{
		skillPlan.GET("/all", skillPlanH.ListAllSkillPlans)
		skillPlan.GET("/:id", skillPlanH.GetSkillPlan)
		skillPlan.GET("/:id/check/me", skillPlanH.CheckUserCharacters)
		// 管理操作（需要 FC 或 admin）
		skillPlan.GET("", middleware.RequireRole(model.RoleFC, model.RoleAdmin), skillPlanH.ListSkillPlans)
		skillPlan.POST("", middleware.RequireRole(model.RoleFC, model.RoleAdmin), skillPlanH.CreateSkillPlan)
		skillPlan.PUT("/:id", middleware.RequireRole(model.RoleFC, model.RoleAdmin), skillPlanH.UpdateSkillPlan)
		skillPlan.DELETE("/:id", middleware.RequireRole(model.RoleFC, model.RoleAdmin), skillPlanH.DeleteSkillPlan)
		skillPlan.GET("/:id/check", middleware.RequireRole(model.RoleFC, model.RoleAdmin), skillPlanH.CheckAllCharacters)
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

	// ─── 击杀邮件查询 ───
	killmailH := handler.NewKillmailHandler()
	info.POST("/killmails", killmailH.GetCharacterKillmails)

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
		// 抽奖
		lotteryH := handler.NewLotteryHandler()
		shop.POST("/lottery/list", lotteryH.ListActivities)
		shop.POST("/lottery/draw", lotteryH.Draw)
		shop.POST("/lottery/records", lotteryH.GetMyRecords)
	}

	// ─── 文件上传（需要登录）───
	uploadH := handler.NewUploadHandler()
	auth.POST("/upload/image", uploadH.UploadImage)

	// ─── SRP 补损 ───
	srpH := handler.NewSrpHandler()
	srp := auth.Group("/srp")
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

	// 系统基础配置
	sysConfigH := handler.NewSysConfigHandler()
	admin.GET("/basic-config", sysConfigH.GetBasicConfig)
	admin.PUT("/basic-config", sysConfigH.UpdateBasicConfig)

	// SeAT 配置（管理员）
	admin.GET("/seat-config", seatH.GetSeatConfig)
	admin.PUT("/seat-config", seatH.UpdateSeatConfig)

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
		adminShopOrder.POST("/ship", adminShopH.AdminShipOrder)
	}
	adminShopRedeem := admin.Group("/shop/redeem")
	{
		adminShopRedeem.POST("/list", adminShopH.AdminListRedeemCodes)
	}

	// 抽奖管理（管理员）
	adminLotteryH := handler.NewLotteryHandler()
	adminLottery := admin.Group("/shop/lottery")
	{
		adminLottery.POST("/list", adminLotteryH.AdminListActivities)
		adminLottery.POST("/add", adminLotteryH.AdminCreateActivity)
		adminLottery.POST("/edit", adminLotteryH.AdminUpdateActivity)
		adminLottery.POST("/delete", adminLotteryH.AdminDeleteActivity)
		adminLottery.POST("/prize/add", adminLotteryH.AdminCreatePrize)
		adminLottery.POST("/prize/edit", adminLotteryH.AdminUpdatePrize)
		adminLottery.POST("/prize/delete", adminLotteryH.AdminDeletePrize)
		adminLottery.POST("/records", adminLotteryH.AdminListRecords)
		adminLottery.POST("/records/deliver", adminLotteryH.AdminUpdateRecordDelivery)
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

		// SeAT 分组映射
		adminAutoRole.GET("/seat-roles", autoRoleH.GetAllSeatRoles)
		adminAutoRole.GET("/seat-role-mappings", autoRoleH.ListSeatRoleMappings)
		adminAutoRole.POST("/seat-role-mappings", autoRoleH.CreateSeatRoleMapping)
		adminAutoRole.DELETE("/seat-role-mappings/:id", autoRoleH.DeleteSeatRoleMapping)

		// 手动触发同步
		adminAutoRole.POST("/sync", autoRoleH.TriggerSync)

		// 操作日志
		adminAutoRole.GET("/logs", autoRoleH.ListAutoRoleLogs)

		// 准入名单管理
		adminAutoRole.GET("/allow-list/:type", autoRoleH.ListAllowedEntities)
		adminAutoRole.POST("/allow-list/:type", autoRoleH.AddAllowedEntity)
		adminAutoRole.DELETE("/allow-list/:type/:id", autoRoleH.RemoveAllowedEntity)

		// EVE 实体模糊搜索（zkillboard autocomplete 代理）
		adminAutoRole.GET("/eve-search", autoRoleH.SearchEveEntities)
	}

	// Webhook 配置（管理员）
	webhookH := handler.NewWebhookHandler()
	adminWebhook := admin.Group("/webhook")
	{
		adminWebhook.GET("/config", webhookH.GetConfig)
		adminWebhook.PUT("/config", webhookH.SetConfig)
		adminWebhook.POST("/test", webhookH.TestWebhook)
	}

	// SDE 数据管理（管理员）
	adminSde := admin.Group("/sde")
	{
		adminSde.GET("/version", sdeH.GetVersion)
		adminSde.POST("/update", sdeH.TriggerUpdate)
	}

	// ─── 军团管理（管理员） ───
	corpAdmin := auth.Group("/corp", middleware.RequireRole(model.RoleAdmin, model.RoleSuperAdmin))
	{
		incentiveH := handler.NewFleetBattleIncentiveHandler()
		corpIncentive := corpAdmin.Group("/battle-incentives")
		{
			corpIncentive.GET("", incentiveH.ListAll)
			corpIncentive.PUT("/:fleet_type", incentiveH.Update)
		}
		// 手动补发 FC 带队奖励
		corpAdmin.POST("/fleets/:id/lead-reward", incentiveH.IssueFCLeadReward)
	}
}
