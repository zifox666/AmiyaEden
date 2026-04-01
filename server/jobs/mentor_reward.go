package jobs

import (
	"amiya-eden/global"
	"amiya-eden/internal/service"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

func registerMentorRewardJob(c *cron.Cron) {
	svc := service.NewMentorRewardService()
	id, err := c.AddFunc("0 0 3 * * *", func() {
		result, err := svc.ProcessRewards(time.Now())
		if err != nil {
			global.Logger.Error("导师奖励处理失败", zap.Error(err))
			return
		}
		global.Logger.Info("导师奖励处理完成",
			zap.Int("processed_relationships", result.ProcessedRelationships),
			zap.Int("rewards_distributed", result.RewardsDistributed),
			zap.Float64("total_coin_awarded", result.TotalCoinAwarded),
			zap.Int("graduated_count", result.GraduatedCount),
		)
	})
	if err != nil {
		global.Logger.Error("注册导师奖励每日任务失败", zap.Error(err))
		return
	}
	global.Logger.Info("注册导师奖励每日任务成功", zap.Int("entry_id", int(id)))
}
