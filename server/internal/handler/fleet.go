package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/repository"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"
	"math"
	"strconv"
	"strings"

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
	page, pageSize, err := parsePaginationQuery(c, 20, 100)
	if err != nil {
		response.Fail(c, response.CodeParamError, err.Error())
		return
	}

	filter := repository.FleetFilter{
		Importance: c.Query("importance"),
	}
	if fcStr := c.Query("fc_user_id"); fcStr != "" {
		if id, err := strconv.ParseUint(fcStr, 10, 64); err == nil && id <= math.MaxUint32 {
			fcID := uint(id)
			filter.FCUserID = &fcID
		}
	}

	records, total, err := h.svc.ListFleets(page, pageSize, filter)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OKWithPage(c, records, total, page, pageSize)
}

// GetMyFleets 获取当前用户参与过的舰队列表
func (h *FleetHandler) GetMyFleets(c *gin.Context) {
	userID := middleware.GetUserID(c)
	fleets, err := h.svc.GetMyFleets(userID)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, fleets)
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
	userRoles := middleware.GetUserRoles(c)
	fleet, err := h.svc.RefreshESIFleetID(fleetID, userID, userRoles)
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
	userRoles := middleware.GetUserRoles(c)
	fleet, err := h.svc.UpdateFleet(fleetID, userID, userRoles, &req)
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
	userRoles := middleware.GetUserRoles(c)
	if err := h.svc.DeleteFleet(fleetID, userID, userRoles); err != nil {
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

type manualAddFleetMembersRequest struct {
	CharacterNames []string `json:"character_names" binding:"required"`
}

// GetMembersWithPap 分页查询舰队成员（含 PAP 信息）
func (h *FleetHandler) GetMembersWithPap(c *gin.Context) {
	fleetID := c.Param("id")
	if fleetID == "" {
		response.Fail(c, response.CodeParamError, "缺少舰队ID")
		return
	}
	page, pageSize, err := parsePaginationQuery(c, 260, 260)
	if err != nil {
		response.Fail(c, response.CodeParamError, err.Error())
		return
	}

	list, total, err := h.svc.ListMembersWithPap(fleetID, page, pageSize)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OKWithPage(c, list, total, page, pageSize)
}

// ManualAddMembers 手动添加舰队成员
func (h *FleetHandler) ManualAddMembers(c *gin.Context) {
	fleetID := c.Param("id")
	if fleetID == "" {
		response.Fail(c, response.CodeParamError, "缺少舰队ID")
		return
	}

	var req manualAddFleetMembersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}

	userID := middleware.GetUserID(c)
	userRoles := middleware.GetUserRoles(c)
	result, err := h.svc.ManualAddMembers(fleetID, userID, userRoles, &service.ManualAddFleetMembersRequest{
		CharacterNames: req.CharacterNames,
	})
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

// SyncESIMembers 从 ESI 拉取当前舰队成员并同步到数据库
func (h *FleetHandler) SyncESIMembers(c *gin.Context) {
	fleetID := c.Param("id")
	if fleetID == "" {
		response.Fail(c, response.CodeParamError, "缺少舰队ID")
		return
	}
	userID := middleware.GetUserID(c)
	userRoles := middleware.GetUserRoles(c)
	members, err := h.svc.SyncESIMembers(fleetID, userID, userRoles)
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
	userRoles := middleware.GetUserRoles(c)
	if err := h.svc.IssuePap(fleetID, userID, userRoles); err != nil {
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

// GetCorporationPapSummary 获取军团 PAP 汇总
func (h *FleetHandler) GetCorporationPapSummary(c *gin.Context) {
	page, pageSize, err := parseLedgerPaginationQuery(c, 200)
	if err != nil {
		response.Fail(c, response.CodeParamError, err.Error())
		return
	}
	year, _ := strconv.Atoi(c.Query("year"))
	period := c.DefaultQuery("period", service.CorporationPapPeriodLastMonth)
	corpTickerParam := c.Query("corp_tickers")

	var corpTickers []string
	if corpTickerParam != "" {
		corpTickers = strings.Split(corpTickerParam, ",")
	}

	result, err := h.svc.GetCorporationPapSummary(page, pageSize, period, year, corpTickers)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}

	response.OK(c, result)
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
	userRoles := middleware.GetUserRoles(c)
	invite, err := h.svc.CreateInvite(fleetID, userID, userRoles)
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
	if err != nil || inviteID > math.MaxUint32 {
		response.Fail(c, response.CodeParamError, "无效的邀请ID")
		return
	}
	userID := middleware.GetUserID(c)
	userRoles := middleware.GetUserRoles(c)
	if err := h.svc.DeactivateInvite(uint(inviteID), userID, userRoles); err != nil {
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
//  ESI 人物舰队信息
// ─────────────────────────────────────────────

// GetCharacterFleetInfo 获取人物当前所在的 ESI 舰队
func (h *FleetHandler) GetCharacterFleetInfo(c *gin.Context) {
	charID, err := strconv.ParseInt(c.Param("character_id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeParamError, "无效的人物 ID")
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

// ─────────────────────────────────────────────
//  Webhook Ping
// ─────────────────────────────────────────────

// PingFleet 手动触发舰队 Webhook Ping
func (h *FleetHandler) PingFleet(c *gin.Context) {
	fleetID := c.Param("id")
	if fleetID == "" {
		response.Fail(c, response.CodeParamError, "缺少舰队ID")
		return
	}
	userID := middleware.GetUserID(c)
	userRoles := middleware.GetUserRoles(c)
	if err := h.svc.PingFleet(fleetID, userID, userRoles); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}
