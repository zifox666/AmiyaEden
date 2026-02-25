package handler

import (
	"amiya-eden/jobs"
	"amiya-eden/pkg/eve/esi"
	"amiya-eden/pkg/response"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// ESIRefreshHandler ESI 数据刷新队列处理器
type ESIRefreshHandler struct{}

func NewESIRefreshHandler() *ESIRefreshHandler {
	return &ESIRefreshHandler{}
}

// TaskInfoItem 任务定义信息（用于前端展示所有可用任务）
type TaskInfoItem struct {
	Name             string   `json:"name"`
	Description      string   `json:"description"`
	Priority         int      `json:"priority"`
	ActiveInterval   string   `json:"active_interval"`
	InactiveInterval string   `json:"inactive_interval"`
	RequiredScopes   []string `json:"required_scopes"`
}

// GetTasks 获取所有已注册的刷新任务定义
//
// GET /api/v1/esi/refresh/tasks
func (h *ESIRefreshHandler) GetTasks(c *gin.Context) {
	allTasks := esi.AllTasks()
	items := make([]TaskInfoItem, 0, len(allTasks))
	for _, t := range allTasks {
		scopes := make([]string, 0)
		for _, s := range t.RequiredScopes() {
			scopes = append(scopes, s.Scope)
		}
		items = append(items, TaskInfoItem{
			Name:             t.Name(),
			Description:      t.Description(),
			Priority:         int(t.Priority()),
			ActiveInterval:   formatDuration(t.Interval().Active),
			InactiveInterval: formatDuration(t.Interval().Inactive),
			RequiredScopes:   scopes,
		})
	}
	// 按优先级排序
	sort.Slice(items, func(i, j int) bool {
		return items[i].Priority < items[j].Priority
	})
	response.OK(c, items)
}

// GetStatuses 获取所有任务的运行时状态（支持分页和筛选）
//
// GET /api/v1/esi/refresh/statuses?current=1&size=20&task_name=xxx&status=xxx
func (h *ESIRefreshHandler) GetStatuses(c *gin.Context) {
	queue := jobs.GetESIQueue()
	if queue == nil {
		response.OK(c, gin.H{
			"records": []interface{}{},
			"current": 1,
			"size":    20,
			"total":   0,
		})
		return
	}

	all := queue.GetAllStatuses()

	// 筛选
	taskNameFilter := c.Query("task_name")
	statusFilter := c.Query("status")

	filtered := make([]*esi.TaskStatus, 0, len(all))
	for _, s := range all {
		if taskNameFilter != "" && s.TaskName != taskNameFilter {
			continue
		}
		if statusFilter != "" && s.Status != statusFilter {
			continue
		}
		filtered = append(filtered, s)
	}

	total := len(filtered)

	// 分页
	current, _ := strconv.Atoi(c.DefaultQuery("current", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if current < 1 {
		current = 1
	}
	if size < 1 {
		size = 20
	}

	start := (current - 1) * size
	if start > total {
		start = total
	}
	end := start + size
	if end > total {
		end = total
	}

	response.OK(c, gin.H{
		"records": filtered[start:end],
		"current": current,
		"size":    size,
		"total":   total,
	})
}

// RunTaskRequest 手动触发单个任务的请求（指定角色）
type RunTaskRequest struct {
	TaskName    string `json:"task_name" binding:"required"`
	CharacterID int64  `json:"character_id" binding:"required"`
}

// RunTask 手动触发指定任务（指定角色）
//
// POST /api/v1/esi/refresh/run
func (h *ESIRefreshHandler) RunTask(c *gin.Context) {
	var req RunTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "参数错误: "+err.Error())
		return
	}

	queue := jobs.GetESIQueue()
	if queue == nil {
		response.Fail(c, response.CodeBizError, "刷新队列未初始化")
		return
	}

	if err := queue.RunTask(req.TaskName, req.CharacterID); err != nil {
		response.Fail(c, response.CodeBizError, "任务触发失败: "+err.Error())
		return
	}

	response.OK(c, gin.H{"message": "任务已触发"})
}

// RunTaskByNameRequest 按任务名称触发所有角色
type RunTaskByNameRequest struct {
	TaskName string `json:"task_name" binding:"required"`
}

// RunTaskByName 手动触发指定任务（所有角色）
//
// POST /api/v1/esi/refresh/run-task
func (h *ESIRefreshHandler) RunTaskByName(c *gin.Context) {
	var req RunTaskByNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "参数错误: "+err.Error())
		return
	}

	queue := jobs.GetESIQueue()
	if queue == nil {
		response.Fail(c, response.CodeBizError, "刷新队列未初始化")
		return
	}

	go func() {
		_ = queue.RunTaskByName(req.TaskName)
	}()

	response.OK(c, gin.H{"message": fmt.Sprintf("任务 %s 已触发（所有角色）", req.TaskName)})
}

// RunAll 手动触发全量刷新
//
// POST /api/v1/esi/refresh/run-all
func (h *ESIRefreshHandler) RunAll(c *gin.Context) {
	queue := jobs.GetESIQueue()
	if queue == nil {
		response.Fail(c, response.CodeBizError, "刷新队列未初始化")
		return
	}

	go queue.Run() // 异步执行，避免超时
	response.OK(c, gin.H{"message": "全量刷新已触发"})
}

// formatDuration 格式化 time.Duration 为可读字符串
func formatDuration(d time.Duration) string {
	if d >= 24*time.Hour {
		days := int(d / (24 * time.Hour))
		if days == 1 {
			return "1 Day"
		}
		return fmt.Sprintf("%d Days", days)
	}
	if d >= time.Hour {
		hours := int(d / time.Hour)
		if hours == 1 {
			return "1 Hour"
		}
		return fmt.Sprintf("%d Hours", hours)
	}
	if d >= time.Minute {
		minutes := int(d / time.Minute)
		if minutes == 1 {
			return "1 Minute"
		}
		return fmt.Sprintf("%d Minutes", minutes)
	}
	return d.String()
}
