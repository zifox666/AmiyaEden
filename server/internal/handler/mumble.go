package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"

	"github.com/gin-gonic/gin"
)

type MumbleHandler struct {
	svc *service.MumbleService
}

func NewMumbleHandler() *MumbleHandler {
	return &MumbleHandler{
		svc: service.NewMumbleService(),
	}
}

func (h *MumbleHandler) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	profile, err := h.svc.GetProfile(userID)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, profile)
}

func (h *MumbleHandler) ResetPassword(c *gin.Context) {
	userID := middleware.GetUserID(c)
	account, err := h.svc.ResetPassword(userID)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, account)
}

func (h *MumbleHandler) GetConfig(c *gin.Context) {
	response.OK(c, h.svc.GetConfig())
}

func (h *MumbleHandler) UpdateConfig(c *gin.Context) {
	var req service.UpdateMumbleConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	if err := h.svc.UpdateConfig(req); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *MumbleHandler) ListRoleGroups(c *gin.Context) {
	mappings, err := h.svc.ListRoleGroups()
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, mappings)
}

func (h *MumbleHandler) UpdateRoleGroups(c *gin.Context) {
	var req service.UpdateMumbleRoleGroupsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	if err := h.svc.UpdateRoleGroups(req); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *MumbleHandler) ICEAuthenticate(c *gin.Context) {
	if !h.svc.CheckICEAuthSecret(c.GetHeader("X-Mumble-Auth-Secret")) {
		response.Fail(c, response.CodeForbidden, "权限不足")
		return
	}

	var req service.MumbleICEAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}

	result, err := h.svc.AuthenticateICE(req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}
