package router

import (
	"amiya-eden/internal/handler"
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/model"
	"amiya-eden/pkg/response"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes is the AT-mode route allowlist. Legacy handlers remain in
// source but are intentionally never attached to the Gin engine.
func RegisterRoutes(r *gin.Engine) {
	r.Static("/uploads", "./uploads")
	r.NoRoute(func(c *gin.Context) { response.Fail(c, response.CodeNotFound, "资源不存在") })
	r.GET("/health", handler.NewHealthHandler().Check)
	api := r.Group("/api/v1")

	eveSSO := handler.NewEveSSOHandler()
	eue := api.Group("/sso/eve")
	{
		eue.GET("/login", eveSSO.Login)
		eue.GET("/callback", eveSSO.Callback)
		eue.GET("/scopes", eveSSO.GetScopes)
	}
	seatSSO := handler.NewSeatSSOHandler()
	seat := api.Group("/sso/seat")
	{
		seat.GET("/enabled", seatSSO.Enabled)
		seat.GET("/login", seatSSO.Login)
		seat.GET("/callback", seatSSO.Callback)
	}
	// The login shell needs the public site title and corporation logo. Only
	// these two display fields are exposed; updates remain admin-only below.
	sysConfigH := handler.NewSysConfigHandler()
	api.GET("/system/basic-config", sysConfigH.GetBasicConfig)
	recruitmentH := handler.NewRecruitmentHandler()
	api.GET("/recruitment/public/:token", recruitmentH.PublicContext)
	sdeH := handler.NewSdeHandler()
	sde := api.Group("/sde")
	{
		sde.GET("/version", sdeH.GetVersion)
		sde.POST("/types", sdeH.GetTypes)
		sde.POST("/names", sdeH.GetNames)
		sde.POST("/search", sdeH.FuzzySearch)
	}

	auth := api.Group("", middleware.JWTAuth())
	eueAuth := auth.Group("/sso/eve")
	{
		eueAuth.GET("/characters", eveSSO.GetMyCharacters)
		eueAuth.GET("/bind", eveSSO.BindLogin)
		eueAuth.POST("/transfer-confirm", eveSSO.TransferConfirm)
		eueAuth.PUT("/primary/:character_id", eveSSO.SetPrimary)
		eueAuth.DELETE("/characters/:character_id", eveSSO.Unbind)
	}
	seatAuth := auth.Group("/sso/seat")
	{
		seatAuth.GET("/bind", seatSSO.Bind)
		seatAuth.GET("/binding", seatSSO.GetSeatBinding)
		seatAuth.DELETE("/binding", seatSSO.Unbind)
	}

	meH := handler.NewMeHandler()
	auth.GET("/me", meH.GetMe)
	profileH := handler.NewProfileHandler()
	auth.GET("/profile", profileH.Get)
	auth.PUT("/profile", profileH.Update)
	uploadH := handler.NewUploadHandler()
	auth.POST("/recruitment/images", uploadH.UploadImage)
	menuH := handler.NewMenuHandler()
	auth.GET("/menu/list", menuH.GetMenuList)
	combatLogH := handler.NewCombatLogHandler()
	combatLogs := auth.Group("/combat-logs", middleware.RequireRole(model.RoleUser, model.RoleFC, model.RoleRecruiter, model.RoleAdmin, model.RoleSuperAdmin))
	{
		combatLogs.POST("", combatLogH.Upload)
		combatLogs.GET("", combatLogH.List)
		combatLogs.GET("/:id", combatLogH.Get)
	}

	// Only the four retained character data families are readable in AT mode.
	infoH := handler.NewEveInfoHandler()
	info := auth.Group("/info")
	{
		info.POST("/skills", infoH.GetCharacterSkills)
		info.POST("/implants", infoH.GetCharacterImplants)
		fittingsH := handler.NewFittingsHandler()
		info.POST("/fittings", fittingsH.GetFittings)
	}

	// Skill planning remains part of the retained main-character skill workflow.
	skillPlanH := handler.NewSkillPlanHandler()
	skillPlan := auth.Group("/operation/skill-plans")
	{
		skillPlan.GET("/all", skillPlanH.ListAllSkillPlans)
		skillPlan.GET("/:id", skillPlanH.GetSkillPlan)
		skillPlan.GET("/:id/check/me", skillPlanH.CheckUserCharacters)
		manage := skillPlan.Group("", middleware.RequireRole(model.RoleFC, model.RoleAdmin, model.RoleSuperAdmin))
		manage.GET("", skillPlanH.ListSkillPlans)
		manage.POST("", skillPlanH.CreateSkillPlan)
		manage.PUT("/:id", skillPlanH.UpdateSkillPlan)
		manage.DELETE("/:id", skillPlanH.DeleteSkillPlan)
		manage.GET("/:id/check", skillPlanH.CheckAllCharacters)
	}

	auth.POST("/recruitment/public/:token/applications", recruitmentH.Submit)
	auth.GET("/recruitment/my-applications", recruitmentH.ListMyApplications)
	auth.PUT("/recruitment/applications/:id/withdraw", recruitmentH.Withdraw)
	recruitment := auth.Group("/recruitment", middleware.RequireRole(model.RoleRecruiter, model.RoleAdmin, model.RoleSuperAdmin))
	{
		recruitment.GET("/questionnaires", recruitmentH.ListQuestionnaires)
		recruitment.POST("/questionnaires", recruitmentH.CreateQuestionnaire)
		recruitment.GET("/questionnaires/:id", recruitmentH.GetQuestionnaire)
		recruitment.PUT("/questionnaires/:id", recruitmentH.UpdateQuestionnaire)
		recruitment.GET("/links", recruitmentH.ListLinks)
		recruitment.POST("/links", recruitmentH.CreateLink)
		recruitment.PUT("/links/:id/active", recruitmentH.SetLinkActive)
		recruitment.GET("/applications", recruitmentH.ListApplications)
		recruitment.GET("/applications/:id", recruitmentH.GetApplication)
		recruitment.PUT("/applications/:id/review", recruitmentH.Review)
	}
	auth.POST("/recruitment/privacy-cleanup", middleware.RequireRole(model.RoleAdmin, model.RoleSuperAdmin), recruitmentH.Cleanup)

	registerATRoutes(auth)

	// Retained platform administration: RBAC, settings, SeAT, SDE and the
	// restricted ESI refresh queue.
	esiH := handler.NewESIRefreshHandler()
	esiRefresh := auth.Group("/esi/refresh", middleware.RequireRole(model.RoleAdmin, model.RoleSuperAdmin))
	{
		esiRefresh.GET("/tasks", esiH.GetTasks)
		esiRefresh.GET("/statuses", esiH.GetStatuses)
		esiRefresh.POST("/run", esiH.RunTask)
		esiRefresh.POST("/run-task", esiH.RunTaskByName)
		esiRefresh.POST("/run-all", esiH.RunAll)
	}
	admin := auth.Group("/system", middleware.RequireRole(model.RoleAdmin, model.RoleSuperAdmin))
	admin.PUT("/basic-config", sysConfigH.UpdateBasicConfig)
	admin.GET("/seat-config", seatSSO.GetSeatConfig)
	admin.PUT("/seat-config", seatSSO.UpdateSeatConfig)
	adminSDE := admin.Group("/sde")
	{
		adminSDE.GET("/version", sdeH.GetVersion)
		adminSDE.POST("/update", sdeH.TriggerUpdate)
	}
	adminMenu := admin.Group("/menu")
	{
		adminMenu.GET("/tree", menuH.GetMenuTree)
		adminMenu.POST("", menuH.CreateMenu)
		adminMenu.PUT("/:id", menuH.UpdateMenu)
		adminMenu.DELETE("/:id", menuH.DeleteMenu)
	}
	roleH := handler.NewRoleHandler()
	adminRole := admin.Group("/role")
	{
		adminRole.GET("", roleH.ListRoles)
		adminRole.GET("/all", roleH.ListAllRoles)
		adminRole.GET("/:id", roleH.GetRole)
		adminRole.POST("", roleH.CreateRole)
		adminRole.PUT("/:id", roleH.UpdateRole)
		adminRole.DELETE("/:id", roleH.DeleteRole)
		adminRole.GET("/:id/menus", roleH.GetRoleMenus)
		adminRole.PUT("/:id/menus", roleH.SetRoleMenus)
	}
	userH := handler.NewUserHandler()
	adminUser := admin.Group("/user")
	{
		adminUser.GET("", userH.ListUsers)
		adminUser.GET("/:id", userH.GetUser)
		adminUser.PUT("/:id", userH.UpdateUser)
		adminUser.DELETE("/:id", userH.DeleteUser)
		adminUser.GET("/:id/roles", roleH.GetUserRoles)
		adminUser.PUT("/:id/roles", roleH.SetUserRoles)
		adminUser.POST("/:id/impersonate", middleware.RequireRole(model.RoleSuperAdmin), userH.ImpersonateUser)
	}
}

func registerATRoutes(auth *gin.RouterGroup) {
	atScoreH := handler.NewATScoreHandler()
	atScore := auth.Group("/at/score-rules")
	{
		atScore.GET("", atScoreH.List)
		atScore.GET("/:id", atScoreH.Get)
		manage := atScore.Group("", middleware.RequireRole(model.RoleAdmin, model.RoleSuperAdmin))
		manage.POST("/import", atScoreH.Import)
		manage.POST("/:id/publish", atScoreH.Publish)
		manage.POST("/validate", atScoreH.Validate)
	}
	atSetH := handler.NewATSetHandler()
	atSets := auth.Group("/at/sets")
	{
		atSets.GET("", atSetH.List)
		atSets.GET("/:id", atSetH.Get)
		atSets.GET("/:id/eft", atSetH.EFT)
		manage := atSets.Group("", middleware.RequireRole(model.RoleFC, model.RoleAdmin, model.RoleSuperAdmin))
		manage.POST("/parse-eft", atSetH.ParseEFT)
		manage.POST("", atSetH.Create)
		manage.PUT("/:id", atSetH.Update)
		manage.DELETE("/:id", atSetH.Delete)
	}
	atFits := auth.Group("/at/fits")
	{
		atFits.GET("", atSetH.ListFits)
		atFits.GET("/:id", atSetH.GetFit)
		manage := atFits.Group("", middleware.RequireRole(model.RoleFC, model.RoleAdmin, model.RoleSuperAdmin))
		manage.POST("", atSetH.SaveFit)
		manage.PUT("/:id", atSetH.SaveFit)
		manage.DELETE("/:id", atSetH.DeleteFit)
	}
	atMatchH := handler.NewATMatchHandler()
	atMatches := auth.Group("/at/matches")
	{
		atMatches.GET("", atMatchH.List)
		atMatches.GET("/by-uuid/:uuid", atMatchH.GetByUUID)
		atMatches.GET("/:id", atMatchH.Get)
		atMatches.POST("/:id/register", atMatchH.Register)
		atMatches.POST("/:id/positions/self", atMatchH.SavePosition)
		atMatches.GET("/:id/positions/:slot/eft", atMatchH.PositionEFT)
		manage := atMatches.Group("", middleware.RequireRole(model.RoleFC, model.RoleAdmin, model.RoleSuperAdmin))
		manage.GET("/users/search", atMatchH.SearchContributionUsers)
		manage.POST("", atMatchH.Create)
		manage.PUT("/:id/status", atMatchH.Transition)
		manage.PUT("/:id/registrations", atMatchH.DecideRegistration)
		manage.PUT("/:id/positions", atMatchH.AssignPosition)
		manage.POST("/:id/rounds", atMatchH.CreateRound)
		manage.GET("/rounds/:id/roster", atMatchH.GetRoster)
		manage.PUT("/rounds/:id/roster", atMatchH.SetRoster)
		manage.POST("/rounds/:id/finish", atMatchH.FinishRound)
		manage.POST("/rounds/:id/reopen", atMatchH.ReopenRound)
		manage.POST("/rounds/:id/confirm", atMatchH.ConfirmRound)
		atMatches.POST("/:id/reset", middleware.RequireRole(model.RoleAdmin, model.RoleSuperAdmin), atMatchH.Reset)
	}
	atKMH := handler.NewATKillmailHandler()
	atKM := auth.Group("/at/killmails")
	{
		atKM.POST("/submit", middleware.RequirePermission("at:km:submit"), atKMH.Submit)
		atKM.GET("", atKMH.List)
		review := atKM.Group("", middleware.RequireRole(model.RoleFC, model.RoleAdmin, model.RoleSuperAdmin))
		review.GET("/pending", atKMH.Pending)
		review.GET("/:id", atKMH.Get)
		review.GET("/:id/suggestions", atKMH.Suggestions)
		review.POST("/:id/assign", atKMH.Assign)
		review.POST("/rounds/:id/resolve-conflict", atKMH.Resolve)
	}
	atBRH := handler.NewATBRHandler()
	atBR := auth.Group("/at/br")
	{
		atBR.GET("", atBRH.List)
		atBR.GET("/:id", atBRH.Get)
		atBR.GET("/rounds/:id/latest", atBRH.Latest)
		atBR.POST("/rounds/:id/build", middleware.RequireRole(model.RoleFC, model.RoleAdmin, model.RoleSuperAdmin), atBRH.Build)
	}
	atContributions := auth.Group("/at/contributions")
	{
		atContributions.GET("/seasons", atMatchH.ListSeasons)
		atContributions.GET("/seasons/:id/cleanup-ships", atMatchH.ListCleanup)
		atContributions.GET("/match-types", atMatchH.ListMatchTypes)
		atContributions.GET("/leaderboard", atMatchH.Leaderboard)
		atContributions.GET("/me", atMatchH.MyContributionLedger)
		admin := atContributions.Group("", middleware.RequireRole(model.RoleAdmin, model.RoleSuperAdmin))
		admin.POST("/seasons", atMatchH.SaveSeason)
		admin.PUT("/seasons/:id", atMatchH.SaveSeason)
		admin.PUT("/seasons/:id/cleanup-ships", atMatchH.SetCleanup)
		admin.POST("/match-types", atMatchH.SaveMatchType)
		admin.PUT("/match-types/:id", atMatchH.SaveMatchType)
		admin.GET("/users/search", atMatchH.SearchContributionUsers)
		admin.POST("/manual", atMatchH.ManualContribution)
	}
}
