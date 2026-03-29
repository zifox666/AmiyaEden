package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"

	"github.com/gin-gonic/gin"
)

// FittingsHandler 装配处理器
type FittingsHandler struct {
	svc *service.FittingsService
}

func NewFittingsHandler() *FittingsHandler {
	return &FittingsHandler{
		svc: service.NewFittingsService(),
	}
}

// GetFittings POST /info/fittings
// 获取当前用户名下所有人物的装配列表
func (h *FittingsHandler) GetFittings(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req service.FittingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "参数错误: "+err.Error())
		return
	}

	result, err := h.svc.GetFittings(userID, &req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

// SaveFitting POST /info/fittings/save
// 保存装配（如果传了 fitting_id 就先删后增，否则新增）
func (h *FittingsHandler) SaveFitting(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req service.SaveFittingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "参数错误: "+err.Error())
		return
	}

	result, err := h.svc.SaveFitting(userID, &req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}
