package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type NewbroCaptainHandler struct {
	reportSvc *service.NewbroReportService
	affSvc    *service.NewbroAffiliationService
}

func NewNewbroCaptainHandler() *NewbroCaptainHandler {
	return &NewbroCaptainHandler{
		reportSvc: service.NewNewbroReportService(),
		affSvc:    service.NewNewbroAffiliationService(),
	}
}

func (h *NewbroCaptainHandler) GetOverview(c *gin.Context) {
	userID := middleware.GetUserID(c)
	result, err := h.reportSvc.GetCaptainOverview(userID)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *NewbroCaptainHandler) GetPlayers(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, pageSize, err := parsePaginationQuery(c, 20, 100)
	if err != nil {
		response.Fail(c, response.CodeParamError, err.Error())
		return
	}
	status := c.DefaultQuery("status", "all")
	result, total, err := h.reportSvc.ListCaptainPlayers(userID, status, page, pageSize)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OKWithPage(c, result, total, page, pageSize)
}

func (h *NewbroCaptainHandler) GetAttributions(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, pageSize, err := parsePaginationQuery(c, 20, 100)
	if err != nil {
		response.Fail(c, response.CodeParamError, err.Error())
		return
	}
	playerUserID, err := parseOptionalUintQueryParam("player_user_id", c.Query("player_user_id"))
	if err != nil {
		response.Fail(c, response.CodeParamError, err.Error())
		return
	}
	startDate, err := parseOptionalNewbroDate(c.Query("start_date"), false)
	if err != nil {
		response.Fail(c, response.CodeParamError, err.Error())
		return
	}
	endDate, err := parseOptionalNewbroDate(c.Query("end_date"), true)
	if err != nil {
		response.Fail(c, response.CodeParamError, err.Error())
		return
	}
	summary, result, total, err := h.reportSvc.ListCaptainAttributions(userID, service.CaptainAttributionListRequest{
		Page:         page,
		PageSize:     pageSize,
		PlayerUserID: playerUserID,
		RefType:      c.Query("ref_type"),
		StartDate:    startDate,
		EndDate:      endDate,
	})
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, gin.H{
		"summary":   summary,
		"list":      result,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func parseOptionalUintQueryParam(field, raw string) (*uint, error) {
	if raw == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || parsed > math.MaxUint32 {
		return nil, fmt.Errorf("invalid parameter %s", field)
	}
	value := uint(parsed)
	return &value, nil
}

func (h *NewbroCaptainHandler) GetRewardSettlements(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, pageSize, err := parseLedgerPaginationQuery(c, 20)
	if err != nil {
		response.Fail(c, response.CodeParamError, err.Error())
		return
	}
	summary, result, total, err := h.reportSvc.ListCaptainRewardSettlements(userID, page, pageSize)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, gin.H{
		"summary":   summary,
		"list":      result,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *NewbroCaptainHandler) ListEligiblePlayers(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, pageSize, err := parsePaginationQuery(c, 20, 100)
	if err != nil {
		response.Fail(c, response.CodeParamError, err.Error())
		return
	}

	result, total, err := h.affSvc.ListCaptainEligiblePlayers(userID, c.Query("keyword"), page, pageSize)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OKWithPage(c, result, total, page, pageSize)
}

type captainEnrollPlayerRequest struct {
	PlayerUserID uint `json:"player_user_id" binding:"required"`
}

func (h *NewbroCaptainHandler) EnrollPlayer(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req captainEnrollPlayerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "invalid request")
		return
	}

	result, err := h.affSvc.EnrollPlayer(userID, req.PlayerUserID)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

type captainEndAffiliationRequest struct {
	PlayerUserID uint `json:"player_user_id" binding:"required"`
}

func (h *NewbroCaptainHandler) EndAffiliation(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req captainEndAffiliationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "invalid request")
		return
	}
	if err := h.affSvc.EndAffiliation(userID, req.PlayerUserID); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, gin.H{})
}

func parseOptionalNewbroDate(raw string, endOfDay bool) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}
	value, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return nil, fmt.Errorf("invalid date: expected YYYY-MM-DD")
	}
	if endOfDay {
		v := value.Add(24*time.Hour - time.Nanosecond)
		return &v, nil
	}
	return &value, nil
}
