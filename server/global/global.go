package global

import (
	"amiya-eden/config"

	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	// Config 全局配置
	Config *config.Config

	// Logger 全局日志
	Logger *zap.Logger

	// DB 全局数据库连接
	DB *gorm.DB

	// Redis 全局 Redis 客户端
	Redis *redis.Client

	// Cron 全局定时任务调度器
	Cron *cron.Cron
)
