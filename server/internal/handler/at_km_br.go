package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"
	"github.com/gin-gonic/gin"
)

type ATKillmailHandler struct{ svc *service.ATKillmailService }

func NewATKillmailHandler() *ATKillmailHandler {
	return &ATKillmailHandler{svc: service.NewATKillmailService()}
}
func (h *ATKillmailHandler) Submit(c *gin.Context) {
	var input struct {
		Text string `json:"text"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	result, err := h.svc.Submit(middleware.GetUserID(c), input.Text)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}
func (h *ATKillmailHandler) Pending(c *gin.Context) {
	result, err := h.svc.Pending()
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}
func (h *ATKillmailHandler) List(c *gin.Context) {
	result, err := h.svc.List()
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}
func (h *ATKillmailHandler) Get(c *gin.Context) {
	result, err := h.svc.Get(c.Param("id"))
	if err != nil {
		response.Fail(c, response.CodeNotFound, err.Error())
		return
	}
	response.OK(c, result)
}
func (h *ATKillmailHandler) Suggestions(c *gin.Context) {
	result, err := h.svc.Suggestions(c.Param("id"))
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}
func (h *ATKillmailHandler) Assign(c *gin.Context) {
	var input struct {
		MatchID uint `json:"match_id"`
		RoundID uint `json:"round_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	status, err := h.svc.Assign(middleware.GetUserID(c), c.Param("id"), input.MatchID, input.RoundID, atAdmin(c))
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, gin.H{"status": status})
}
func (h *ATKillmailHandler) Resolve(c *gin.Context) {
	id, ok := atUintParam(c, "id")
	if !ok {
		return
	}
	var input struct {
		ValidKMID string `json:"valid_km_id"`
		Reason    string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	if err := h.svc.ResolveConflict(middleware.GetUserID(c), id, input.ValidKMID, input.Reason, atAdmin(c)); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}

type ATBRHandler struct{ svc *service.ATBRService }

func NewATBRHandler() *ATBRHandler { return &ATBRHandler{svc: service.NewATBRService()} }
func (h *ATBRHandler) Build(c *gin.Context) {
	id, ok := atUintParam(c, "id")
	if !ok {
		return
	}
	var input struct {
		RevisionReason string `json:"revision_reason"`
	}
	_ = c.ShouldBindJSON(&input)
	result, err := h.svc.Build(middleware.GetUserID(c), id, input.RevisionReason, atAdmin(c))
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}
func (h *ATBRHandler) Latest(c *gin.Context) {
	id, ok := atUintParam(c, "id")
	if !ok {
		return
	}
	result, err := h.svc.GetLatest(id)
	if err != nil {
		response.Fail(c, response.CodeNotFound, err.Error())
		return
	}
	response.OK(c, result)
}
func (h *ATBRHandler) Get(c *gin.Context) {
	id, ok := atUintParam(c, "id")
	if !ok {
		return
	}
	result, err := h.svc.Get(id)
	if err != nil {
		response.Fail(c, response.CodeNotFound, err.Error())
		return
	}
	response.OK(c, result)
}
func (h *ATBRHandler) List(c *gin.Context) {
	result, err := h.svc.List()
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}
