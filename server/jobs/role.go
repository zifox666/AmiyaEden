package jobs

import (
	"amiya-eden/global"
	"amiya-eden/internal/repository"
	"amiya-eden/internal/service"
	"context"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

func RegisterRoleJobs(c *cron.Cron) {
	id, err := c.AddFunc("0 0/30 * * * ?", roleCheckTask)
	if err != nil {
		global.Logger.Error("注册角色检查定时任务失败", zap.Error(err))
		return
	}
	global.Logger.Info("注册角色检查定时任务成功", zap.Int("entry_id", int(id)))
}

// roleCheckTask 遍历所有用户，根据军团准入列表调整用户权限
func roleCheckTask() {
	// 未配置允许军团列表时跳过
	if len(global.Config.App.AllowCorporations) == 0 {
		return
	}

	ctx := context.Background()
	userRepo := repository.NewUserRepository()
	rollSvc := service.NewRoleService()

	ids, err := userRepo.ListAllIDs()
	if err != nil {
		global.Logger.Error("[CorpCheck] 查询用户 ID 列表失败", zap.Error(err))
		return
	}

	global.Logger.Info("[CorpCheck] 开始军团准入检查", zap.Int("users", len(ids)))
	for _, uid := range ids {
		if err := rollSvc.CheckCorpAccessAndAdjustRole(ctx, uid); err != nil {
			global.Logger.Warn("[CorpCheck] 检查失败",
				zap.Uint("user_id", uid),
				zap.Error(err))
		}
	}
	global.Logger.Info("[CorpCheck] 军团准入检查完成")
}
