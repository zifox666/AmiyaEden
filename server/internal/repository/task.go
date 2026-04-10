package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var errTaskExecutionIDRequired = errors.New("task execution id is required")

type TaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository() *TaskRepository {
	return &TaskRepository{db: global.DB}
}

func NewTaskRepositoryWithDB(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) dbOrGlobal() *gorm.DB {
	if r != nil && r.db != nil {
		return r.db
	}
	return global.DB
}

func (r *TaskRepository) GetSchedule(taskName string) (*model.TaskSchedule, error) {
	var schedule model.TaskSchedule
	err := r.dbOrGlobal().Where("task_name = ?", taskName).First(&schedule).Error
	if err != nil {
		return nil, err
	}
	return &schedule, nil
}

func (r *TaskRepository) ListAllSchedules() ([]model.TaskSchedule, error) {
	var schedules []model.TaskSchedule
	err := r.dbOrGlobal().Order("task_name ASC").Find(&schedules).Error
	return schedules, err
}

func (r *TaskRepository) UpsertSchedule(schedule *model.TaskSchedule) error {
	return r.dbOrGlobal().Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "task_name"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"cron_expr",
			"updated_by",
			"updated_at",
		}),
	}).Create(schedule).Error
}

func (r *TaskRepository) CreateExecution(execution *model.TaskExecution) error {
	return r.dbOrGlobal().Create(execution).Error
}

func (r *TaskRepository) UpdateExecution(execution *model.TaskExecution) error {
	if execution == nil || execution.ID == 0 {
		return errTaskExecutionIDRequired
	}

	result := r.dbOrGlobal().Model(&model.TaskExecution{}).
		Where("id = ?", execution.ID).
		Updates(map[string]any{
			"status":      execution.Status,
			"finished_at": execution.FinishedAt,
			"duration_ms": execution.DurationMs,
			"error":       execution.Error,
			"summary":     execution.Summary,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *TaskRepository) ListExecutions(taskName, status string, page, pageSize int) ([]model.TaskExecution, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	query := r.dbOrGlobal().Model(&model.TaskExecution{})
	if taskName != "" {
		query = query.Where("task_name = ?", taskName)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var executions []model.TaskExecution
	err := query.
		Order("started_at DESC").
		Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&executions).Error
	if err != nil {
		return nil, 0, err
	}

	return executions, total, nil
}

func (r *TaskRepository) GetLastExecution(taskName string) (*model.TaskExecution, error) {
	var execution model.TaskExecution
	err := r.dbOrGlobal().
		Where("task_name = ?", taskName).
		Order("started_at DESC").
		Order("id DESC").
		First(&execution).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &execution, nil
}

func (r *TaskRepository) GetLastExecutions(taskNames []string) (map[string]*model.TaskExecution, error) {
	if len(taskNames) == 0 {
		return map[string]*model.TaskExecution{}, nil
	}

	var executions []model.TaskExecution
	err := r.dbOrGlobal().
		Table("task_executions AS te").
		Select("te.*").
		Joins(
			`LEFT JOIN task_executions AS newer ON newer.task_name = te.task_name AND (newer.started_at > te.started_at OR (newer.started_at = te.started_at AND newer.id > te.id))`,
		).
		Where("te.task_name IN ?", taskNames).
		Where("newer.id IS NULL").
		Order("te.task_name ASC").
		Find(&executions).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]*model.TaskExecution, len(executions))
	for i := range executions {
		result[executions[i].TaskName] = &executions[i]
	}

	return result, nil
}

func (r *TaskRepository) DeleteExecutionsOlderThan(cutoff time.Time) (int64, error) {
	result := r.dbOrGlobal().Where("started_at < ?", cutoff).Delete(&model.TaskExecution{})
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}
