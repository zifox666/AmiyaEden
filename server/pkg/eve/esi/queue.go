package esi

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"amiya-eden/internal/service"
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ─────────────────────────────────────────────
//  队列引擎
// ─────────────────────────────────────────────

// Queue ESI 数据刷新队列
type Queue struct {
	client   *Client
	ssoSvc   *service.EveSSOService
	charRepo *repository.EveCharacterRepository

	mu       sync.RWMutex
	statuses map[string]*TaskStatus // key: "taskName:characterID"

	// 并发控制：同一时间最多执行的任务数
	concurrency int
}

// NewQueue 创建刷新队列
func NewQueue() *Queue {
	return &Queue{
		client:      NewClient(),
		ssoSvc:      service.NewEveSSOService(),
		charRepo:    repository.NewEveCharacterRepository(),
		statuses:    make(map[string]*TaskStatus),
		concurrency: 5, // 默认 5 并发
	}
}

// SetConcurrency 设置最大并发数
func (q *Queue) SetConcurrency(n int) {
	if n < 1 {
		n = 1
	}
	q.concurrency = n
}

// ─────────────────────────────────────────────
//  调度入口
// ─────────────────────────────────────────────

// Run 执行一次完整的刷新调度
// 由 cron 定时触发
func (q *Queue) Run() {
	ctx := context.Background()
	global.Logger.Info("[ESI Queue] 开始刷新调度")

	// 1. 获取所有有 refresh_token 的角色
	characters, err := q.charRepo.ListAllWithToken()
	if err != nil {
		global.Logger.Error("[ESI Queue] 获取角色列表失败", zap.Error(err))
		return
	}

	if len(characters) == 0 {
		global.Logger.Info("[ESI Queue] 没有需要刷新的角色")
		return
	}

	// 2. 检测角色活跃度
	activityMap := q.checkActivity(ctx, characters)

	// 3. 获取所有任务并按优先级排序
	allTasks := AllTasks()
	sortedTasks := sortTasksByPriority(allTasks)

	// 4. 构建待执行任务列表
	type pendingJob struct {
		task      RefreshTask
		character model.EveCharacter
		isActive  bool
	}
	var jobs []pendingJob

	for _, task := range sortedTasks {
		for i := range characters {
			char := characters[i]
			isActive := activityMap[char.CharacterID]

			// 检查角色是否有该任务所需的 scope
			if !q.hasRequiredScopes(char, task) {
				continue
			}

			// 检查是否需要刷新（基于上次执行时间和刷新间隔）
			if !q.needsRefresh(task, char.CharacterID, isActive) {
				continue
			}

			jobs = append(jobs, pendingJob{
				task:      task,
				character: char,
				isActive:  isActive,
			})
		}
	}

	if len(jobs) == 0 {
		global.Logger.Info("[ESI Queue] 没有需要执行的任务")
		return
	}

	global.Logger.Info("[ESI Queue] 开始执行刷新任务",
		zap.Int("total_jobs", len(jobs)),
		zap.Int("characters", len(characters)),
	)

	// 5. 使用信号量控制并发执行
	sem := make(chan struct{}, q.concurrency)
	var wg sync.WaitGroup

	for _, job := range jobs {
		wg.Add(1)
		sem <- struct{}{} // 占位

		go func(j pendingJob) {
			defer wg.Done()
			defer func() { <-sem }() // 释放

			q.executeTask(ctx, j.task, j.character, j.isActive)
		}(job)
	}

	wg.Wait()
	global.Logger.Info("[ESI Queue] 刷新调度完成")
}

// RunTask 手动执行某个指定任务（管理页面触发）
func (q *Queue) RunTask(taskName string, characterID int64) error {
	task, ok := GetTask(taskName)
	if !ok {
		return fmt.Errorf("task %q not found", taskName)
	}

	char, err := q.charRepo.GetByCharacterID(characterID)
	if err != nil {
		return fmt.Errorf("character %d not found: %w", characterID, err)
	}

	ctx := context.Background()
	isActive := q.checkSingleActivity(ctx, *char)

	q.executeTask(ctx, task, *char, isActive)
	return nil
}

// RunTaskByName 对所有拥有所需 scope 的角色执行指定任务
func (q *Queue) RunTaskByName(taskName string) error {
	task, ok := GetTask(taskName)
	if !ok {
		return fmt.Errorf("task %q not found", taskName)
	}

	characters, err := q.charRepo.ListAllWithToken()
	if err != nil {
		return fmt.Errorf("list characters: %w", err)
	}

	ctx := context.Background()
	activityMap := q.checkActivity(ctx, characters)

	sem := make(chan struct{}, q.concurrency)
	var wg sync.WaitGroup

	for i := range characters {
		char := characters[i]
		if !q.hasRequiredScopes(char, task) {
			continue
		}
		isActive := activityMap[char.CharacterID]

		wg.Add(1)
		sem <- struct{}{}
		go func(ch model.EveCharacter, active bool) {
			defer wg.Done()
			defer func() { <-sem }()
			q.executeTask(ctx, task, ch, active)
		}(char, isActive)
	}

	wg.Wait()
	return nil
}

// ─────────────────────────────────────────────
//  内部方法
// ─────────────────────────────────────────────

