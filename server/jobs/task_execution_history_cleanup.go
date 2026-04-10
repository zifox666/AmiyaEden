package jobs

import (
	"amiya-eden/global"
	"amiya-eden/internal/repository"
	"amiya-eden/internal/taskregistry"
	"context"
	"time"

	"go.uber.org/zap"
)

func registerTaskExecutionHistoryCleanupTask(reg *taskregistry.Registry) {
	repo := repository.NewTaskRepository()

	reg.Register(taskregistry.TaskDefinition{
		Name:        "task_execution_history_cleanup",
		Description: "Remove task execution history older than three months",
		Category:    taskregistry.TaskCategorySystem,
		Type:        taskregistry.TaskTypeRecurring,
		DefaultCron: "0 0 4 1 * *",
		RunFunc: func(ctx context.Context) error {
			cutoff := time.Now().UTC().AddDate(0, -3, 0)
			deleted, err := repo.DeleteExecutionsOlderThan(cutoff)
			if err != nil {
				global.Logger.Error("任务执行历史清理失败", zap.Error(err), zap.Time("cutoff", cutoff))
				return err
			}

			global.Logger.Info(
				"任务执行历史清理完成",
				zap.Int64("deleted_rows", deleted),
				zap.Time("cutoff", cutoff),
			)
			return nil
		},
	})

	global.Logger.Info("注册任务执行历史清理任务成功", zap.String("task_name", "task_execution_history_cleanup"))
}
