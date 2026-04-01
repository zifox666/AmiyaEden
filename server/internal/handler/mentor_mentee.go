package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"

	"github.com/gin-gonic/gin"
)

type MentorMenteeHandler struct {
	svc *service.MentorService
}

func NewMentorMenteeHandler() *MentorMenteeHandler {
	return &MentorMenteeHandler{svc: service.NewMentorService()}
}

func (h *MentorMenteeHandler) ListMentors(c *gin.Context) {
	userID := middleware.GetUserID(c)
	result, err := h.svc.ListMentorCandidates(userID)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *MentorMenteeHandler) GetMyStatus(c *gin.Context) {
	userID := middleware.GetUserID(c)
	result, err := h.svc.GetMyMenteeStatus(userID)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

type applyForMentorRequest struct {
	MentorUserID uint `json:"mentor_user_id" binding:"required"`
}

func (h *MentorMenteeHandler) ApplyForMentor(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req applyForMentorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "invalid request")
		return
	}
	result, err := h.svc.ApplyForMentor(userID, req.MentorUserID)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}
