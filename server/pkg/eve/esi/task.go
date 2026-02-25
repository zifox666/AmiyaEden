package esi

import (
	"errors"
	"fmt"
	"time"
)

// ─────────────────────────────────────────────
//  错误定义
// ─────────────────────────────────────────────

var (
	// ErrNotModified ESI 返回 304
	ErrNotModified = errors.New("ESI: not modified")
)

// ─────────────────────────────────────────────
//  优先级
// ─────────────────────────────────────────────

// Priority 任务优先级（数字越小优先级越高）
type Priority int

const (
	PriorityCritical Priority = 1  // 关键任务（如 killmail）
	PriorityHigh     Priority = 10 // 高优先级
	PriorityNormal   Priority = 50 // 标准优先级
	PriorityLow      Priority = 90 // 低优先级
)

// ─────────────────────────────────────────────
//  刷新间隔配置
// ─────────────────────────────────────────────

// RefreshInterval 刷新间隔配置
type RefreshInterval struct {
	Active   time.Duration // 活跃角色刷新间隔
	Inactive time.Duration // 不活跃角色刷新间隔
}

// ─────────────────────────────────────────────
//  任务接口
// ─────────────────────────────────────────────

// TaskScope 定义任务所需的 ESI scope
type TaskScope struct {
	Scope       string // ESI scope 字符串
	Description string // scope 描述
}

// RefreshTask ESI 数据刷新任务接口
// 每种数据类型实现此接口；每个任务放在独立的 .go 文件中
type RefreshTask interface {
	// Name 任务唯一标识（如 "character_assets"）
	Name() string

	// Description 任务可读描述
	Description() string

	// Priority 任务优先级
	Priority() Priority

	// Interval 返回刷新间隔配置
	Interval() RefreshInterval

	// RequiredScopes 任务所需的 ESI scope 列表
	RequiredScopes() []TaskScope

	// Execute 执行数据刷新
	// characterID: 角色 ID
	// accessToken: 当前有效的 ESI access_token
	// client: ESI HTTP 客户端
	// 返回 error（nil 表示成功）
	Execute(ctx *TaskContext) error
}

// TaskContext 任务执行上下文
type TaskContext struct {
	CharacterID int64
	AccessToken string
	Client      *Client
	IsActive    bool // 角色是否活跃
}

// ─────────────────────────────────────────────
//  批量任务接口（可选实现）
// ─────────────────────────────────────────────

// BatchTask 批量执行的任务（如 affiliation 可一次查询多个角色）
type BatchTask interface {
	RefreshTask

	// ExecuteBatch 批量执行
	// characterIDs: 需要刷新的角色 ID 列表
	ExecuteBatch(client *Client, characterIDs []int64) error
}

// ─────────────────────────────────────────────
//  任务注册表
// ─────────────────────────────────────────────

var registry = make(map[string]RefreshTask)

// Register 注册一个刷新任务
// 通常在各任务文件的 init() 中调用
func Register(task RefreshTask) {
	name := task.Name()
	if _, exists := registry[name]; exists {
		panic(fmt.Sprintf("ESI refresh task %q already registered", name))
	}
	registry[name] = task
}

// GetTask 根据名称获取任务
func GetTask(name string) (RefreshTask, bool) {
	t, ok := registry[name]
	return t, ok
}

// AllTasks 获取所有已注册的任务（返回副本）
func AllTasks() map[string]RefreshTask {
	result := make(map[string]RefreshTask, len(registry))
	for k, v := range registry {
		result[k] = v
	}
	return result
}

// TaskNames 获取所有已注册的任务名称
func TaskNames() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}

// ─────────────────────────────────────────────
//  任务状态（用于可视化）
// ─────────────────────────────────────────────

// TaskStatus 单个任务在某角色下的状态
type TaskStatus struct {
	TaskName    string     `json:"task_name"`
	Description string     `json:"description"`
	CharacterID int64      `json:"character_id"`
	Priority    Priority   `json:"priority"`
	LastRun     *time.Time `json:"last_run,omitempty"`
	NextRun     *time.Time `json:"next_run,omitempty"`
	Status      string     `json:"status"` // pending | running | success | failed
	Error       string     `json:"error,omitempty"`
}
