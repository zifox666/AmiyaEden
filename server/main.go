package main

import (
	"amiya-eden/bootstrap"
	"amiya-eden/global"
	"amiya-eden/jobs"
	"amiya-eden/pkg/jwt"
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func main() {
	// 初始化配置
	bootstrap.InitConfig()

	// 初始化日志
	bootstrap.InitLogger()

	// 初始化 JWT 密钥
	jwt.Init(global.Config.JWT.Secret)

	// 初始化数据库
	bootstrap.InitDB()

	// 初始化 Redis
	bootstrap.InitRedis()

	// 初始化定时任务
	bootstrap.InitCron()

	// 启动时异步检查 SDE 是否有新版本
	go jobs.SdeCheckOnStartup()

	// 将 ESI 任务模块的 scope 注册到 SSO 服务
	bootstrap.InitScopes()

	// 初始化路由
	r := bootstrap.InitRouter()

	// 启动 HTTP 服务
	srv := &http.Server{
		Addr:    ":" + global.Config.Server.Port,
		Handler: r,
	}

	go func() {
		global.Logger.Info("服务启动", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			global.Logger.Fatal("服务启动失败", zap.Error(err))
		}
	}()

	// 优雅关停
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	global.Logger.Info("正在关闭服务...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 停止定时任务
	global.Cron.Stop()

	if err := srv.Shutdown(ctx); err != nil {
		global.Logger.Error("服务关闭异常", zap.Error(err))
	}

	global.Logger.Info("服务已退出")
}
