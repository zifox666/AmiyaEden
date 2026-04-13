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

	// 每 30 分钟刷新 SeAT 用户分组并同步权限
	sid, err := c.AddFunc("0 5/30 * * * ?", seatRoleSyncTask)
	if err != nil {
		global.Logger.Error("注册 SeAT 分组同步定时任务失败", zap.Error(err))
		return
	}
	global.Logger.Info("注册 SeAT 分组同步定时任务成功", zap.Int("entry_id", int(sid)))
}

// autoRoleSyncTask 根据 ESI 军团角色 + 头衔映射，自动同步所有用户权限
func autoRoleSyncTask() {
	ctx := context.Background()
	autoRoleSvc := service.NewAutoRoleService()
	autoRoleSvc.SyncAllUsersAutoRoles(ctx)
}

// seatRoleSyncTask 刷新 SeAT 用户 token 和分组，然后同步自动权限
func seatRoleSyncTask() {
	ctx := context.Background()

	// 1. 刷新所有 SeAT 用户的 token 和 groups
	seatSvc := service.NewSeatSSOService()
	seatSvc.RefreshAllSeatUserGroups(ctx)

	// 2. 重新同步自动权限
	autoRoleSvc := service.NewAutoRoleService()
	autoRoleSvc.SyncAllUsersAutoRoles(ctx)
}
