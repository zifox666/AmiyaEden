package jobs

import (
	"amiya-eden/global"
	"amiya-eden/internal/service"
	"context"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// RegisterAutoRoleJobs 注册自动权限同步定时任务
func RegisterAutoRoleJobs(c *cron.Cron) {
	// 每 10 分钟执行一次自动权限同步（在 ESI 刷新之后）
	id, err := c.AddFunc("0 2/10 * * * ?", autoRoleSyncTask)
	if err != nil {
		global.Logger.Error("注册自动权限同步定时任务失败", zap.Error(err))
		return
	}
	global.Logger.Info("注册自动权限同步定时任务成功", zap.Int("entry_id", int(id)))
}

// autoRoleSyncTask 根据 ESI 军团职权 + 头衔映射，自动同步所有用户权限
func autoRoleSyncTask() {
	ctx := context.Background()
	autoRoleSvc := service.NewAutoRoleService()
	autoRoleSvc.SyncAllUsersAutoRoles(ctx)
}
