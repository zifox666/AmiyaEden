package jobs

import (
	"amiya-eden/global"
	"amiya-eden/internal/repository"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/eve/esi"
	"context"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// esiQueue 全局 ESI 刷新队列实例
var esiQueue *esi.Queue

// GetESIQueue 获取 ESI 刷新队列实例（供 handler 层使用）
func GetESIQueue() *esi.Queue {
	return esiQueue
}

// registerESIRefreshJob 注册 ESI 数据刷新定时任务
func registerESIRefreshJob(c *cron.Cron) {
	esiQueue = esi.NewQueue(
		service.NewEveSSOService(),
		repository.NewEveCharacterRepository(),
	)

	rollSvc := service.NewRoleService()
	autoRoleSvc := service.NewAutoRoleService()

	runSigninSecuritySync := func(characterID int64, userID uint) {
		ctx := context.Background()

		// 先刷新 affiliation，确保 corporation_id 是当前值。
		if err := esiQueue.RunTask("character_affiliation", characterID); err != nil {
			global.Logger.Warn("[ESI SyncHook] affiliation 任务执行失败",
				zap.Int64("character_id", characterID),
				zap.Error(err),
			)
		}

		// 再刷新 corp roles，让 allow_corporations 过滤作用在最新军团上。
		if err := esiQueue.RunTask("character_corp_roles", characterID); err != nil {
			global.Logger.Warn("[ESI SyncHook] corp roles 任务执行失败",
				zap.Int64("character_id", characterID),
				zap.Error(err),
			)
		}

		if err := rollSvc.CheckCorpAccessAndAdjustRole(ctx, userID); err != nil {
			global.Logger.Warn("[ESI SyncHook] 权限检查失败",
				zap.Int64("character_id", characterID),
				zap.Uint("user_id", userID),
				zap.Error(err),
			)
		}
		if err := autoRoleSvc.SyncUserAutoRoles(ctx, userID); err != nil {
			global.Logger.Warn("[ESI SyncHook] 自动权限同步失败",
				zap.Int64("character_id", characterID),
				zap.Uint("user_id", userID),
				zap.Error(err),
			)
		}
	}

	// 注入同步钩子：在 JWT 生成前同步拉取最小安全数据并重算权限
	service.OnNewCharacterSyncFunc = func(characterID int64, userID uint) {
		runSigninSecuritySync(characterID, userID)
	}

	// 注入新角色全量刷新钩子：SSO 回调完成后后台异步执行，跑全部 ESI 任务，完成后补一次军团准入检查 + 自动权限同步
	service.OnNewCharacterFunc = func(characterID int64, userID uint) {
		ctx := context.Background()
		esiQueue.RunAllForCharacter(ctx, characterID)
		if err := rollSvc.CheckCorpAccessAndAdjustRole(ctx, userID); err != nil {
			global.Logger.Warn("[ESI FullRefreshHook] 权限检查失败",
				zap.Int64("character_id", characterID),
				zap.Uint("user_id", userID),
				zap.Error(err),
			)
		}
		// ESI 全量刷新完成后同步自动权限（corp_roles + titles 已入库）
		_ = autoRoleSvc.SyncUserAutoRoles(ctx, userID)
	}

	// 注入已有角色绑定/重登录同步钩子：JWT 生成前先刷新 affiliation / corp roles，再重算权限
	service.OnExistingCharacterSyncFunc = func(characterID int64, userID uint) {
		runSigninSecuritySync(characterID, userID)
	}

	// 注入舰队 PAP 发放时的 KM 刷新触发钩子
	service.FleetKMRefreshFunc = func(characterID int64) {
		if err := esiQueue.RunTask("character_killmails", characterID); err != nil {
			global.Logger.Warn("[Fleet KM] 触发 KM 刷新失败",
				zap.Int64("character_id", characterID),
				zap.Error(err),
			)
		}
	}

	// 注入自动 SRP 处理钩子
	autoSrpSvc := service.NewAutoSrpService()
	service.FleetAutoSRPFunc = func(fleetID string) {
		autoSrpSvc.ProcessAutoSRP(fleetID)
	}

	// 每 5 分钟执行一次调度（队列内部根据各任务间隔判断是否需要刷新）
	id, err := c.AddFunc("0 */5 * * * *", func() {
		esiQueue.Run()
	})
	if err != nil {
		global.Logger.Error("注册 ESI 刷新定时任务失败", zap.Error(err))
		return
	}
	global.Logger.Info("注册 ESI 刷新定时任务成功", zap.Int("entry_id", int(id)))
}
