package esi

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// ─────────────────────────────────────────────
//  Character Notifications 角色通知
//  GET /characters/{character_id}/notifications
//  默认刷新间隔: 1 Day / 不活跃: 7 Days
// ─────────────────────────────────────────────

func init() {
	Register(&NotificationsTask{})
}

// NotificationsTask 角色通知刷新任务
type NotificationsTask struct{}

func (t *NotificationsTask) Name() string        { return "character_notifications" }
func (t *NotificationsTask) Description() string { return "角色通知消息" }
func (t *NotificationsTask) Priority() Priority  { return PriorityNormal }

func (t *NotificationsTask) Interval() RefreshInterval {
	return RefreshInterval{
		Active:   24 * time.Hour,
		Inactive: 7 * 24 * time.Hour,
	}
}

func (t *NotificationsTask) RequiredScopes() []TaskScope {
	return []TaskScope{
		{Scope: "esi-characters.read_notifications.v1", Description: "读取角色通知"},
	}
}

// Notification 通知消息
type Notification struct {
	NotificationID int64     `json:"notification_id"`
	SenderID       int64     `json:"sender_id"`
	SenderType     string    `json:"sender_type"`
	Text           *string   `json:"text,omitempty"`
	Timestamp      time.Time `json:"timestamp"`
	Type           string    `json:"type"`
	IsRead         *bool     `json:"is_read,omitempty"`
}

func (t *NotificationsTask) Execute(ctx *TaskContext) error {
	bgCtx := context.Background()
	path := fmt.Sprintf("/characters/%d/notifications/", ctx.CharacterID)

	var notifications []Notification
	if err := ctx.Client.Get(bgCtx, path, ctx.AccessToken, &notifications); err != nil {
		return fmt.Errorf("fetch notifications: %w", err)
	}

	global.Logger.Debug("[ESI] 角色通知刷新完成",
		zap.Int64("character_id", ctx.CharacterID),
		zap.Int("count", len(notifications)),
	)

	// 入库：使用 upsert 避免重复
	for _, n := range notifications {
		record := model.EveCharacterNotification{
			CharacterID:    ctx.CharacterID,
			NotificationID: n.NotificationID,
			SenderID:       n.SenderID,
			SenderType:     n.SenderType,
			Text:           n.Text,
			Timestamp:      n.Timestamp,
			Type:           n.Type,
			IsRead:         n.IsRead,
		}
		if err := global.DB.Where("character_id = ? AND notification_id = ?", ctx.CharacterID, n.NotificationID).
			Assign(record).FirstOrCreate(&record).Error; err != nil {
			global.Logger.Warn("[ESI] 通知入库失败",
				zap.Int64("notification_id", n.NotificationID),
				zap.Error(err),
			)
		}
	}

	return nil
}
