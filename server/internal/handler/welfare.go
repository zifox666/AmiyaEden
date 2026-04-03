package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	evidenceMaxBytes = 2048 << 10 // 2MB
)

var evidenceAllowedMIME = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

// UploadEvidence POST /welfare/upload-evidence
// 接收图片文件，验证大小和类型，返回 base64 data URL（不写入文件系统）
func (h *WelfareHandler) UploadEvidence(c *gin.Context) {
	uploadImageAsDataURL(c, evidenceMaxBytes, evidenceAllowedMIME)
}

// WelfareHandler 福利 HTTP 处理器
type WelfareHandler struct {
	svc *service.WelfareService
}

func NewWelfareHandler() *WelfareHandler {
	return &WelfareHandler{svc: service.NewWelfareService()}
}

// ─────────────────────────────────────────────
//  管理员端（全部 POST）
// ─────────────────────────────────────────────

// adminWelfareCreateRequest 创建福利请求
type adminWelfareCreateRequest struct {
	Name             string `json:"name" binding:"required"`
	Description      string `json:"description"`
	DistMode         string `json:"dist_mode" binding:"required,oneof=per_user per_character"`
	PayByFuxiCoin    *int   `json:"pay_by_fuxi_coin"`
	RequireSkillPlan bool   `json:"require_skill_plan"`
	SkillPlanIDs     []uint `json:"skill_plan_ids"`
	MaxCharAgeMonths *int   `json:"max_char_age_months"`
	MinimumPap       *int   `json:"minimum_pap"`
	RequireEvidence  bool   `json:"require_evidence"`
	ExampleEvidence  string `json:"example_evidence"`
	Status           int8   `json:"status"`
	SortOrder        int    `json:"sort_order"`
}

// AdminCreateWelfare POST /system/welfare/add
func (h *WelfareHandler) AdminCreateWelfare(c *gin.Context) {
	var req adminWelfareCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}

	w := &model.Welfare{
		Name:             req.Name,
		Description:      req.Description,
		DistMode:         req.DistMode,
		PayByFuxiCoin:    req.PayByFuxiCoin,
		RequireSkillPlan: req.RequireSkillPlan,
		SkillPlanIDs:     req.SkillPlanIDs,
		MaxCharAgeMonths: req.MaxCharAgeMonths,
		MinimumPap:       req.MinimumPap,
		RequireEvidence:  req.RequireEvidence,
		ExampleEvidence:  req.ExampleEvidence,
		Status:           req.Status,
		SortOrder:        req.SortOrder,
		CreatedBy:        middleware.GetUserID(c),
	}

	if err := h.svc.AdminCreateWelfare(w); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, w)
}

// adminWelfareUpdateRequest 更新福利请求
type adminWelfareUpdateRequest struct {
	ID uint `json:"id" binding:"required"`
	service.AdminUpdateWelfareRequest
}

// AdminUpdateWelfare POST /system/welfare/edit
func (h *WelfareHandler) AdminUpdateWelfare(c *gin.Context) {
	var req adminWelfareUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}

	w, err := h.svc.AdminUpdateWelfare(req.ID, &req.AdminUpdateWelfareRequest)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, w)
}

// adminWelfareDeleteRequest 删除福利请求
type adminWelfareDeleteRequest struct {
	ID uint `json:"id" binding:"required"`
}

// AdminDeleteWelfare POST /system/welfare/delete
func (h *WelfareHandler) AdminDeleteWelfare(c *gin.Context) {
	var req adminWelfareDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}

	if err := h.svc.AdminDeleteWelfare(req.ID); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}

// adminWelfareListRequest 福利列表请求
type adminWelfareListRequest struct {
	Current int    `json:"current"`
	Size    int    `json:"size"`
	Status  *int8  `json:"status"`
	Name    string `json:"name"`
}

// AdminListWelfares POST /system/welfare/list
func (h *WelfareHandler) AdminListWelfares(c *gin.Context) {
	var req adminWelfareListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Current = 1
		req.Size = 20
	}
	req.Current, req.Size = normalizePagination(req.Current, req.Size, 20, 100)

	filter := repository.WelfareFilter{
		Status: req.Status,
		Name:   req.Name,
	}

	list, total, err := h.svc.AdminListWelfares(req.Current, req.Size, filter)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OKWithPage(c, list, total, req.Current, req.Size)
}

// adminWelfareReorderRequest 福利排序请求
type adminWelfareReorderRequest struct {
	IDs []uint `json:"ids" binding:"required"`
}

// AdminReorderWelfares POST /system/welfare/reorder
func (h *WelfareHandler) AdminReorderWelfares(c *gin.Context) {
	var req adminWelfareReorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}
	if err := h.svc.AdminReorderWelfares(req.IDs); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}

