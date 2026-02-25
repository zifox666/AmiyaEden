package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"

	"github.com/gin-gonic/gin"
)

// MeHandler 当前登录用户信息处理器
type MeHandler struct {
	userSvc *service.UserService
	ssoSvc  *service.EveSSOService
}

func NewMeHandler() *MeHandler {
	return &MeHandler{
		userSvc: service.NewUserService(),
		ssoSvc:  service.NewEveSSOService(),
	}
}

// GetMe 获取当前登录用户信息及绑定角色
//
// GET /api/v1/me（需要 JWT）
func (h *MeHandler) GetMe(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Fail(c, response.CodeUnauthorized, "未登录")
		return
	}

	user, err := h.userSvc.GetUserByID(userID)
	if err != nil {
		response.Fail(c, response.CodeNotFound, "用户不存在")
		return
	}

	characters, err := h.ssoSvc.GetCharactersByUserID(userID)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}

	response.OK(c, gin.H{
		"user":       user,
		"characters": characters,
	})
}
