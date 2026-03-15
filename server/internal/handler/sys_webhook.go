package handler

import (
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"

	"github.com/gin-gonic/gin"
)

// WebhookHandler Webhook 配置 HTTP 处理器
type WebhookHandler struct {
	svc *service.WebhookService
}

func NewWebhookHandler() *WebhookHandler {
	return &WebhookHandler{svc: service.NewWebhookService()}
}

// GetConfig GET /system/webhook/config
func (h *WebhookHandler) GetConfig(c *gin.Context) {
	cfg, err := h.svc.GetConfig()
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, cfg)
}

// SetConfig PUT /system/webhook/config
func (h *WebhookHandler) SetConfig(c *gin.Context) {
	var req service.WebhookConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}
	if err := h.svc.SetConfig(&req); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}

// TestWebhook POST /system/webhook/test
func (h *WebhookHandler) TestWebhook(c *gin.Context) {
	var req struct {
		URL          string `json:"url" binding:"required"`
		Type         string `json:"type"`
		Content      string `json:"content"`
		OBTargetType string `json:"ob_target_type"`
		OBTargetID   int64  `json:"ob_target_id"`
		OBToken      string `json:"ob_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}
	cfg := &service.WebhookConfig{
		URL:          req.URL,
		Type:         req.Type,
		OBTargetType: req.OBTargetType,
		OBTargetID:   req.OBTargetID,
		OBToken:      req.OBToken,
	}
	if err := h.svc.SendTest(cfg, req.Content); err != nil {
		response.Fail(c, response.CodeBizError, "发送失败: "+err.Error())
		return
	}
	response.OK(c, nil)
}
