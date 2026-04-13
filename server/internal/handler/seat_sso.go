package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"
	"net/url"

	"github.com/gin-gonic/gin"
)

// SeatSSOHandler SeAT SSO 登录处理器
type SeatSSOHandler struct {
	svc *service.SeatSSOService
}

func NewSeatSSOHandler() *SeatSSOHandler {
	return &SeatSSOHandler{svc: service.NewSeatSSOService()}
}

// Enabled 检查 SeAT 登录是否已启用
//
// GET /api/v1/sso/seat/enabled
func (h *SeatSSOHandler) Enabled(c *gin.Context) {
	response.OK(c, gin.H{"enabled": h.svc.IsSeatEnabled()})
}

// Login 发起 SeAT SSO 登录
//
// GET /api/v1/sso/seat/login?redirect=xxx
func (h *SeatSSOHandler) Login(c *gin.Context) {
	redirectURL := c.Query("redirect")

	authURL, err := h.svc.GetSeatAuthURL(c.Request.Context(), redirectURL)
	if err != nil {
		response.Fail(c, response.CodeBizError, "生成 SeAT 授权 URL 失败: "+err.Error())
		return
	}

	response.OK(c, gin.H{"url": authURL})
}

// Callback 处理 SeAT OAuth 回调
//
// GET /api/v1/sso/seat/callback?code=xxx&state=xxx
func (h *SeatSSOHandler) Callback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	errParam := c.Query("error")

	frontendRedirect := h.svc.GetRedirectURLFromState(c.Request.Context(), state)

	redirectError := func(errMsg string) {
		if frontendRedirect != "" {
			target := frontendRedirect + "?error=" + url.QueryEscape(errMsg)
			c.Redirect(302, target)
			return
		}
		response.Fail(c, response.CodeBizError, errMsg)
	}

	if errParam != "" {
		errDesc := c.DefaultQuery("error_description", errParam)
		redirectError("SeAT 授权被拒绝: " + errDesc)
		return
	}

	clientIP := c.ClientIP()
	result, err := h.svc.HandleSeatCallback(c.Request.Context(), code, state, clientIP)
	if err != nil {
		redirectError("SeAT 登录处理失败: " + err.Error())
		return
	}

	if result.RedirectURL != "" {
		c.Redirect(302, result.RedirectURL+"?token="+result.Token+"&provider=seat")
		return
	}

	response.OK(c, gin.H{
		"token": result.Token,
		"user":  result.User,
	})
}

// Bind 发起 SeAT 账号绑定（已登录用户）
//
// GET /api/v1/sso/seat/bind?redirect=xxx
func (h *SeatSSOHandler) Bind(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Fail(c, response.CodeUnauthorized, "未登录")
		return
	}

	redirectURL := c.Query("redirect")

	authURL, err := h.svc.GetSeatBindURL(c.Request.Context(), userID, redirectURL)
	if err != nil {
		response.Fail(c, response.CodeBizError, "生成 SeAT 绑定 URL 失败: "+err.Error())
		return
	}

	response.OK(c, gin.H{"url": authURL})
}

// GetSeatBinding 获取当前用户的 SeAT 绑定信息
//
// GET /api/v1/sso/seat/binding
func (h *SeatSSOHandler) GetSeatBinding(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Fail(c, response.CodeUnauthorized, "未登录")
		return
	}

	su, err := h.svc.GetSeatUserByUserID(userID)
	if err != nil {
		response.OK(c, gin.H{"bound": false})
		return
	}

	response.OK(c, gin.H{
		"bound":         true,
		"seat_user_id":  su.SeatUserID,
		"seat_username": su.SeatUsername,
		"main_char_id":  su.MainCharID,
		"groups":        su.Groups,
	})
}

// Unbind 解除 SeAT 绑定
//
// DELETE /api/v1/sso/seat/binding
func (h *SeatSSOHandler) Unbind(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Fail(c, response.CodeUnauthorized, "未登录")
		return
	}

	if err := h.svc.UnbindSeat(userID); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}

	response.OK(c, nil)
}

// GetSeatConfig 获取 SeAT 配置（管理员）
//
// GET /api/v1/system/seat-config
func (h *SeatSSOHandler) GetSeatConfig(c *gin.Context) {
	response.OK(c, h.svc.GetSeatAdminConfig())
}

// UpdateSeatConfig 更新 SeAT 配置（管理员）
//
// PUT /api/v1/system/seat-config
func (h *SeatSSOHandler) UpdateSeatConfig(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "参数错误")
		return
	}

	if err := h.svc.UpdateSeatConfig(req); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}

	response.OK(c, nil)
}
