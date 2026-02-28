package jobs

import (
	// 确保所有 ESI 刷新任务的 init() 被调用（任务自动注册）
	_ "amiya-eden/pkg/eve/esi"

	"github.com/robfig/cron/v3"
)

// RegisterAll 统一注册所有定时任务
func RegisterAll(c *cron.Cron) {
	registerSdeJob(c)
	registerESIRefreshJob(c)
	registerAlliancePAPJob(c)
	// registerCleanupJob(c)
	// registerReportJob(c)
}
