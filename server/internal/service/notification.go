package service

import (
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"errors"
)

// NotificationService 通知业务逻辑层
type NotificationService struct {
	repo     *repository.NotificationRepository
	charRepo *repository.EveCharacterRepository
}

func NewNotificationService() *NotificationService {
	return &NotificationService{
		repo:     repository.NewNotificationRepository(),
		charRepo: repository.NewEveCharacterRepository(),
	}
}

// ListNotificationsRequest 查询通知列表请求
type ListNotificationsRequest struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Type     string `json:"type"`              // 通知类型筛选（可选）
	IsRead   *bool  `json:"is_read,omitempty"` // 已读状态筛选（可选）
}

// NotificationSummary 通知摘要（含未读数）
type NotificationSummary struct {
	List        []model.EveCharacterNotification `json:"list"`
	Total       int64                            `json:"total"`
	Page        int                              `json:"page"`
	PageSize    int                              `json:"page_size"`
	UnreadCount int64                            `json:"unread_count"`
}

// ListNotifications 获取当前用户所有角色的通知列表
func (s *NotificationService) ListNotifications(userID uint, req *ListNotificationsRequest) (*NotificationSummary, error) {
	// 1. 获取用户绑定的所有角色
	chars, err := s.charRepo.ListByUserID(userID)
	if err != nil {
		return nil, errors.New("获取角色列表失败")
	}
	if len(chars) == 0 {
		return &NotificationSummary{
			List:     []model.EveCharacterNotification{},
			Total:    0,
			Page:     req.Page,
			PageSize: req.PageSize,
		}, nil
	}

	charIDs := make([]int64, 0, len(chars))
	for _, c := range chars {
		charIDs = append(charIDs, c.CharacterID)
	}

	// 2. 构建筛选条件
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	filter := repository.NotificationFilter{
		CharacterIDs: charIDs,
		Type:         req.Type,
		IsRead:       req.IsRead,
	}

	// 3. 查询通知列表
	list, total, err := s.repo.ListByCharacterIDs(page, pageSize, filter)
	if err != nil {
		return nil, errors.New("查询通知失败")
	}

	// 4. 查询未读数
	unreadCount, _ := s.repo.CountUnread(charIDs)

	return &NotificationSummary{
		List:        list,
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
		UnreadCount: unreadCount,
	}, nil
}

// GetUnreadCount 获取当前用户所有角色的未读通知数量
func (s *NotificationService) GetUnreadCount(userID uint) (int64, error) {
	chars, err := s.charRepo.ListByUserID(userID)
	if err != nil {
		return 0, errors.New("获取角色列表失败")
	}
	if len(chars) == 0 {
		return 0, nil
	}

	charIDs := make([]int64, 0, len(chars))
	for _, c := range chars {
		charIDs = append(charIDs, c.CharacterID)
	}

	return s.repo.CountUnread(charIDs)
}

// MarkAsReadRequest 标记已读请求
type MarkAsReadRequest struct {
	NotificationIDs []uint `json:"notification_ids" binding:"required"`
}

// MarkAsRead 将指定通知标记为已读
func (s *NotificationService) MarkAsRead(req *MarkAsReadRequest) error {
	if len(req.NotificationIDs) == 0 {
		return errors.New("通知 ID 列表不能为空")
	}
	return s.repo.MarkAsRead(req.NotificationIDs)
}

// MarkAllAsRead 将当前用户所有角色的通知标记为已读
func (s *NotificationService) MarkAllAsRead(userID uint) error {
	chars, err := s.charRepo.ListByUserID(userID)
	if err != nil {
		return errors.New("获取角色列表失败")
	}
	if len(chars) == 0 {
		return nil
	}

	charIDs := make([]int64, 0, len(chars))
	for _, c := range chars {
		charIDs = append(charIDs, c.CharacterID)
	}

	return s.repo.MarkAllAsRead(charIDs)
}
