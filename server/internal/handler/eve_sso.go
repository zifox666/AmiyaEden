package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"
	"fmt"
	"net/url"

	"github.com/gin-gonic/gin"
)

// EveSSOHandler EVE SSO 登录处理器
type EveSSOHandler struct {
	svc *service.EveSSOService
}

func NewEveSSOHandler() *EveSSOHandler {
	return &EveSSOHandler{svc: service.NewEveSSOService()}
}

// Login 发起 EVE SSO 登录，重定向到 EVE 授权页面
//
// GET /api/v1/sso/eve/login?redirect=xxx&scopes=esi-xxx.v1,esi-yyy.v1
func (h *EveSSOHandler) Login(c *gin.Context) {
	redirectURL := c.Query("redirect")
	scopesParam := c.Query("scopes")

	var extraScopes []string
	if scopesParam != "" {
		for _, s := range splitCSV(scopesParam) {
			if s != "" {
				extraScopes = append(extraScopes, s)
			}
		}
	}

	authURL, err := h.svc.GetAuthURL(c.Request.Context(), extraScopes, redirectURL)
	if err != nil {
		response.Fail(c, response.CodeBizError, "生成授权 URL 失败: "+err.Error())
		return
	}

	response.OK(c, gin.H{"url": authURL})
}

// Callback 处理 EVE SSO OAuth 回调
//
// GET /api/v1/sso/eve/callback?code=xxx&state=xxx
func (h *EveSSOHandler) Callback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	errParam := c.Query("error")

	// 尝试从 state 中恢复前端 redirect URL，用于错误时也能跳回前端
	frontendRedirect := h.svc.GetRedirectURLFromState(c.Request.Context(), state)

	// 错误重定向辅助函数：带 error 参数跳回前端 callback 页面
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
		redirectError("EVE SSO 授权被拒绝: " + errDesc)
		return
	}

	clientIP := c.ClientIP()
	result, err := h.svc.HandleCallback(c.Request.Context(), code, state, clientIP)
	if err != nil {
		redirectError("登录处理失败: " + err.Error())
		return
	}

	// 如果有前端重定向地址，则带 token 跳转
	if result.RedirectURL != "" {
		c.Redirect(302, result.RedirectURL+"?token="+result.Token)
		return
	}

	response.OK(c, gin.H{
		"token":     result.Token,
		"user":      result.User,
		"character": result.Character,
	})
}

// GetScopes 获取所有已注册的 ESI Scope 列表
//
// GET /api/v1/sso/eve/scopes
func (h *EveSSOHandler) GetScopes(c *gin.Context) {
	scopes := service.GetRegisteredScopes()
	response.OK(c, scopes)
}

// GetMyCharacters 获取当前用户绑定的所有 EVE 角色
//
// GET /api/v1/sso/eve/characters（需要 JWT）
func (h *EveSSOHandler) GetMyCharacters(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Fail(c, response.CodeUnauthorized, "未登录")
		return
	}

	chars, err := h.svc.GetCharactersByUserID(userID)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, chars)
}

// BindLogin 发起「绑定新角色」的 EVE SSO 授权
// 与 Login 类似，但 state 中记录当前用户 ID，回调时将角色绑到该用户
//
// GET /api/v1/sso/eve/bind?redirect=xxx&scopes=esi-xxx.v1
func (h *EveSSOHandler) BindLogin(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Fail(c, response.CodeUnauthorized, "未登录")
		return
	}

	redirectURL := c.Query("redirect")
	scopesParam := c.Query("scopes")

	var extraScopes []string
	if scopesParam != "" {
		for _, s := range splitCSV(scopesParam) {
			if s != "" {
				extraScopes = append(extraScopes, s)
			}
		}
	}

	authURL, err := h.svc.GetBindAuthURL(c.Request.Context(), userID, extraScopes, redirectURL)
	if err != nil {
		response.Fail(c, response.CodeBizError, "生成授权 URL 失败: "+err.Error())
		return
	}

	response.OK(c, gin.H{"url": authURL})
}

// SetPrimary 设置主角色
//
// PUT /api/v1/sso/eve/primary/:character_id
func (h *EveSSOHandler) SetPrimary(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Fail(c, response.CodeUnauthorized, "未登录")
		return
	}

	cidStr := c.Param("character_id")
	var characterID int64
	if _, err := fmt.Sscanf(cidStr, "%d", &characterID); err != nil || characterID <= 0 {
		response.Fail(c, response.CodeParamError, "无效的角色 ID")
		return
	}

	if err := h.svc.SetPrimaryCharacter(userID, characterID); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}

	response.OK(c, nil)
}

// Unbind 解除绑定角色
//
// DELETE /api/v1/sso/eve/characters/:character_id
func (h *EveSSOHandler) Unbind(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Fail(c, response.CodeUnauthorized, "未登录")
		return
	}

	cidStr := c.Param("character_id")
	var characterID int64
	if _, err := fmt.Sscanf(cidStr, "%d", &characterID); err != nil || characterID <= 0 {
		response.Fail(c, response.CodeParamError, "无效的角色 ID")
		return
	}

	if err := h.svc.UnbindCharacter(userID, characterID); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}

	response.OK(c, nil)
}

// splitCSV 按逗号或空格分割字符串
func splitCSV(s string) []string {
	var result []string
	for _, part := range splitAny(s, ",; ") {
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func splitAny(s string, seps string) []string {
	splitter := func(r rune) bool {
		for _, sep := range seps {
			if r == sep {
				return true
			}
		}
		return false
	}
	result := []string{}
	start := 0
	for i, r := range s {
		if splitter(r) {
			if i > start {
				result = append(result, s[start:i])
			}
			start = i + 1
		}
	}
	if start < len(s) {
		result = append(result, s[start:])
	}
	return result
}
