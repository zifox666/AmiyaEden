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

// GetRates  GET /system/pap-exchange/rates
// 查询 PAP 兑换配置（汇率 + FC 工资）
func (h *PAPExchangeHandler) GetRates(c *gin.Context) {
	config, err := h.svc.GetConfig()
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, config)
}

// SetRates  PUT /system/pap-exchange/rates
// 批量更新 PAP 兑换配置（汇率 + FC 工资）
func (h *PAPExchangeHandler) SetRates(c *gin.Context) {
	var req service.UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}
	config, err := h.svc.UpdateConfig(&req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, config)
}
