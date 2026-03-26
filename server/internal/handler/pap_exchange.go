package handler

import (
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"

	"github.com/gin-gonic/gin"
)

// PAPExchangeHandler PAP 兑换汇率 HTTP 处理器
type PAPExchangeHandler struct {
	svc *service.PAPExchangeService
}

func NewPAPExchangeHandler() *PAPExchangeHandler {
	return &PAPExchangeHandler{svc: service.NewPAPExchangeService()}
}

// GetRates  GET /system/pap/exchange-rates
// 查询 PAP 类型兑换汇率列表
func (h *PAPExchangeHandler) GetRates(c *gin.Context) {
	rates, err := h.svc.GetRates()
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, rates)
}

// SetRates  PUT /system/pap/exchange-rates
// 批量更新 PAP 类型兑换汇率
func (h *PAPExchangeHandler) SetRates(c *gin.Context) {
	var req []service.SetRateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}
	if err := h.svc.SetRates(req); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	rates, _ := h.svc.GetRates()
	response.OK(c, rates)
}
