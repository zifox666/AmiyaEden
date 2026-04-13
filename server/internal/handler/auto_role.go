package handler

import (
	"amiya-eden/internal/model"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"
	"context"
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

// ─── SeAT Role Mapping ───

// ListSeatRoleMappings 获取所有 SeAT 分组映射
func (h *AutoRoleHandler) ListSeatRoleMappings(c *gin.Context) {
	mappings, err := h.svc.ListSeatRoleMappings()
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, mappings)
}

// GetAllSeatRoles 获取所有 SeAT 分组名列表（供前端选择）
func (h *AutoRoleHandler) GetAllSeatRoles(c *gin.Context) {
	roles, err := h.svc.GetAllSeatRoles()
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, roles)
}

type createSeatRoleMappingRequest struct {
	SeatRole string `json:"seat_role" binding:"required"`
	RoleID   uint   `json:"role_id"   binding:"required"`
}

// CreateSeatRoleMapping 创建 SeAT 分组映射
func (h *AutoRoleHandler) CreateSeatRoleMapping(c *gin.Context) {
	var req createSeatRoleMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	mapping, err := h.svc.CreateSeatRoleMapping(req.SeatRole, req.RoleID)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, mapping)
}

// DeleteSeatRoleMapping 删除 SeAT 分组映射
func (h *AutoRoleHandler) DeleteSeatRoleMapping(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeParamError, "无效的映射ID")
		return
	}
	if err := h.svc.DeleteSeatRoleMapping(uint(id)); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}

// ─── 手动同步 ───

// TriggerSync 手动触发自动权限同步
func (h *AutoRoleHandler) TriggerSync(c *gin.Context) {
	go func(ctx context.Context) {
		h.svc.SyncAllUsersBasicAccess(ctx)
		h.svc.SyncAllUsersAutoRoles(ctx)
	}(c.Request.Context())
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

// ─── 准入名单管理 ───

// ListAllowedEntities 获取指定名单类型的所有实体
// GET /auto-role/allow-list/:type
func (h *AutoRoleHandler) ListAllowedEntities(c *gin.Context) {
	listType := c.Param("type")
	entities, err := h.svc.ListAllowedEntities(listType)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, entities)
}

type addAllowedEntityRequest struct {
	EntityID   int64  `json:"entity_id"   binding:"required"`
	EntityType string `json:"entity_type" binding:"required"`
	EntityName string `json:"entity_name" binding:"required"`
}

// AddAllowedEntity 添加实体到名单
// POST /auto-role/allow-list/:type
func (h *AutoRoleHandler) AddAllowedEntity(c *gin.Context) {
	listType := c.Param("type")
	var req addAllowedEntityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	e := &model.AllowedEntity{
		ListType:   listType,
		EntityID:   req.EntityID,
		EntityType: req.EntityType,
		EntityName: req.EntityName,
	}
	if err := h.svc.AddAllowedEntity(e); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, e)
}

// RemoveAllowedEntity 从名单中删除实体
// DELETE /auto-role/allow-list/:type/:id
func (h *AutoRoleHandler) RemoveAllowedEntity(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeParamError, "无效的ID")
		return
	}
	if err := h.svc.RemoveAllowedEntity(uint(id)); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}

// ─── EVE 实体搜索 ───

// SearchEveEntities 通过 zkillboard 模糊搜索 EVE 联盟/军团
// GET /auto-role/eve-search?q=...
func (h *AutoRoleHandler) SearchEveEntities(c *gin.Context) {
	q := c.Query("q")
	results, err := h.svc.SearchEveEntities(q)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, results)
}
