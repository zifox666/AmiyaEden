package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RoleHandler struct {
	svc *service.RoleService
}

func NewRoleHandler() *RoleHandler {
	return &RoleHandler{svc: service.NewRoleService()}
}

// ListRoleDefinitions 返回系统角色定义列表（纯内存，不查库）
func (h *RoleHandler) ListRoleDefinitions(c *gin.Context) {
	response.OK(c, h.svc.ListRoleDefinitions())
}

// ─── 用户角色管理 ───

func (h *RoleHandler) GetUserRoles(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id > math.MaxUint32 {
		response.Fail(c, response.CodeParamError, "无效的用户ID")
		return
	}
	roles, err := h.svc.GetUserRoles(uint(id))
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, roles)
}

type setUserRolesRequest struct {
	RoleCodes []string `json:"role_codes"`
}

func (h *RoleHandler) SetUserRoles(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || userID > math.MaxUint32 {
		response.Fail(c, response.CodeParamError, "无效的用户ID")
		return
	}
	var req setUserRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	operatorID := middleware.GetUserID(c)
	operatorRoles := middleware.GetUserRoles(c)
	if err := h.svc.SetUserRoles(c.Request.Context(), operatorID, operatorRoles, uint(userID), req.RoleCodes); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}
