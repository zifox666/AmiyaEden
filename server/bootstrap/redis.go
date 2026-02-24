package bootstrap

import (
	"amiya-eden/global"
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// InitRedis 初始化 Redis 连接
func InitRedis() {
	cfg := global.Config.Redis

	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		global.Logger.Fatal("Redis 连接失败", zap.Error(err))
	}

	global.Redis = rdb
	global.Logger.Info("Redis 连接成功", zap.String("addr", cfg.Addr), zap.Int("db", cfg.DB))
}
