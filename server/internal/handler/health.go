package handler

import (
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct{ svc *service.HealthService }

func NewHealthHandler() *HealthHandler { return &HealthHandler{svc: service.NewHealthService()} }

func (h *HealthHandler) Check(c *gin.Context) {
	status := h.svc.Check()
	if !status.Healthy() {
		response.Unavailable(c, "依赖服务不可用", status)
		return
	}
	response.OK(c, status)
}
