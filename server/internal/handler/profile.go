package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/model"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"

	"github.com/gin-gonic/gin"
)

type ProfileHandler struct{ svc *service.UserService }

func NewProfileHandler() *ProfileHandler { return &ProfileHandler{svc: service.NewUserService()} }

func (h *ProfileHandler) Get(c *gin.Context) {
	profile, err := h.svc.GetProfile(middleware.GetUserID(c))
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, profile)
}

func (h *ProfileHandler) Update(c *gin.Context) {
	var req model.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	profile, err := h.svc.UpdateProfile(middleware.GetUserID(c), req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, profile)
}
