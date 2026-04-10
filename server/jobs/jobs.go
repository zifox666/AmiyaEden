package jobs

import (
	// 确保所有 ESI 刷新任务的 init() 被调用（任务自动注册）
	"amiya-eden/internal/taskregistry"
	_ "amiya-eden/pkg/eve/esi"
)

// RegisterAll 统一注册所有定时任务
func RegisterAll(reg *taskregistry.Registry) {
	registerESIRefreshTask(reg)
	registerAutoSrpTask(reg)
	registerAlliancePAPTasks(reg)
	registerNewbroSupportTasks(reg)
	registerCorpAccessCheckTask(reg)
	registerAutoRoleSyncTask(reg)
	registerTaskExecutionHistoryCleanupTask(reg)
	registerMentorRewardTask(reg)
	// registerCleanupJob(c)
	// registerReportJob(c)
}
