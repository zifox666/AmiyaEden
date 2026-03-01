package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"

	"github.com/gin-gonic/gin"
)

// EveInfoHandler EVE 角色信息处理器
type EveInfoHandler struct {
	svc *service.EveInfoService
}

func NewEveInfoHandler() *EveInfoHandler {
	return &EveInfoHandler{
		svc: service.NewEveInfoService(),
	}
}

// GetWalletJournal POST /info/wallet
// 获取指定角色的钱包余额和流水记录
func (h *EveInfoHandler) GetWalletJournal(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req service.InfoWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "参数错误: "+err.Error())
		return
	}

	result, err := h.svc.GetWalletJournal(userID, &req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

// GetCharacterSkills POST /info/skills
// 获取指定角色的技能列表和学习队列
func (h *EveInfoHandler) GetCharacterSkills(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req service.InfoSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "参数错误: "+err.Error())
		return
	}

	result, err := h.svc.GetCharacterSkills(userID, &req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}