// executeTask 执行单个任务
func (q *Queue) executeTask(ctx context.Context, task RefreshTask, char model.EveCharacter, isActive bool) {
	statusKey := fmt.Sprintf("%s:%d", task.Name(), char.CharacterID)

	// 更新状态为 running
	q.setStatus(statusKey, &TaskStatus{
		TaskName:    task.Name(),
		Description: task.Description(),
		CharacterID: char.CharacterID,
		Priority:    task.Priority(),
		Status:      "running",
	})

	// 获取有效 Token
	accessToken, err := q.ssoSvc.GetValidToken(ctx, char.CharacterID)
	if err != nil {
		global.Logger.Error("[ESI Queue] 获取 Token 失败",
			zap.String("task", task.Name()),
			zap.Int64("character_id", char.CharacterID),
			zap.Error(err),
		)
		q.setStatus(statusKey, &TaskStatus{
			TaskName:    task.Name(),
			Description: task.Description(),
			CharacterID: char.CharacterID,
			Priority:    task.Priority(),
			Status:      "failed",
			Error:       err.Error(),
		})
		return
	}

	// 执行任务
	taskCtx := &TaskContext{
		CharacterID: char.CharacterID,
		AccessToken: accessToken,
		Client:      q.client,
		IsActive:    isActive,
	}

	if err := task.Execute(taskCtx); err != nil {
		global.Logger.Error("[ESI Queue] 任务执行失败",
			zap.String("task", task.Name()),
			zap.Int64("character_id", char.CharacterID),
			zap.Error(err),
		)
		q.setStatus(statusKey, &TaskStatus{
			TaskName:    task.Name(),
			Description: task.Description(),
			CharacterID: char.CharacterID,
			Priority:    task.Priority(),
			Status:      "failed",
			Error:       err.Error(),
		})
		return
	}

	// 成功：记录上次执行时间
	now := time.Now()
	interval := task.Interval()
	nextDur := interval.Active
	if !isActive {
		nextDur = interval.Inactive
	}
	nextRun := now.Add(nextDur)

	q.setStatus(statusKey, &TaskStatus{
		TaskName:    task.Name(),
		Description: task.Description(),
		CharacterID: char.CharacterID,
		Priority:    task.Priority(),
		LastRun:     &now,
		NextRun:     &nextRun,
		Status:      "success",
	})

	// 将上次执行时间持久化到 Redis
	q.setLastRun(task.Name(), char.CharacterID, now)

	global.Logger.Debug("[ESI Queue] 任务执行成功",
		zap.String("task", task.Name()),
		zap.Int64("character_id", char.CharacterID),
	)
}

// needsRefresh 判断任务是否需要刷新
func (q *Queue) needsRefresh(task RefreshTask, characterID int64, isActive bool) bool {
	lastRun, err := q.getLastRun(task.Name(), characterID)
	if err != nil {
		return true // 没有记录则需要刷新
	}

	interval := task.Interval()
	dur := interval.Active
	if !isActive {
		dur = interval.Inactive
	}

	return time.Since(lastRun) >= dur
}

// hasRequiredScopes 检查角色是否拥有任务所需的 scope
func (q *Queue) hasRequiredScopes(char model.EveCharacter, task RefreshTask) bool {
	charScopes := strings.Fields(char.Scopes)
	scopeSet := make(map[string]struct{}, len(charScopes))
	for _, s := range charScopes {
		scopeSet[s] = struct{}{}
	}

	for _, required := range task.RequiredScopes() {
		if _, ok := scopeSet[required.Scope]; !ok {
			return false
		}
	}
	return true
}

// ─────────────────────────────────────────────
//  Redis 状态存储
// ─────────────────────────────────────────────

const (
	lastRunKeyPrefix = "esi:refresh:lastrun:" // esi:refresh:lastrun:{task}:{characterID}
)

func (q *Queue) setLastRun(taskName string, characterID int64, t time.Time) {
	key := fmt.Sprintf("%s%s:%d", lastRunKeyPrefix, taskName, characterID)
	global.Redis.Set(context.Background(), key, t.Unix(), 0)
}

func (q *Queue) getLastRun(taskName string, characterID int64) (time.Time, error) {
	key := fmt.Sprintf("%s%s:%d", lastRunKeyPrefix, taskName, characterID)
	val, err := global.Redis.Get(context.Background(), key).Int64()
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(val, 0), nil
}

// ─────────────────────────────────────────────
//  状态管理（可视化用）
// ─────────────────────────────────────────────

func (q *Queue) setStatus(key string, status *TaskStatus) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.statuses[key] = status
}

// GetAllStatuses 获取所有任务状态（用于 API 展示）
func (q *Queue) GetAllStatuses() []*TaskStatus {
	q.mu.RLock()
	defer q.mu.RUnlock()

	result := make([]*TaskStatus, 0, len(q.statuses))
	for _, s := range q.statuses {
		result = append(result, s)
	}

	// 按优先级排序
	sort.Slice(result, func(i, j int) bool {
		if result[i].Priority != result[j].Priority {
			return result[i].Priority < result[j].Priority
		}
		return result[i].TaskName < result[j].TaskName
	})
	return result
}

// GetTaskStatuses 获取指定任务的所有角色状态
func (q *Queue) GetTaskStatuses(taskName string) []*TaskStatus {
	q.mu.RLock()
	defer q.mu.RUnlock()

	var result []*TaskStatus
	prefix := taskName + ":"
	for key, s := range q.statuses {
		if strings.HasPrefix(key, prefix) {
			result = append(result, s)
		}
	}
	return result
}

// ─────────────────────────────────────────────
//  辅助
// ─────────────────────────────────────────────

// sortTasksByPriority 按优先级排序任务
func sortTasksByPriority(tasks map[string]RefreshTask) []RefreshTask {
	sorted := make([]RefreshTask, 0, len(tasks))
	for _, t := range tasks {
		sorted = append(sorted, t)
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Priority() < sorted[j].Priority()
	})
	return sorted
}
