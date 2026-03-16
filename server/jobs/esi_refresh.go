package jobs

import (
	"amiya-eden/global"
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
	esiQueue = esi.NewQueue()

	rollSvc := service.NewRoleService()
	autoRoleSvc := service.NewAutoRoleService()

	// 注入同步钩子：仅执行 affiliation 拉取 + 军团准入检查（在 JWT 生成前同步调用）
	service.OnNewCharacterSyncFunc = func(characterID int64, userID uint) {
		ctx := context.Background()
		// RunTask 内部同步执行，affiliation 为公开接口，速度快（~100ms）
		if err := esiQueue.RunTask("character_affiliation", characterID); err != nil {
			global.Logger.Warn("[ESI SyncHook] affiliation 任务执行失败",
				zap.Int64("character_id", characterID),
				zap.Error(err),
			)
		}
		_ = rollSvc.CheckCorpAccessAndAdjustRole(ctx, userID)
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

	// 注入已有角色绑定/重登录钩子：corp_id 已知，直接检查准入 + 自动权限同步
	service.OnCharacterBindFunc = func(userID uint) {
		ctx := context.Background()
		_ = rollSvc.CheckCorpAccessAndAdjustRole(ctx, userID)
		_ = autoRoleSvc.SyncUserAutoRoles(ctx, userID)
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
