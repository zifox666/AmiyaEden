package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"

	"github.com/gin-gonic/gin"
)

// DashboardHandler 工作台 HTTP 处理器
type DashboardHandler struct {
	svc *service.DashboardService
}

func NewDashboardHandler() *DashboardHandler {
	return &DashboardHandler{
		svc: service.NewDashboardService(),
	}
}

// GetDashboard POST /dashboard
// 获取工作台所有数据（卡片统计 + 舰队参与 + PAP 月度统计 + 补损列表）
func (h *DashboardHandler) GetDashboard(c *gin.Context) {
	userID := middleware.GetUserID(c)

	result, err := h.svc.GetDashboard(userID)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}
