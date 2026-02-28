package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
)

// NotificationRepository 角色通知数据访问层
type NotificationRepository struct{}

func NewNotificationRepository() *NotificationRepository {
	return &NotificationRepository{}
}

// NotificationFilter 通知查询筛选条件
type NotificationFilter struct {
	CharacterIDs []int64 // 角色 ID 列表
	Type         string  // 通知类型（可选）
	IsRead       *bool   // 已读状态（可选）
}

// ListByCharacterIDs 查询多个角色的通知（分页，按时间倒序）
func (r *NotificationRepository) ListByCharacterIDs(page, pageSize int, filter NotificationFilter) ([]model.EveCharacterNotification, int64, error) {
	var list []model.EveCharacterNotification
	var total int64

	db := global.DB.Model(&model.EveCharacterNotification{})

	if len(filter.CharacterIDs) > 0 {
		db = db.Where("character_id IN ?", filter.CharacterIDs)
	}
	if filter.Type != "" {
		db = db.Where("type = ?", filter.Type)
	}
	if filter.IsRead != nil {
		db = db.Where("is_read = ?", *filter.IsRead)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := db.Order("timestamp DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// CountUnread 统计未读通知数量
func (r *NotificationRepository) CountUnread(characterIDs []int64) (int64, error) {
	var count int64
	isRead := false
	db := global.DB.Model(&model.EveCharacterNotification{}).Where("is_read = ? OR is_read IS NULL", isRead)
	if len(characterIDs) > 0 {
		db = db.Where("character_id IN ?", characterIDs)
	}
	err := db.Count(&count).Error
	return count, err
}

// MarkAsRead 将指定通知标记为已读
func (r *NotificationRepository) MarkAsRead(notificationIDs []uint) error {
	isRead := true
	return global.DB.Model(&model.EveCharacterNotification{}).
		Where("id IN ?", notificationIDs).
		Update("is_read", isRead).Error
}

// MarkAllAsRead 将指定角色的所有通知标记为已读
func (r *NotificationRepository) MarkAllAsRead(characterIDs []int64) error {
	isRead := true
	return global.DB.Model(&model.EveCharacterNotification{}).
		Where("character_id IN ?", characterIDs).
		Update("is_read", isRead).Error
}
