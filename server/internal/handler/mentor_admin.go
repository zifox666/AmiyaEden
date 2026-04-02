package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
)

type MentorAdminHandler struct {
	svc         *service.MentorService
	rewardSvc   *service.MentorRewardService
	settingsSvc mentorAdminSettingsService
}

type mentorAdminSettingsService interface {
	GetSettings() service.MentorSettings
	UpdateSettings(cfg service.MentorSettings) (service.MentorSettings, error)
}

func NewMentorAdminHandler() *MentorAdminHandler {
	return &MentorAdminHandler{
		svc:         service.NewMentorService(),
		rewardSvc:   service.NewMentorRewardService(),
		settingsSvc: service.NewMentorSettingsService(),
	}
}

func (h *MentorAdminHandler) ListAllRelationships(c *gin.Context) {
	page, pageSize, err := parsePaginationQuery(c, 20, 200)
	if err != nil {
		response.Fail(c, response.CodeParamError, err.Error())
		return
	}
	result, total, err := h.svc.AdminListAllRelationships(c.Query("status"), c.Query("keyword"), page, pageSize)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OKWithPage(c, result, total, page, pageSize)
}

func (h *MentorAdminHandler) ListRewardDistributions(c *gin.Context) {
	page, pageSize, err := parseLedgerPaginationQuery(c, 200)
	if err != nil {
		response.Fail(c, response.CodeParamError, err.Error())
		return
	}
	result, total, err := h.rewardSvc.ListAdminRewardDistributions(page, pageSize, c.Query("keyword"))
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OKWithPage(c, result, total, page, pageSize)
}

type revokeRelationshipRequest struct {
	RelationshipID uint `json:"relationship_id" binding:"required"`
}

func (h *MentorAdminHandler) RevokeRelationship(c *gin.Context) {
	adminUserID := middleware.GetUserID(c)
	var req revokeRelationshipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "invalid request")
		return
	}
	if err := h.svc.AdminRevokeRelationship(adminUserID, req.RelationshipID); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, gin.H{})
}

func (h *MentorAdminHandler) GetRewardStages(c *gin.Context) {
	stages, err := h.rewardSvc.GetStages()
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, stages)
}

type updateRewardStagesRequest struct {
	Stages []service.MentorRewardStageInput `json:"stages" binding:"required"`
}

type updateMentorSettingsRequest struct {
	MaxCharacterSP    int64 `json:"max_character_sp" binding:"required,gt=0"`
	MaxAccountAgeDays int   `json:"max_account_age_days" binding:"required,gt=0"`
}

func (h *MentorAdminHandler) GetSettings(c *gin.Context) {
	response.OK(c, h.settingsSvc.GetSettings())
}

func (h *MentorAdminHandler) UpdateSettings(c *gin.Context) {
	var req updateMentorSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "invalid request: "+err.Error())
		return
	}

	updated, err := h.settingsSvc.UpdateSettings(service.MentorSettings{
		MaxCharacterSP:    req.MaxCharacterSP,
		MaxAccountAgeDays: req.MaxAccountAgeDays,
	})
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}

	response.OK(c, updated)
}

func (h *MentorAdminHandler) UpdateRewardStages(c *gin.Context) {
	var req updateRewardStagesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "invalid request")
		return
	}
	result, err := h.rewardSvc.UpdateStages(req.Stages)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *MentorAdminHandler) RunRewardProcessing(c *gin.Context) {
	result, err := h.rewardSvc.ProcessRewards(time.Now())
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}
