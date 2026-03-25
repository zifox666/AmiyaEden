package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"

	"github.com/gin-gonic/gin"
)

type CorpStructureHandler struct {
	svc *service.CorpStructureService
}

func NewCorpStructureHandler() *CorpStructureHandler {
	return &CorpStructureHandler{
		svc: service.NewCorpStructureService(),
	}
}

// ListStructures POST /operation/corp-structures/list
func (h *CorpStructureHandler) ListStructures(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req service.CorpStructureListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "参数错误: "+err.Error())
		return
	}

	result, err := h.svc.ListCorpStructures(userID, &req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

// GetCorpIDs GET /operation/corp-structures/corps
func (h *CorpStructureHandler) GetCorpIDs(c *gin.Context) {
	userID := middleware.GetUserID(c)

	corpIDs, err := h.svc.GetUserCorpIDs(userID)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, corpIDs)
}
