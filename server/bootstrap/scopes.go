package bootstrap

import (
	"amiya-eden/internal/service"
	"amiya-eden/pkg/eve/esi"

	"go.uber.org/zap"

	"amiya-eden/global"
)

// InitScopes 将 ESI 任务模块声明的 scope 注册到 SSO 服务
// 必须在 InitCron 之后调用（确保 ESI 任务的 init() 已执行）
func InitScopes() {
	tasks := esi.AllTasks()

	var count int
	for _, task := range tasks {
		for _, ts := range task.RequiredScopes() {
			service.RegisterScope(task.Name(), ts.Scope, ts.Description, true)
			count++
		}
	}

	global.Logger.Info("ESI scope 注册完成",
		zap.Int("task_count", len(tasks)),
		zap.Int("scope_count", count),
	)

	// ── 舰队模块额外 scope（非 ESI Task 驱动） ──
	service.RegisterScope("fleet", "esi-fleets.read_fleet.v1", "读取舰队信息与成员列表", true)
	service.RegisterScope("fleet", "esi-fleets.write_fleet.v1", "更新舰队 MOTD、邀请成员", true)
}
