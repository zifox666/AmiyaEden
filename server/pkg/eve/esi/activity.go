package esi

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

const (
	// InactiveDays 超过此天数未登录视为不活跃
	InactiveDays = 7

	// activityCachePrefix Redis 缓存活跃状态
	activityCachePrefix = "esi:activity:"
	// activityCacheTTL 活跃状态缓存时间
	activityCacheTTL = 1 * time.Hour
)

// OnlineStatus ESI 在线状态响应
type OnlineStatus struct {
	LastLogin  *time.Time `json:"last_login"`
	LastLogout *time.Time `json:"last_logout"`
	Logins     int        `json:"logins"`
	Online     bool       `json:"online"`
}

// checkActivity 批量检测角色活跃度
// 返回 map[characterID]isActive
func (q *Queue) checkActivity(ctx context.Context, characters []model.EveCharacter) map[int64]bool {
	result := make(map[int64]bool, len(characters))

	for _, char := range characters {
		result[char.CharacterID] = q.checkSingleActivity(ctx, char)
	}

	return result
}

// checkSingleActivity 检测单个角色活跃度
func (q *Queue) checkSingleActivity(ctx context.Context, char model.EveCharacter) bool {
	// 先查 Redis 缓存
	cacheKey := fmt.Sprintf("%s%d", activityCachePrefix, char.CharacterID)
	val, err := global.Redis.Get(ctx, cacheKey).Result()
	if err == nil {
		return val == "1"
	}

	// 查 ESI
	accessToken, err := q.ssoSvc.GetValidToken(ctx, char.CharacterID)
	if err != nil {
		global.Logger.Warn("[ESI Activity] 获取 Token 失败，默认活跃",
			zap.Int64("character_id", char.CharacterID),
			zap.Error(err),
		)
		return true // 获取失败默认视为活跃
	}

	path := fmt.Sprintf("/characters/%d/online/", char.CharacterID)
	var status OnlineStatus
	if err := q.client.Get(ctx, path, accessToken, &status); err != nil {
		global.Logger.Warn("[ESI Activity] 查询在线状态失败，默认活跃",
			zap.Int64("character_id", char.CharacterID),
			zap.Error(err),
		)
		return true
	}

	isActive := true
	if status.LastLogin != nil {
		isActive = time.Since(*status.LastLogin) < time.Duration(InactiveDays)*24*time.Hour
	}

	// 缓存结果
	activeVal := "0"
	if isActive {
		activeVal = "1"
	}
	global.Redis.Set(ctx, cacheKey, activeVal, activityCacheTTL)

	return isActive
}
