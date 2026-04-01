package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"

	"github.com/gin-gonic/gin"
)

type MentorMentorHandler struct {
	svc       *service.MentorService
	rewardSvc *service.MentorRewardService
}

func NewMentorMentorHandler() *MentorMentorHandler {
	return &MentorMentorHandler{
		svc:       service.NewMentorService(),
		rewardSvc: service.NewMentorRewardService(),
	}
}

func (h *MentorMentorHandler) ListPendingApplications(c *gin.Context) {
	userID := middleware.GetUserID(c)
	result, err := h.svc.ListPendingApplications(userID)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *MentorMentorHandler) ListMyMentees(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, pageSize, err := parsePaginationQuery(c, 20, 100)
	if err != nil {
		response.Fail(c, response.CodeParamError, err.Error())
		return
	}
	result, total, err := h.svc.ListMyMentees(userID, c.DefaultQuery("status", "active"), page, pageSize)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OKWithPage(c, result, total, page, pageSize)
}

func (h *MentorMentorHandler) GetRewardStages(c *gin.Context) {
	stages, err := h.rewardSvc.GetStages()
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, stages)
}

type mentorActionRequest struct {
	RelationshipID uint `json:"relationship_id" binding:"required"`
}

func (h *MentorMentorHandler) AcceptApplication(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req mentorActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "invalid request")
		return
	}
	if err := h.svc.AcceptApplication(userID, req.RelationshipID); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, gin.H{})
}

func (h *MentorMentorHandler) RejectApplication(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req mentorActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "invalid request")
		return
	}
	if err := h.svc.RejectApplication(userID, req.RelationshipID); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, gin.H{})
}
