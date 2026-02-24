package jobs

import (
	"amiya-eden/global"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// RegisterAll 统一注册所有定时任务
func RegisterAll(c *cron.Cron) {
	registerExampleJob(c)
	// registerCleanupJob(c)
	// registerReportJob(c)
}

// registerExampleJob 示例任务：每分钟执行一次
func registerExampleJob(c *cron.Cron) {
	id, err := c.AddFunc("0 * * * * *", exampleTask)
	if err != nil {
		global.Logger.Error("注册示例定时任务失败", zap.Error(err))
		return
	}
	global.Logger.Info("注册示例定时任务成功", zap.Int("entry_id", int(id)))
}

// exampleTask 示例任务执行逻辑
func exampleTask() {
	global.Logger.Info("[定时任务] 示例任务执行中...")

	// TODO: 在此编写业务逻辑
	// 例如：清理过期数据、发送统计报表、同步远程数据等

	global.Logger.Info("[定时任务] 示例任务执行完毕")
}
