package handler

import (
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户 HTTP 处理器
type UserHandler struct {
	svc *service.UserService
}

func NewUserHandler() *UserHandler {
	return &UserHandler{svc: service.NewUserService()}
}

// Get 获取用户详情
func (h *UserHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeParamError, "无效的用户ID")
		return
	}
	user, err := h.svc.GetUserByID(uint(id))
	if err != nil {
		response.Fail(c, response.CodeNotFound, "用户不存在")
		return
	}
	response.OK(c, user)
}

// List 获取用户列表
func (h *UserHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	users, total, err := h.svc.ListUsers(page, pageSize)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OKWithPage(c, users, total, page, pageSize)
}

// Delete 删除用户
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeParamError, "无效的用户ID")
		return
	}
	if err := h.svc.DeleteUser(uint(id)); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}
