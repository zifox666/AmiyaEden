package repository

import (
	"amiya-eden/internal/model"
	"errors"
	"fmt"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newTaskTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s_%d?mode=memory&cache=shared", t.Name(), time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.TaskSchedule{}, &model.TaskExecution{}); err != nil {
		t.Fatalf("auto migrate task models: %v", err)
	}

	return db
}

func TestTaskRepository_GetScheduleMissing(t *testing.T) {
	db := newTaskTestDB(t)
	repo := &TaskRepository{db: db}

	schedule, err := repo.GetSchedule("missing-task")
	if err == nil {
		t.Fatal("expected missing schedule to return an error")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected gorm.ErrRecordNotFound, got %v", err)
	}
	if schedule != nil {
		t.Fatalf("expected missing schedule to return nil, got %#v", schedule)
	}
}

func TestTaskRepository_UpsertSchedule(t *testing.T) {
	db := newTaskTestDB(t)
	repo := &TaskRepository{db: db}

	initial := &model.TaskSchedule{
		TaskName:  "task-a",
		CronExpr:  "0 */10 * * * *",
		UpdatedBy: 1,
	}
	if err := repo.UpsertSchedule(initial); err != nil {
		t.Fatalf("insert schedule: %v", err)
	}

	updated := &model.TaskSchedule{
		TaskName:  "task-a",
		CronExpr:  "0 */30 * * * *",
		UpdatedBy: 2,
		UpdatedAt: time.Date(2026, time.April, 10, 12, 30, 0, 0, time.UTC),
	}
	if err := repo.UpsertSchedule(updated); err != nil {
		t.Fatalf("update schedule: %v", err)
	}

	schedule, err := repo.GetSchedule("task-a")
	if err != nil {
		t.Fatalf("get schedule: %v", err)
	}
	if schedule == nil {
		t.Fatal("expected persisted schedule")
	}
	if schedule.CronExpr != updated.CronExpr {
		t.Fatalf("cron expr = %q, want %q", schedule.CronExpr, updated.CronExpr)
	}
	if schedule.UpdatedBy != updated.UpdatedBy {
		t.Fatalf("updated by = %d, want %d", schedule.UpdatedBy, updated.UpdatedBy)
	}

	var count int64
	if err := db.Model(&model.TaskSchedule{}).Where("task_name = ?", "task-a").Count(&count).Error; err != nil {
		t.Fatalf("count schedules: %v", err)
	}
	if count != 1 {
		t.Fatalf("schedule row count = %d, want 1", count)
	}
}

func TestTaskRepository_ListAllSchedules(t *testing.T) {
	db := newTaskTestDB(t)
	repo := &TaskRepository{db: db}

	for _, schedule := range []model.TaskSchedule{
		{TaskName: "task-b", CronExpr: "0 0 * * * *", UpdatedBy: 1},
		{TaskName: "task-a", CronExpr: "0 30 * * * *", UpdatedBy: 2},
	} {
		schedule := schedule
		if err := repo.UpsertSchedule(&schedule); err != nil {
			t.Fatalf("upsert schedule %s: %v", schedule.TaskName, err)
		}
	}

	schedules, err := repo.ListAllSchedules()
	if err != nil {
		t.Fatalf("list schedules: %v", err)
	}
	if len(schedules) != 2 {
		t.Fatalf("schedule count = %d, want 2", len(schedules))
	}
	if schedules[0].TaskName != "task-a" || schedules[1].TaskName != "task-b" {
		t.Fatalf("unexpected schedule order: %#v", schedules)
	}
}

func TestTaskRepository_CreateExecution(t *testing.T) {
	db := newTaskTestDB(t)
	repo := &TaskRepository{db: db}
	startedAt := time.Date(2026, time.April, 10, 13, 0, 0, 0, time.UTC)

	exec := &model.TaskExecution{
		TaskName:  "task-a",
		Trigger:   "manual",
		Status:    "running",
		StartedAt: startedAt,
	}
	if err := repo.CreateExecution(exec); err != nil {
		t.Fatalf("create execution: %v", err)
	}
	if exec.ID == 0 {
		t.Fatal("expected execution ID to be assigned")
	}

	var stored model.TaskExecution
	if err := db.First(&stored, exec.ID).Error; err != nil {
		t.Fatalf("reload execution: %v", err)
	}
	if stored.TaskName != exec.TaskName || stored.Status != exec.Status || !stored.StartedAt.Equal(startedAt) {
		t.Fatalf("stored execution = %#v, want %#v", stored, *exec)
	}
}

