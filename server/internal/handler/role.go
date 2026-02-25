package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/model"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RoleHandler struct {
	svc *service.RoleService
}

func NewRoleHandler() *RoleHandler {
	return &RoleHandler{svc: service.NewRoleService()}
}

// ─── 角色 CRUD ───

func (h *RoleHandler) ListRoles(c *gin.Context) {
	current, _ := strconv.Atoi(c.DefaultQuery("current", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	roles, total, err := h.svc.ListRoles(current, size)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, gin.H{"records": roles, "current": current, "size": size, "total": total})
}

func (h *RoleHandler) ListAllRoles(c *gin.Context) {
	roles, err := h.svc.ListAllRoles()
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, roles)
}

func (h *RoleHandler) GetRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeParamError, "无效的角色ID")
		return
	}
	role, err := h.svc.GetRole(uint(id))
	if err != nil {
		response.Fail(c, response.CodeNotFound, "角色不存在")
		return
	}
	response.OK(c, role)
}

type createRoleRequest struct {
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Sort        int    `json:"sort"`
}

func (h *RoleHandler) CreateRole(c *gin.Context) {
	var req createRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	role := &model.Role{Code: req.Code, Name: req.Name, Description: req.Description, Sort: req.Sort}
	if err := h.svc.CreateRole(role); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, role)
}

type updateRoleReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Sort        int    `json:"sort"`
}

func (h *RoleHandler) UpdateRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeParamError, "无效的角色ID")
		return
	}
	var req updateRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	role := &model.Role{Name: req.Name, Description: req.Description, Sort: req.Sort}
	role.ID = uint(id)
	if err := h.svc.UpdateRole(role); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *RoleHandler) DeleteRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeParamError, "无效的角色ID")
		return
	}
	if err := h.svc.DeleteRole(uint(id)); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}

// ─── 角色权限（菜单）管理 ───

func (h *RoleHandler) GetRoleMenus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeParamError, "无效的角色ID")
		return
	}
	menuIDs, err := h.svc.GetRoleMenuIDs(uint(id))
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	if menuIDs == nil {
		menuIDs = []uint{}
	}
	response.OK(c, menuIDs)
}

type setRoleMenusRequest struct {
	MenuIDs []uint `json:"menu_ids"`
}

func (h *RoleHandler) SetRoleMenus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeParamError, "无效的角色ID")
		return
	}
	var req setRoleMenusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	if err := h.svc.SetRoleMenus(c.Request.Context(), uint(id), req.MenuIDs); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}

// ─── 用户角色管理 ───

func (h *RoleHandler) GetUserRoles(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeParamError, "无效的用户ID")
		return
	}
	roles, err := h.svc.GetUserRoles(uint(id))
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	if roles == nil {
		roles = []model.Role{}
	}
	response.OK(c, roles)
}

type setUserRolesRequest struct {
	RoleIDs []uint `json:"role_ids"`
}

func (h *RoleHandler) SetUserRoles(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeParamError, "无效的用户ID")
		return
	}
	var req setUserRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	operatorRoles := middleware.GetUserRoles(c)
	if err := h.svc.SetUserRoles(c.Request.Context(), operatorRoles, uint(userID), req.RoleIDs); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}