// ─────────────────────────────────────────────
//  管理端 - 导入历史记录
// ─────────────────────────────────────────────

// adminImportRecordsRequest 导入历史记录请求
type adminImportRecordsRequest struct {
	WelfareID uint   `json:"welfare_id" binding:"required"`
	CSV       string `json:"csv" binding:"required"`
}

// AdminImportRecords POST /system/welfare/import
func (h *WelfareHandler) AdminImportRecords(c *gin.Context) {
	var req adminImportRecordsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}

	count, err := h.svc.ImportWelfareRecords(&service.ImportWelfareRecordsRequest{
		WelfareID: req.WelfareID,
		CSV:       req.CSV,
	})
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, gin.H{"count": count})
}

// ─────────────────────────────────────────────
//  管理端 - 福利审批
// ─────────────────────────────────────────────

// adminApplicationListRequest 福利申请列表请求
type adminApplicationListRequest struct {
	Current int    `json:"current"`
	Size    int    `json:"size"`
	Status  string `json:"status"`
	Keyword string `json:"keyword"`
}

// AdminListApplications POST /system/welfare/applications
func (h *WelfareHandler) AdminListApplications(c *gin.Context) {
	var req adminApplicationListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Current = 1
		req.Size = 50
	}
	req.Current, req.Size = normalizeLedgerPagination(req.Current, req.Size)

	var filter repository.WelfareApplicationFilter
	if strings.Contains(req.Status, ",") {
		filter.StatusIn = strings.Split(req.Status, ",")
	} else {
		filter.Status = req.Status
	}
	filter.Keyword = req.Keyword

	list, total, err := h.svc.AdminListApplications(req.Current, req.Size, filter)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OKWithPage(c, list, total, req.Current, req.Size)
}

// adminDeleteApplicationRequest 删除申请记录请求
type adminDeleteApplicationRequest struct {
	ID uint `json:"id" binding:"required"`
}

// AdminDeleteApplication POST /system/welfare/applications/delete
func (h *WelfareHandler) AdminDeleteApplication(c *gin.Context) {
	var req adminDeleteApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}
	if err := h.svc.AdminDeleteApplication(req.ID); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}

// adminReviewApplicationRequest 审批请求
type adminReviewApplicationRequest struct {
	ID     uint   `json:"id" binding:"required"`
	Action string `json:"action" binding:"required,oneof=deliver reject"`
}

// AdminReviewApplication POST /system/welfare/review
func (h *WelfareHandler) AdminReviewApplication(c *gin.Context) {
	var req adminReviewApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}

	reviewerID := middleware.GetUserID(c)
	mailSummary, err := h.svc.AdminReviewApplication(req.ID, reviewerID, &service.AdminReviewApplicationRequest{
		Action: req.Action,
	})
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, service.MailActionResult{MailAttemptSummary: mailSummary})
}

// ─────────────────────────────────────────────
//  用户端
// ─────────────────────────────────────────────

// GetEligibleWelfares POST /welfare/eligible
func (h *WelfareHandler) GetEligibleWelfares(c *gin.Context) {
	userID := middleware.GetUserID(c)
	result, err := h.svc.GetEligibleWelfares(userID)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

// applyForWelfareRequest 申请福利请求
type applyForWelfareRequest struct {
	WelfareID     uint   `json:"welfare_id" binding:"required"`
	CharacterID   int64  `json:"character_id"`
	EvidenceImage string `json:"evidence_image"`
}

// ApplyForWelfare POST /welfare/apply
func (h *WelfareHandler) ApplyForWelfare(c *gin.Context) {
	var req applyForWelfareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	app, err := h.svc.ApplyForWelfare(userID, &service.ApplyForWelfareRequest{
		WelfareID:     req.WelfareID,
		CharacterID:   req.CharacterID,
		EvidenceImage: req.EvidenceImage,
	})
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, app)
}

// myApplicationsRequest 查询我的申请请求
type myApplicationsRequest struct {
	Current int    `json:"current"`
	Size    int    `json:"size"`
	Status  string `json:"status"`
}

// ListMyApplications POST /welfare/my-applications
func (h *WelfareHandler) ListMyApplications(c *gin.Context) {
	var req myApplicationsRequest
	_ = c.ShouldBindJSON(&req) // current/size/status are optional
	req.Current, req.Size = normalizePagination(req.Current, req.Size, 10, 100)

	userID := middleware.GetUserID(c)
	result, total, err := h.svc.ListMyApplications(userID, req.Current, req.Size, req.Status)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OKWithPage(c, result, total, req.Current, req.Size)
}
