package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"

	"github.com/gin-gonic/gin"
)

type BadgeHandler struct {
	svc *service.BadgeService
}

func NewBadgeHandler() *BadgeHandler {
	return &BadgeHandler{svc: service.NewBadgeService()}
}

func (h *BadgeHandler) GetBadgeCounts(c *gin.Context) {
	counts, err := h.svc.GetBadgeCounts(middleware.GetUserID(c), middleware.GetUserRoles(c))
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}

	response.OK(c, counts)
}
