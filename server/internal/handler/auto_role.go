package handler

import (
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// AutoRoleHandler ESI 自动权限映射管理
type AutoRoleHandler struct {
	svc *service.AutoRoleService
}

func NewAutoRoleHandler() *AutoRoleHandler {
	return &AutoRoleHandler{svc: service.NewAutoRoleService()}
}

// ─── ESI Role Mapping ───

// ListEsiRoleMappings 获取所有 ESI 角色映射
func (h *AutoRoleHandler) ListEsiRoleMappings(c *gin.Context) {
	mappings, err := h.svc.ListEsiRoleMappings()
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, mappings)
}

// GetAllEsiRoles 获取所有可用的 ESI 军团角色名列表
func (h *AutoRoleHandler) GetAllEsiRoles(c *gin.Context) {
	response.OK(c, h.svc.GetAllEsiRoles())
}

type createEsiRoleMappingRequest struct {
	EsiRole string `json:"esi_role" binding:"required"`
	RoleID  uint   `json:"role_id"  binding:"required"`
}

// CreateEsiRoleMapping 创建 ESI 角色映射
func (h *AutoRoleHandler) CreateEsiRoleMapping(c *gin.Context) {
	var req createEsiRoleMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	mapping, err := h.svc.CreateEsiRoleMapping(req.EsiRole, req.RoleID)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, mapping)
}

// DeleteEsiRoleMapping 删除 ESI 角色映射
func (h *AutoRoleHandler) DeleteEsiRoleMapping(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeParamError, "无效的映射ID")
		return
	}
	if err := h.svc.DeleteEsiRoleMapping(uint(id)); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}

// ─── ESI Title Mapping ───

// ListEsiTitleMappings 获取所有 ESI 头衔映射
func (h *AutoRoleHandler) ListEsiTitleMappings(c *gin.Context) {
	mappings, err := h.svc.ListEsiTitleMappings()
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, mappings)
}

// ListCorpTitles 获取数据库中所有军团头衔（用于前端下拉选择）
func (h *AutoRoleHandler) ListCorpTitles(c *gin.Context) {
	titles, err := h.svc.ListCorpTitles()
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, titles)
}

type createEsiTitleMappingRequest struct {
	CorporationID int64  `json:"corporation_id" binding:"required"`
	TitleID       int    `json:"title_id"       binding:"required"`
	TitleName     string `json:"title_name"`
	RoleID        uint   `json:"role_id"        binding:"required"`
}

// CreateEsiTitleMapping 创建 ESI 头衔映射
func (h *AutoRoleHandler) CreateEsiTitleMapping(c *gin.Context) {
	var req createEsiTitleMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	mapping, err := h.svc.CreateEsiTitleMapping(req.CorporationID, req.TitleID, req.TitleName, req.RoleID)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, mapping)
}

// DeleteEsiTitleMapping 删除 ESI 头衔映射
func (h *AutoRoleHandler) DeleteEsiTitleMapping(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeParamError, "无效的映射ID")
		return
	}
	if err := h.svc.DeleteEsiTitleMapping(uint(id)); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}

// ─── 手动同步 ───

// TriggerSync 手动触发自动权限同步
func (h *AutoRoleHandler) TriggerSync(c *gin.Context) {
	go h.svc.SyncAllUsersAutoRoles(c.Request.Context())
	response.OK(c, "同步任务已触发")
}

// ─── 同步日志 ───

// ListAutoRoleLogs 分页查询自动权限操作日志
func (h *AutoRoleHandler) ListAutoRoleLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("current", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	logs, total, err := h.svc.ListAutoRoleLogs(page, size)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OKWithPage(c, logs, total, page, size)
}