func TestTaskRepository_UpdateExecution(t *testing.T) {
	db := newTaskTestDB(t)
	repo := &TaskRepository{db: db}
	startedAt := time.Date(2026, time.April, 10, 13, 15, 0, 0, time.UTC)

	exec := &model.TaskExecution{
		TaskName:  "task-a",
		Trigger:   "cron",
		Status:    "running",
		StartedAt: startedAt,
	}
	if err := repo.CreateExecution(exec); err != nil {
		t.Fatalf("create execution: %v", err)
	}

	finishedAt := startedAt.Add(5 * time.Second)
	durationMs := int64(5000)
	exec.Status = "success"
	exec.FinishedAt = &finishedAt
	exec.DurationMs = &durationMs
	exec.Summary = "completed"
	if err := repo.UpdateExecution(exec); err != nil {
		t.Fatalf("update execution: %v", err)
	}

	var stored model.TaskExecution
	if err := db.First(&stored, exec.ID).Error; err != nil {
		t.Fatalf("reload execution: %v", err)
	}
	if stored.Status != "success" {
		t.Fatalf("status = %q, want success", stored.Status)
	}
	if stored.FinishedAt == nil || !stored.FinishedAt.Equal(finishedAt) {
		t.Fatalf("finished_at = %v, want %v", stored.FinishedAt, finishedAt)
	}
	if stored.DurationMs == nil || *stored.DurationMs != durationMs {
		t.Fatalf("duration_ms = %v, want %d", stored.DurationMs, durationMs)
	}
	if stored.Summary != "completed" {
		t.Fatalf("summary = %q, want completed", stored.Summary)
	}
}

func TestTaskRepository_UpdateExecutionMissingIDDoesNotInsert(t *testing.T) {
	db := newTaskTestDB(t)
	repo := &TaskRepository{db: db}

	exec := &model.TaskExecution{
		TaskName:  "task-a",
		Trigger:   "cron",
		Status:    "success",
		StartedAt: time.Date(2026, time.April, 10, 13, 20, 0, 0, time.UTC),
	}

	if err := repo.UpdateExecution(exec); err == nil {
		t.Fatal("expected update execution without ID to return an error")
	}

	var count int64
	if err := db.Model(&model.TaskExecution{}).Count(&count).Error; err != nil {
		t.Fatalf("count executions: %v", err)
	}
	if count != 0 {
		t.Fatalf("execution row count = %d, want 0", count)
	}
}

func TestTaskRepository_ListExecutions(t *testing.T) {
	db := newTaskTestDB(t)
	repo := &TaskRepository{db: db}
	base := time.Date(2026, time.April, 10, 14, 0, 0, 0, time.UTC)

	fixtures := []model.TaskExecution{
		{TaskName: "task-a", Trigger: "cron", Status: "success", StartedAt: base.Add(1 * time.Minute)},
		{TaskName: "task-a", Trigger: "cron", Status: "failed", StartedAt: base.Add(2 * time.Minute)},
		{TaskName: "task-a", Trigger: "manual", Status: "success", StartedAt: base.Add(3 * time.Minute)},
		{TaskName: "task-b", Trigger: "cron", Status: "success", StartedAt: base.Add(4 * time.Minute)},
	}
	for _, exec := range fixtures {
		exec := exec
		if err := repo.CreateExecution(&exec); err != nil {
			t.Fatalf("create execution for %s: %v", exec.TaskName, err)
		}
	}

	execs, total, err := repo.ListExecutions("task-a", "success", 1, 1)
	if err != nil {
		t.Fatalf("list executions: %v", err)
	}
	if total != 2 {
		t.Fatalf("total = %d, want 2", total)
	}
	if len(execs) != 1 {
		t.Fatalf("page size = %d, want 1", len(execs))
	}
	if execs[0].StartedAt != base.Add(3*time.Minute) {
		t.Fatalf("first page started_at = %v, want %v", execs[0].StartedAt, base.Add(3*time.Minute))
	}

	nextPage, nextTotal, err := repo.ListExecutions("task-a", "success", 2, 1)
	if err != nil {
		t.Fatalf("list executions page 2: %v", err)
	}
	if nextTotal != 2 {
		t.Fatalf("page 2 total = %d, want 2", nextTotal)
	}
	if len(nextPage) != 1 {
		t.Fatalf("page 2 size = %d, want 1", len(nextPage))
	}
	if nextPage[0].StartedAt != base.Add(1*time.Minute) {
		t.Fatalf("second page started_at = %v, want %v", nextPage[0].StartedAt, base.Add(1*time.Minute))
	}
}

