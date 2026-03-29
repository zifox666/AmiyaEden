package jobs

import (
	"amiya-eden/global"
	"amiya-eden/internal/service"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// registerAlliancePAPJob 注册联盟 PAP 定时任务:
//   - 每小时整点刷新当月数据
//   - 每月第一天 01:00 补拉上月数据并归档
func registerAlliancePAPJob(c *cron.Cron) {
	svc := service.NewAlliancePAPService()

	// ── 每小时整点刷新当月 ──
	hourlyID, err := c.AddFunc("0 0 * * * *", func() {
		now := time.Now()
		global.Logger.Info("开始联盟 PAP 小时刷新", zap.Int("year", now.Year()), zap.Int("month", int(now.Month())))
		svc.FetchAllUsers(now.Year(), int(now.Month()))
	})
	if err != nil {
		global.Logger.Error("注册联盟 PAP 小时任务失败", zap.Error(err))
		return
	}
	global.Logger.Info("注册联盟 PAP 小时任务成功", zap.Int("entry_id", int(hourlyID)))

	// ── 每月第一天 01:00 归档上月并拉取最终数据 ──
	monthlyID, err := c.AddFunc("0 0 1 1 * *", func() {
		now := time.Now()
		// 上月
		lastMonth := now.AddDate(0, -1, 0)
		year := lastMonth.Year()
		month := int(lastMonth.Month())

		global.Logger.Info("开始联盟 PAP 月度归档", zap.Int("year", year), zap.Int("month", month))

		// 先拉取最终数据
		svc.FetchAllUsers(year, month)

		// 统一调用 SettleMonth 完成归档 + 兑换 + 差额补偿
		// operatorID=0 表示系统自动操作，corporationIDs 为空即处理所有军团
		result, err := svc.SettleMonth(year, month, true, 0, nil)
		if err != nil {
			global.Logger.Error("联盟 PAP 月度结算失败", zap.Int("year", year), zap.Int("month", month), zap.Error(err))
		} else {
			global.Logger.Info("联盟 PAP 月度结算完成",
				zap.Int("year", year),
				zap.Int("month", month),
				zap.Int("settled_users", result.TotalUsers),
				zap.Int("skipped_users", result.SkippedUsers),
				zap.Float64("total_wallet", result.TotalWallet),
				zap.Int("adjusted_users", result.AdjustedUsers),
				zap.Float64("total_adjusted", result.TotalAdjusted),
			)
		}
	})
	if err != nil {
		global.Logger.Error("注册联盟 PAP 月度归档任务失败", zap.Error(err))
		return
	}
	global.Logger.Info("注册联盟 PAP 月度归档任务成功", zap.Int("entry_id", int(monthlyID)))
}
