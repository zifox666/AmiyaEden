package jobs

import (
	"amiya-eden/global"
	"amiya-eden/pkg/eve/esi"

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
