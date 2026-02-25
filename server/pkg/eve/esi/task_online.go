package esi

import (
	"amiya-eden/global"
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// ─────────────────────────────────────────────
//  Character Online 在线状态（用于活跃度检测）
//  GET /characters/{character_id}/online
//  常驻 scope，用于判断角色是否活跃
//  默认刷新间隔: 30 Minutes / 不活跃: 2 Hours
// ─────────────────────────────────────────────

func init() {
	Register(&OnlineTask{})
}

// OnlineTask 在线状态刷新任务
type OnlineTask struct{}

func (t *OnlineTask) Name() string        { return "character_online" }
func (t *OnlineTask) Description() string { return "角色在线状态（活跃度检测）" }
func (t *OnlineTask) Priority() Priority  { return PriorityHigh }

func (t *OnlineTask) Interval() RefreshInterval {
	return RefreshInterval{
		Active:   30 * time.Minute,
		Inactive: 2 * time.Hour,
	}
}

func (t *OnlineTask) RequiredScopes() []TaskScope {
	return []TaskScope{
		{Scope: "esi-location.read_online.v1", Description: "读取角色在线状态"},
	}
}

func (t *OnlineTask) Execute(ctx *TaskContext) error {
	bgCtx := context.Background()
	path := fmt.Sprintf("/characters/%d/online/", ctx.CharacterID)

	var status OnlineStatus
	if err := ctx.Client.Get(bgCtx, path, ctx.AccessToken, &status); err != nil {
		return fmt.Errorf("fetch online status: %w", err)
	}

	// 更新 Redis 活跃状态缓存
	isActive := true
	if status.LastLogin != nil {
		isActive = time.Since(*status.LastLogin) < time.Duration(InactiveDays)*24*time.Hour
	}

	cacheKey := fmt.Sprintf("%s%d", activityCachePrefix, ctx.CharacterID)
	activeVal := "0"
	if isActive {
		activeVal = "1"
	}
	global.Redis.Set(bgCtx, cacheKey, activeVal, activityCacheTTL)

	global.Logger.Debug("[ESI] 在线状态刷新完成",
		zap.Int64("character_id", ctx.CharacterID),
		zap.Bool("online", status.Online),
		zap.Bool("active", isActive),
	)

	return nil
}
