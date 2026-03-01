package handler

import (
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type EveSkillHandler struct {
	svc *service.EveSkillService
}

func NewEveSkillHandler() *EveSkillHandler {
	return &EveSkillHandler{
		svc: service.NewEveSkillService(),
	}
}

// GetEveCharacterSkills GET /eve/character/:id/skills
// 获取角色技能信息（总 SP + 技能列表 + 每个 group_id 的技能数量）
func (h *EveSkillHandler) GetEveCharacterSkills(c *gin.Context) {
	characterID := c.Param("id")
	if characterID == "" {
		response.Fail(c, response.CodeParamError, "Invalid character ID")
		return
	}

	characterIDInt, err := strconv.Atoi(characterID)
	if err != nil {
		response.Fail(c, response.CodeParamError, "Invalid character ID")
		return
	}

	result, err := h.svc.GetEveCharacterSkills(characterIDInt)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}
