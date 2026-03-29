package handler

import (
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"
	"math"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type NewbroAdminHandler struct {
	reportSvc   *service.NewbroReportService
	syncSvc     *service.CaptainBountySyncService
	rewardSvc   *service.CaptainRewardProcessingService
	settingsSvc *service.NewbroSettingsService
}

func NewNewbroAdminHandler() *NewbroAdminHandler {
	return &NewbroAdminHandler{
		reportSvc:   service.NewNewbroReportService(),
		syncSvc:     service.NewCaptainBountySyncService(),
		rewardSvc:   service.NewCaptainRewardProcessingService(),
		settingsSvc: service.NewNewbroSettingsService(),
	}
}

func (h *NewbroAdminHandler) ListCaptains(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("current", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	result, total, err := h.reportSvc.ListAllCaptainOverviews(page, size, c.Query("keyword"))
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OKWithPage(c, result, total, page, size)
}

func (h *NewbroAdminHandler) GetCaptainDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil || id > math.MaxUint32 {
		response.Fail(c, response.CodeParamError, "invalid user_id")
		return
	}
	result, err := h.reportSvc.GetAdminCaptainDetail(uint(id))
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *NewbroAdminHandler) RunAttributionSync(c *gin.Context) {
	result, err := h.syncSvc.RunSync(time.Now())
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *NewbroAdminHandler) RunRewardProcessing(c *gin.Context) {
	result, err := h.rewardSvc.Run(time.Now())
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

type UpdateNewbroSettingsRequest struct {
	MaxCharacterSP          int64    `json:"max_character_sp" binding:"required,gt=0"`
	MultiCharacterSP        int64    `json:"multi_character_sp" binding:"required,gt=0"`
	MultiCharacterThreshold int      `json:"multi_character_threshold" binding:"required,gt=0"`
	RefreshIntervalDays     int      `json:"refresh_interval_days" binding:"required,gt=0"`
	BonusRate               *float64 `json:"bonus_rate" binding:"required,gte=0"`
}

func (h *NewbroAdminHandler) GetSettings(c *gin.Context) {
	response.OK(c, h.settingsSvc.GetSettings())
}

func (h *NewbroAdminHandler) UpdateSettings(c *gin.Context) {
	var req UpdateNewbroSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "invalid request: "+err.Error())
		return
	}

	updated, err := h.settingsSvc.UpdateSettings(service.NewbroSettings{
		MaxCharacterSP:          req.MaxCharacterSP,
		MultiCharacterSP:        req.MultiCharacterSP,
		MultiCharacterThreshold: req.MultiCharacterThreshold,
		RefreshIntervalDays:     req.RefreshIntervalDays,
		BonusRate:               *req.BonusRate,
	})
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}

	response.OK(c, updated)
}

func (h *NewbroAdminHandler) ListAffiliationHistory(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("current", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	changeStartDate, err := parseOptionalNewbroDate(c.Query("change_start_date"), false)
	if err != nil {
		response.Fail(c, response.CodeParamError, err.Error())
		return
	}
	changeEndDate, err := parseOptionalNewbroDate(c.Query("change_end_date"), true)
	if err != nil {
		response.Fail(c, response.CodeParamError, err.Error())
		return
	}

	result, total, err := h.reportSvc.ListAdminAffiliationHistory(service.AdminAffiliationHistoryListRequest{
		Page:            page,
		PageSize:        size,
		CaptainSearch:   c.Query("captain_search"),
		PlayerSearch:    c.Query("player_search"),
		ChangeStartDate: changeStartDate,
		ChangeEndDate:   changeEndDate,
	})
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OKWithPage(c, result, total, page, size)
}

func (h *NewbroAdminHandler) ListRewardSettlements(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("current", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "200"))
	summary, result, total, err := h.reportSvc.ListAdminRewardSettlements(page, size)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, gin.H{
		"summary":   summary,
		"list":      result,
		"total":     total,
		"page":      page,
		"page_size": size,
	})
}
