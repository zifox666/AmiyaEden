package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/repository"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SrpHandler 补损 HTTP 处理器
type SrpHandler struct {
	svc *service.SrpService
}

func NewSrpHandler() *SrpHandler {
	return &SrpHandler{svc: service.NewSrpService()}
}

// ─────────────────────────────────────────────
//  舰船价格表
// ─────────────────────────────────────────────

// ListShipPrices GET /srp/prices
func (h *SrpHandler) ListShipPrices(c *gin.Context) {
	keyword := c.Query("keyword")
	list, err := h.svc.ListShipPrices(keyword)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, list)
}

// UpsertShipPrice POST /srp/prices
func (h *SrpHandler) UpsertShipPrice(c *gin.Context) {
	var req service.UpsertShipPriceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}
	userID := middleware.GetUserID(c)
	p, err := h.svc.UpsertShipPrice(userID, &req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, p)
}

// DeleteShipPrice DELETE /srp/prices/:id
func (h *SrpHandler) DeleteShipPrice(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		response.Fail(c, response.CodeParamError, "无效的 ID")
		return
	}
	if err := h.svc.DeleteShipPrice(uint(id)); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}

// ─────────────────────────────────────────────
//  申请（用户端）
// ─────────────────────────────────────────────

// SubmitApplication POST /srp/applications
func (h *SrpHandler) SubmitApplication(c *gin.Context) {
	var req service.SubmitApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}
	userID := middleware.GetUserID(c)
	app, err := h.svc.SubmitApplication(userID, &req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, app)
}

// ListMyApplications GET /srp/applications/my
func (h *SrpHandler) ListMyApplications(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("current", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	userID := middleware.GetUserID(c)

	list, total, err := h.svc.ListMyApplications(userID, page, size)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, gin.H{
		"records": list,
		"current": page,
		"size":    size,
		"total":   total,
	})
}

// GetMyKillmails GET /srp/my-killmails?character_id=xxx
func (h *SrpHandler) GetMyKillmails(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var characterID int64
	if cidStr := c.Query("character_id"); cidStr != "" {
		if cid, err := strconv.ParseInt(cidStr, 10, 64); err == nil {
			characterID = cid
		}
	}
	kms, err := h.svc.GetMyKillmails(userID, characterID)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, kms)
}

// GetFleetKillmails GET /srp/fleet-killmails?fleet_id=xxx
func (h *SrpHandler) GetFleetKillmails(c *gin.Context) {
	fleetID := c.Query("fleet_id")
	if fleetID == "" {
		response.Fail(c, response.CodeParamError, "缺少 fleet_id 参数")
		return
	}
	userID := middleware.GetUserID(c)
	kms, err := h.svc.GetFleetKillmails(userID, fleetID)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, kms)
}

// ─────────────────────────────────────────────
//  申请管理（管理端：srp/fc/admin）
// ─────────────────────────────────────────────

// ListApplications GET /srp/manage/applications
func (h *SrpHandler) ListApplications(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("current", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	filter := repository.SrpApplicationFilter{
		ReviewStatus: c.Query("review_status"),
		PayoutStatus: c.Query("payout_status"),
	}
	if fleetID := c.Query("fleet_id"); fleetID != "" {
		filter.FleetID = &fleetID
	}
	if charIDStr := c.Query("character_id"); charIDStr != "" {
		if cid, err := strconv.ParseInt(charIDStr, 10, 64); err == nil {
			filter.CharacterID = &cid
		}
	}

	list, total, err := h.svc.ListApplications(page, size, filter)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, gin.H{
		"records": list,
		"current": page,
		"size":    size,
		"total":   total,
	})
}

// GetApplication GET /srp/manage/applications/:id
func (h *SrpHandler) GetApplication(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		response.Fail(c, response.CodeParamError, "无效的 ID")
		return
	}
	app, err := h.svc.GetApplication(uint(id))
	if err != nil {
		response.Fail(c, response.CodeNotFound, "申请不存在")
		return
	}
	response.OK(c, app)
}

// ReviewApplication PATCH /srp/manage/applications/:id/review
func (h *SrpHandler) ReviewApplication(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		response.Fail(c, response.CodeParamError, "无效的 ID")
		return
	}
	var req service.ReviewApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}
	reviewerID := middleware.GetUserID(c)
	app, err := h.svc.ReviewApplication(reviewerID, uint(id), &req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, app)
}

// Payout PATCH /srp/manage/applications/:id/payout
func (h *SrpHandler) Payout(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		response.Fail(c, response.CodeParamError, "无效的 ID")
		return
	}
	var req service.SrpPayoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}
	payerID := middleware.GetUserID(c)
	app, err := h.svc.Payout(payerID, uint(id), &req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, app)
}
