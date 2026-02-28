package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"

	"github.com/gin-gonic/gin"
)

// NotificationHandler 通知 HTTP 处理器
type NotificationHandler struct {
	svc *service.NotificationService
}

func NewNotificationHandler() *NotificationHandler {
	return &NotificationHandler{svc: service.NewNotificationService()}
}

// ListNotifications POST /notification/list
// 获取当前用户所有角色的通知列表（分页）
func (h *NotificationHandler) ListNotifications(c *gin.Context) {
	var req service.ListNotificationsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}
	userID := middleware.GetUserID(c)
	result, err := h.svc.ListNotifications(userID, &req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

// GetUnreadCount POST /notification/unread-count
// 获取当前用户所有角色的未读通知数量
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID := middleware.GetUserID(c)
	count, err := h.svc.GetUnreadCount(userID)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, gin.H{"unread_count": count})
}

// MarkAsRead POST /notification/read
// 将指定通知标记为已读
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	var req service.MarkAsReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}
	if err := h.svc.MarkAsRead(&req); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}

// MarkAllAsRead POST /notification/read-all
// 将当前用户所有角色的通知标记为已读
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if err := h.svc.MarkAllAsRead(userID); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}
