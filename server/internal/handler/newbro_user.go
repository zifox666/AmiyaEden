package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"

	"github.com/gin-gonic/gin"
)

type NewbroUserHandler struct {
	affSvc *service.NewbroAffiliationService
}

func NewNewbroUserHandler() *NewbroUserHandler {
	return &NewbroUserHandler{affSvc: service.NewNewbroAffiliationService()}
}

func (h *NewbroUserHandler) ListCaptains(c *gin.Context) {
	userID := middleware.GetUserID(c)
	result, err := h.affSvc.ListCaptainCandidates(userID)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *NewbroUserHandler) GetMyAffiliation(c *gin.Context) {
	userID := middleware.GetUserID(c)
	result, err := h.affSvc.GetMyAffiliation(userID)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *NewbroUserHandler) ListMyAffiliationHistory(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, pageSize, err := parseLedgerPaginationQuery(c, 200)
	if err != nil {
		response.Fail(c, response.CodeParamError, err.Error())
		return
	}
	result, total, err := h.affSvc.ListMyAffiliationHistory(userID, page, pageSize)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OKWithPage(c, result, total, page, pageSize)
}

type selectCaptainRequest struct {
	CaptainUserID uint `json:"captain_user_id" binding:"required"`
}

func (h *NewbroUserHandler) SelectCaptain(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req selectCaptainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "invalid request")
		return
	}
	result, err := h.affSvc.SelectCaptain(userID, req.CaptainUserID)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *NewbroUserHandler) EndAffiliation(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if err := h.affSvc.EndAffiliation(userID, userID); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, gin.H{})
}
