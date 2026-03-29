package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"

	"github.com/gin-gonic/gin"
)

// NpcKillHandler NPC 刷怪报表处理器
type NpcKillHandler struct {
	svc *service.NpcKillService
}

func NewNpcKillHandler() *NpcKillHandler {
	return &NpcKillHandler{
		svc: service.NewNpcKillService(),
	}
}

// GetNpcKills POST /info/npc-kills
// 获取当前用户指定人物的刷怪报表
func (h *NpcKillHandler) GetNpcKills(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req service.NpcKillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "参数错误: "+err.Error())
		return
	}

	result, err := h.svc.GetNpcKills(userID, &req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

// GetAllNpcKills POST /info/npc-kills/all
// 获取当前用户名下所有人物的汇总刷怪报表
func (h *NpcKillHandler) GetAllNpcKills(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req service.NpcKillAllRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "参数错误: "+err.Error())
		return
	}

	result, err := h.svc.GetAllNpcKills(userID, &req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

// GetCorpNpcKills POST /corp/npc-kills
// 获取公司内所有成员的刷怪报表（管理员）
func (h *NpcKillHandler) GetCorpNpcKills(c *gin.Context) {
	var req service.NpcKillCorpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "参数错误: "+err.Error())
		return
	}

	result, err := h.svc.GetCorpNpcKills(&req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}
