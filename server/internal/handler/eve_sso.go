package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"

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

	c.Redirect(302, authURL)
}

// Callback 处理 EVE SSO OAuth 回调
//
// GET /api/v1/sso/eve/callback?code=xxx&state=xxx
func (h *EveSSOHandler) Callback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	errParam := c.Query("error")

	if errParam != "" {
		errDesc := c.DefaultQuery("error_description", errParam)
		response.Fail(c, response.CodeUnauthorized, "EVE SSO 授权被拒绝: "+errDesc)
		return
	}

	clientIP := c.ClientIP()
	result, err := h.svc.HandleCallback(c.Request.Context(), code, state, clientIP)
	if err != nil {
		response.Fail(c, response.CodeBizError, "登录处理失败: "+err.Error())
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
