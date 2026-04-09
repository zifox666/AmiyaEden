package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SrpHandler 补损 HTTP 处理器
type SrpHandler struct {
	svc           *service.SrpService
	sysConfigRepo *repository.SysConfigRepository
}

func NewSrpHandler() *SrpHandler {
	return &SrpHandler{
		svc:           service.NewSrpService(),
		sysConfigRepo: repository.NewSysConfigRepository(),
	}
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
	id := requireUintID(c, "id", " ID")
	if id == 0 {
		return
	}
	if err := h.svc.DeleteShipPrice(id); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}

// ─────────────────────────────────────────────
//  SRP 配置（仅 admin）
// ─────────────────────────────────────────────

// SrpConfigResponse SRP 配置响应
type SrpConfigResponse struct {
	AmountLimit float64 `json:"amount_limit"` // SRP 职权单笔上限（ISK）
}

// UpdateSrpConfigRequest 更新 SRP 配置请求
type UpdateSrpConfigRequest struct {
	AmountLimit *float64 `json:"amount_limit" binding:"required,min=0"`
}

// GetSrpConfig GET /srp/config
func (h *SrpHandler) GetSrpConfig(c *gin.Context) {
	amountLimit := h.sysConfigRepo.GetFloat(
		model.SysConfigSRPAmountLimit,
		model.SysConfigDefaultSRPAmountLimit,
	)
	response.OK(c, SrpConfigResponse{AmountLimit: amountLimit})
}

// UpdateSrpConfig PUT /srp/config
func (h *SrpHandler) UpdateSrpConfig(c *gin.Context) {
	var req UpdateSrpConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}
	if err := h.sysConfigRepo.Set(
		model.SysConfigSRPAmountLimit,
		fmt.Sprintf("%g", *req.AmountLimit),
		"SRP 职权单笔审批/发放上限（ISK）",
	); err != nil {
		response.Fail(c, response.CodeBizError, "更新 SRP 配置失败")
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
	page, pageSize, err := parsePaginationQuery(c, 20, 100)
	if err != nil {
		response.Fail(c, response.CodeParamError, err.Error())
		return
	}
	userID := middleware.GetUserID(c)

	list, total, err := h.svc.ListMyApplications(userID, page, pageSize)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OKWithPage(c, list, total, page, pageSize)
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
	fleetID := c.Param("fleet_id")
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
	page, pageSize, err := parseLedgerPaginationQuery(c, 20)
	if err != nil {
		response.Fail(c, response.CodeParamError, err.Error())
		return
	}

	filter := repository.SrpApplicationFilter{
		Tab:          repository.SrpTabType(c.Query("tab")),
		ReviewStatus: c.Query("review_status"),
		PayoutStatus: c.Query("payout_status"),
		Keyword:      c.Query("keyword"),
	}
	if fleetID := c.Query("fleet_id"); fleetID != "" {
		filter.FleetID = &fleetID
	}
	if charIDStr := c.Query("character_id"); charIDStr != "" {
		if cid, err := strconv.ParseInt(charIDStr, 10, 64); err == nil {
			filter.CharacterID = &cid
		}
	}

	list, total, err := h.svc.ListApplications(page, pageSize, filter)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OKWithPage(c, list, total, page, pageSize)
}

// ListBatchPayoutSummary GET /srp/applications/batch-payout-summary
func (h *SrpHandler) ListBatchPayoutSummary(c *gin.Context) {
	list, err := h.svc.ListBatchPayoutSummary()
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, list)
}

// RunFleetAutoApproval PUT /srp/applications/auto-approve
func (h *SrpHandler) RunFleetAutoApproval(c *gin.Context) {
	var req service.RunFleetAutoApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}
	reviewerID := middleware.GetUserID(c)
	callerRoles := middleware.GetUserRoles(c)
	result, err := h.svc.RunFleetAutoApproval(reviewerID, callerRoles, req.FleetID)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

// GetApplication GET /srp/manage/applications/:id
func (h *SrpHandler) GetApplication(c *gin.Context) {
	id := requireUintID(c, "id", " ID")
	if id == 0 {
		return
	}
	app, err := h.svc.GetApplication(id)
	if err != nil {
		response.Fail(c, response.CodeNotFound, "申请不存在")
		return
	}
	response.OK(c, app)
}

// ReviewApplication PATCH /srp/manage/applications/:id/review
func (h *SrpHandler) ReviewApplication(c *gin.Context) {
	id := requireUintID(c, "id", " ID")
	if id == 0 {
		return
	}
	var req service.ReviewApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}
	reviewerID := middleware.GetUserID(c)
	callerRoles := middleware.GetUserRoles(c)
	app, err := h.svc.ReviewApplication(reviewerID, callerRoles, id, &req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, app)
}

// Payout PATCH /srp/manage/applications/:id/payout
func (h *SrpHandler) Payout(c *gin.Context) {
	id := requireUintID(c, "id", " ID")
	if id == 0 {
		return
	}
	var req service.SrpPayoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}
	payerID := middleware.GetUserID(c)
	callerRoles := middleware.GetUserRoles(c)
	app, mailSummary, err := h.svc.Payout(payerID, callerRoles, id, &req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, service.SrpPayoutActionResult{
		SrpApplication:     *app,
		MailAttemptSummary: mailSummary,
	})
}

// BatchPayoutByUser PUT /srp/applications/users/:user_id/payout
func (h *SrpHandler) BatchPayoutByUser(c *gin.Context) {
	userID := requireUintID(c, "user_id", "用户 ID")
	if userID == 0 {
		return
	}
	payerID := middleware.GetUserID(c)
	callerRoles := middleware.GetUserRoles(c)
	result, mailSummary, err := h.svc.BatchPayoutByUser(payerID, callerRoles, userID)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, service.SrpBatchPayoutActionResult{
		SrpBatchPayoutSummaryResponse: *result,
		MailAttemptSummary:            mailSummary,
	})
}

// BatchPayoutAsFuxiCoin PUT /srp/applications/fuxi-payout
func (h *SrpHandler) BatchPayoutAsFuxiCoin(c *gin.Context) {
	payerID := middleware.GetUserID(c)
	callerRoles := middleware.GetUserRoles(c)
	result, mailSummary, err := h.svc.BatchPayoutAsFuxiCoin(payerID, callerRoles)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, service.SrpBatchFuxiPayoutActionResult{
		SrpBatchFuxiPayoutSummary: *result,
		MailAttemptSummary:        mailSummary,
	})
}

// OpenInfoWindow POST /srp/open-info-window
// 通过 ESI 在客户端打开人物信息窗口
func (h *SrpHandler) OpenInfoWindow(c *gin.Context) {
	var req service.OpenInfoWindowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.svc.OpenInfoWindow(userID, &req); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}

// GetKillmailDetail POST /srp/killmails/detail
func (h *SrpHandler) GetKillmailDetail(c *gin.Context) {
	var req service.KillmailDetailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}
	detail, err := h.svc.GetKillmailDetail(&req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, detail)
}
