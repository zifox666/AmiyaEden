package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/repository"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// FleetHandler 舰队 HTTP 处理器
type FleetHandler struct {
	svc *service.FleetService
}

func NewFleetHandler() *FleetHandler {
	return &FleetHandler{
		svc: service.NewFleetService(),
	}
}

// ─────────────────────────────────────────────
//  舰队 CRUD
// ─────────────────────────────────────────────

// CreateFleet 创建舰队
func (h *FleetHandler) CreateFleet(c *gin.Context) {
	var req service.CreateFleetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}
	userID := middleware.GetUserID(c)
	fleet, err := h.svc.CreateFleet(userID, &req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, fleet)
}

// ListFleets 分页查询舰队列表
func (h *FleetHandler) ListFleets(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("current", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	filter := repository.FleetFilter{
		Importance: c.Query("importance"),
	}
	if fcStr := c.Query("fc_user_id"); fcStr != "" {
		if id, err := strconv.ParseUint(fcStr, 10, 64); err == nil {
			fcID := uint(id)
			filter.FCUserID = &fcID
		}
	}

	records, total, err := h.svc.ListFleets(page, size, filter)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, gin.H{
		"records": records,
		"current": page,
		"size":    size,
		"total":   total,
	})
}

// GetFleet 获取舰队详情
func (h *FleetHandler) GetFleet(c *gin.Context) {
	fleetID := c.Param("id")
	if fleetID == "" {
		response.Fail(c, response.CodeParamError, "缺少舰队ID")
		return
	}
	fleet, err := h.svc.GetFleet(fleetID)
	if err != nil {
		response.Fail(c, response.CodeNotFound, "舰队不存在")
		return
	}
	response.OK(c, fleet)
}

// RefreshFleetESI 从 ESI 刷新舰队的 esi_fleet_id
func (h *FleetHandler) RefreshFleetESI(c *gin.Context) {
	fleetID := c.Param("id")
	if fleetID == "" {
		response.Fail(c, response.CodeParamError, "缺少舰队ID")
		return
	}
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	fleet, err := h.svc.RefreshESIFleetID(fleetID, userID, userRole)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, fleet)
}

// UpdateFleet 更新舰队信息
func (h *FleetHandler) UpdateFleet(c *gin.Context) {
	fleetID := c.Param("id")
	if fleetID == "" {
		response.Fail(c, response.CodeParamError, "缺少舰队ID")
		return
	}
	var req service.UpdateFleetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	fleet, err := h.svc.UpdateFleet(fleetID, userID, userRole, &req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, fleet)
}

// DeleteFleet 删除舰队（软删除）
func (h *FleetHandler) DeleteFleet(c *gin.Context) {
	fleetID := c.Param("id")
	if fleetID == "" {
		response.Fail(c, response.CodeParamError, "缺少舰队ID")
		return
	}
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	if err := h.svc.DeleteFleet(fleetID, userID, userRole); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}

// ─────────────────────────────────────────────
//  舰队成员
// ─────────────────────────────────────────────

// GetMembers 获取舰队成员列表
func (h *FleetHandler) GetMembers(c *gin.Context) {
	fleetID := c.Param("id")
	if fleetID == "" {
		response.Fail(c, response.CodeParamError, "缺少舰队ID")
		return
	}
	members, err := h.svc.GetMembers(fleetID)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, members)
}

// SyncESIMembers 从 ESI 拉取当前舰队成员并同步到数据库
func (h *FleetHandler) SyncESIMembers(c *gin.Context) {
	fleetID := c.Param("id")
	if fleetID == "" {
		response.Fail(c, response.CodeParamError, "缺少舰队ID")
		return
	}
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	members, err := h.svc.SyncESIMembers(fleetID, userID, userRole)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, members)
}

// ─────────────────────────────────────────────
//  PAP 发放
// ─────────────────────────────────────────────

// IssuePap 发放 PAP
func (h *FleetHandler) IssuePap(c *gin.Context) {
	fleetID := c.Param("id")
	if fleetID == "" {
		response.Fail(c, response.CodeParamError, "缺少舰队ID")
		return
	}
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	if err := h.svc.IssuePap(fleetID, userID, userRole); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}

// GetPapLogs 获取舰队 PAP 发放记录
func (h *FleetHandler) GetPapLogs(c *gin.Context) {
	fleetID := c.Param("id")
	if fleetID == "" {
		response.Fail(c, response.CodeParamError, "缺少舰队ID")
		return
	}
	logs, err := h.svc.GetPapLogs(fleetID)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, logs)
}

// GetMyPapLogs 获取当前用户的 PAP 记录
func (h *FleetHandler) GetMyPapLogs(c *gin.Context) {
	userID := middleware.GetUserID(c)
	logs, err := h.svc.GetUserPapLogs(userID)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, logs)
}

// ─────────────────────────────────────────────
//  邀请链接
// ─────────────────────────────────────────────

// CreateInvite 创建邀请链接
func (h *FleetHandler) CreateInvite(c *gin.Context) {
	fleetID := c.Param("id")
	if fleetID == "" {
		response.Fail(c, response.CodeParamError, "缺少舰队ID")
		return
	}
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	invite, err := h.svc.CreateInvite(fleetID, userID, userRole)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, invite)
}

// GetInvites 获取舰队邀请链接列表
func (h *FleetHandler) GetInvites(c *gin.Context) {
	fleetID := c.Param("id")
	if fleetID == "" {
		response.Fail(c, response.CodeParamError, "缺少舰队ID")
		return
	}
	invites, err := h.svc.GetInvites(fleetID)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, invites)
}

// DeactivateInvite 禁用邀请链接
func (h *FleetHandler) DeactivateInvite(c *gin.Context) {
	inviteID, err := strconv.ParseUint(c.Param("invite_id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeParamError, "无效的邀请ID")
		return
	}
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	if err := h.svc.DeactivateInvite(uint(inviteID), userID, userRole); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}

// ─────────────────────────────────────────────
//  Join（邀请码加入）
// ─────────────────────────────────────────────

type joinFleetRequest struct {
	Code        string `json:"code" binding:"required"`
	CharacterID int64  `json:"character_id" binding:"required"`
}

// JoinFleet 通过邀请码加入舰队
func (h *FleetHandler) JoinFleet(c *gin.Context) {
	var req joinFleetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.svc.JoinFleet(req.Code, userID, req.CharacterID); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}

// ─────────────────────────────────────────────
//  ESI 角色舰队信息
// ─────────────────────────────────────────────

// GetCharacterFleetInfo 获取角色当前所在的 ESI 舰队
func (h *FleetHandler) GetCharacterFleetInfo(c *gin.Context) {
	charID, err := strconv.ParseInt(c.Param("character_id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeParamError, "无效的角色ID")
		return
	}
	userID := middleware.GetUserID(c)
	info, err := h.svc.GetCharacterFleetInfo(userID, charID)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, info)
}
