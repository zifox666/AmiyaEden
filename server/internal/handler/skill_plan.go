package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SkillPlanHandler struct {
	svc *service.SkillPlanService
}

func NewSkillPlanHandler() *SkillPlanHandler {
	return &SkillPlanHandler{svc: service.NewSkillPlanService()}
}

// CreateSkillPlan 创建技能规划
func (h *SkillPlanHandler) CreateSkillPlan(c *gin.Context) {
	var req service.CreateSkillPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}
	userID := middleware.GetUserID(c)
	result, err := h.svc.Create(userID, &req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

// UpdateSkillPlan 更新技能规划
func (h *SkillPlanHandler) UpdateSkillPlan(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeParamError, "无效的规划ID")
		return
	}
	var req service.UpdateSkillPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}
	result, err := h.svc.Update(uint(id), &req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

// DeleteSkillPlan 删除技能规划
func (h *SkillPlanHandler) DeleteSkillPlan(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeParamError, "无效的规划ID")
		return
	}
	if err := h.svc.Delete(uint(id)); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}

// GetSkillPlan 获取技能规划详情
func (h *SkillPlanHandler) GetSkillPlan(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeParamError, "无效的规划ID")
		return
	}
	lang := c.DefaultQuery("lang", "zh")
	result, err := h.svc.GetByID(uint(id), lang)
	if err != nil {
		response.Fail(c, response.CodeNotFound, err.Error())
		return
	}
	response.OK(c, result)
}

// ListSkillPlans 分页查询技能规划
func (h *SkillPlanHandler) ListSkillPlans(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("current", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	records, total, err := h.svc.List(page, size)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OKWithPage(c, records, total, page, size)
}

// ListAllSkillPlans 查询全部技能规划（下拉用）
func (h *SkillPlanHandler) ListAllSkillPlans(c *gin.Context) {
	records, err := h.svc.ListAll()
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, records)
}

// CheckAllCharacters 检查所有角色的技能规划完成情况
func (h *SkillPlanHandler) CheckAllCharacters(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeParamError, "无效的规划ID")
		return
	}
	lang := c.DefaultQuery("lang", "zh")
	result, err := h.svc.CheckAllCharacters(uint(id), lang)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

// CheckUserCharacters 检查当前用户角色的技能规划完成情况
func (h *SkillPlanHandler) CheckUserCharacters(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeParamError, "无效的规划ID")
		return
	}
	userID := middleware.GetUserID(c)
	lang := c.DefaultQuery("lang", "zh")
	result, err := h.svc.CheckUserCharacters(uint(id), userID, lang)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}
