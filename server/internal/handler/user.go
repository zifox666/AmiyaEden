package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/repository"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户 HTTP 处理器
type UserHandler struct {
	svc    *service.UserService
	ssoSvc *service.EveSSOService
}

func NewUserHandler() *UserHandler {
	return &UserHandler{
		svc:    service.NewUserService(),
		ssoSvc: service.NewEveSSOService(),
	}
}

// GetMe 获取当前登录用户信息及绑定角色
func (h *UserHandler) GetMe(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Fail(c, response.CodeUnauthorized, "未登录")
		return
	}
	user, err := h.svc.GetUserByID(userID)
	if err != nil {
		response.Fail(c, response.CodeNotFound, "用户不存在")
		return
	}
	chars, err := h.ssoSvc.GetCharactersByUserID(userID)
	if err != nil {
		chars = nil
	}
	response.OK(c, gin.H{
		"user":       user,
		"characters": chars,
	})
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

// List 获取用户列表（支持 nickname / status / role 筛选）
func (h *UserHandler) List(c *gin.Context) {
	current, _ := strconv.Atoi(c.DefaultQuery("current", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	filter := repository.UserFilter{
		Nickname: c.Query("nickname"),
		Role:     c.Query("role"),
	}
	if statusStr := c.Query("status"); statusStr != "" {
		if s, err := strconv.Atoi(statusStr); err == nil {
			filter.Status = &s
		}
	}

	users, total, err := h.svc.ListUsers(current, size, filter)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, gin.H{
		"records": users,
		"current": current,
		"size":    size,
		"total":   total,
	})
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

// updateRoleRequest 修改角色请求体
type updateRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

// UpdateRole 修改用户角色（需要 Admin 或以上权限）
func (h *UserHandler) UpdateRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeParamError, "无效的用户ID")
		return
	}
	var req updateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	operatorRole := middleware.GetUserRole(c)
	if err := h.svc.UpdateUserRole(operatorRole, uint(id), req.Role); err != nil {
		response.Fail(c, response.CodeForbidden, err.Error())
		return
	}
	response.OK(c, nil)
}
