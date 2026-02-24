package bootstrap

import (
	"amiya-eden/global"
	"amiya-eden/jobs"

	"github.com/robfig/cron/v3"
)

// InitCron 初始化并启动定时任务调度器
func InitCron() {
	c := cron.New(
		cron.WithSeconds(), // 支持秒级精度
		cron.WithChain(cron.Recover(cron.DefaultLogger)), // panic 恢复
		cron.WithLogger(newCronLogger()),
	)

	// 注册所有定时任务
	jobs.RegisterAll(c)

	c.Start()
	global.Cron = c
	global.Logger.Info("定时任务调度器已启动")
}

// cronLogger 适配 zap 到 cron.Logger 接口
type cronLogger struct{}

func newCronLogger() cron.Logger {
	return &cronLogger{}
}

func (l *cronLogger) Info(msg string, keysAndValues ...interface{}) {
	global.Logger.Sugar().Infow("[Cron] "+msg, keysAndValues...)
}

func (l *cronLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	global.Logger.Sugar().Errorw("[Cron] "+msg, append([]interface{}{"error", err}, keysAndValues...)...)
}
