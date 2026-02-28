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
	esiQueue = esi.NewQueue()

	charRepo := repository.NewEveCharacterRepository()
	rollSvc := service.NewRoleService()

	// 注入新角色全量刷新钩子：ESI 拉取完成后（corp_id 已写入）再做军团准入检查
	service.OnNewCharacterFunc = func(characterID int64) {
		ctx := context.Background()
		esiQueue.RunAllForCharacter(ctx, characterID)
		// ESI 刷新已写入 CorporationID，立即检查准入
		if char, err := charRepo.GetByCharacterID(characterID); err == nil {
			_ = rollSvc.CheckCorpAccessAndAdjustRole(ctx, char.UserID)
		}
	}

	// 注入已有角色绑定/重登录钩子：corp_id 已知，直接检查准入
	service.OnCharacterBindFunc = func(userID uint) {
		_ = rollSvc.CheckCorpAccessAndAdjustRole(context.Background(), userID)
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