func TestTaskRepository_GetLastExecution(t *testing.T) {
	db := newTaskTestDB(t)
	repo := &TaskRepository{db: db}

	last, err := repo.GetLastExecution("task-a")
	if err != nil {
		t.Fatalf("get missing last execution: %v", err)
	}
	if last != nil {
		t.Fatalf("expected no last execution, got %#v", last)
	}

	startedAt := time.Date(2026, time.April, 10, 15, 0, 0, 0, time.UTC)
	first := model.TaskExecution{TaskName: "task-a", Trigger: "cron", Status: "success", StartedAt: startedAt}
	second := model.TaskExecution{TaskName: "task-a", Trigger: "cron", Status: "failed", StartedAt: startedAt.Add(2 * time.Minute)}
	for _, exec := range []model.TaskExecution{first, second} {
		exec := exec
		if err := repo.CreateExecution(&exec); err != nil {
			t.Fatalf("create execution: %v", err)
		}
	}

	last, err = repo.GetLastExecution("task-a")
	if err != nil {
		t.Fatalf("get last execution: %v", err)
	}
	if last == nil {
		t.Fatal("expected last execution")
	}
	if last.Status != "failed" {
		t.Fatalf("last status = %q, want failed", last.Status)
	}
}

func TestTaskRepository_GetLastExecutions(t *testing.T) {
	db := newTaskTestDB(t)
	repo := &TaskRepository{db: db}
	sharedStartedAt := time.Date(2026, time.April, 10, 16, 0, 0, 0, time.UTC)

	fixtures := []model.TaskExecution{
		{TaskName: "task-a", Trigger: "cron", Status: "success", StartedAt: sharedStartedAt.Add(-1 * time.Minute)},
		{TaskName: "task-a", Trigger: "cron", Status: "running", StartedAt: sharedStartedAt},
		{TaskName: "task-a", Trigger: "manual", Status: "failed", StartedAt: sharedStartedAt},
		{TaskName: "task-b", Trigger: "manual", Status: "success", StartedAt: sharedStartedAt.Add(1 * time.Minute)},
		{TaskName: "task-c", Trigger: "cron", Status: "success", StartedAt: sharedStartedAt.Add(2 * time.Minute)},
	}
	for _, exec := range fixtures {
		exec := exec
		if err := repo.CreateExecution(&exec); err != nil {
			t.Fatalf("create execution for %s: %v", exec.TaskName, err)
		}
	}

	lastExecs, err := repo.GetLastExecutions([]string{"task-a", "task-b", "task-missing"})
	if err != nil {
		t.Fatalf("get last executions: %v", err)
	}
	if len(lastExecs) != 2 {
		t.Fatalf("result size = %d, want 2", len(lastExecs))
	}
	if lastExecs["task-a"] == nil {
		t.Fatal("expected task-a result")
	}
	if lastExecs["task-a"].Status != "failed" {
		t.Fatalf("task-a status = %q, want failed", lastExecs["task-a"].Status)
	}
	if lastExecs["task-b"] == nil {
		t.Fatal("expected task-b result")
	}
	if lastExecs["task-b"].Status != "success" {
		t.Fatalf("task-b status = %q, want success", lastExecs["task-b"].Status)
	}
	if _, ok := lastExecs["task-missing"]; ok {
		t.Fatalf("did not expect missing task in result map: %#v", lastExecs["task-missing"])
	}

	empty, err := repo.GetLastExecutions(nil)
	if err != nil {
		t.Fatalf("get last executions for empty input: %v", err)
	}
	if len(empty) != 0 {
		t.Fatalf("empty input result size = %d, want 0", len(empty))
	}
}

func TestTaskRepository_DeleteExecutionsOlderThan(t *testing.T) {
	db := newTaskTestDB(t)
	repo := &TaskRepository{db: db}
	cutoff := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)

	fixtures := []model.TaskExecution{
		{TaskName: "task-old-1", Trigger: "cron", Status: "success", StartedAt: cutoff.Add(-48 * time.Hour)},
		{TaskName: "task-old-2", Trigger: "manual", Status: "failed", StartedAt: cutoff.Add(-time.Hour)},
		{TaskName: "task-new", Trigger: "cron", Status: "success", StartedAt: cutoff.Add(time.Hour)},
	}
	for _, exec := range fixtures {
		exec := exec
		if err := repo.CreateExecution(&exec); err != nil {
			t.Fatalf("create execution for %s: %v", exec.TaskName, err)
		}
	}

	deleted, err := repo.DeleteExecutionsOlderThan(cutoff)
	if err != nil {
		t.Fatalf("DeleteExecutionsOlderThan returned error: %v", err)
	}
	if deleted != 2 {
		t.Fatalf("deleted = %d, want 2", deleted)
	}

	remaining, total, err := repo.ListExecutions("", "", 1, 10)
	if err != nil {
		t.Fatalf("list executions after cleanup: %v", err)
	}
	if total != 1 || len(remaining) != 1 {
		t.Fatalf("remaining executions = %d/%d, want 1/1", len(remaining), total)
	}
	if remaining[0].TaskName != "task-new" {
		t.Fatalf("remaining task = %q, want task-new", remaining[0].TaskName)
	}
}
