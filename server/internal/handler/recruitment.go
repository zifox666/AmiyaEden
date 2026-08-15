package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/model"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RecruitmentHandler struct{ svc *service.RecruitmentService }

func NewRecruitmentHandler() *RecruitmentHandler {
	return &RecruitmentHandler{svc: service.NewRecruitmentService()}
}

func recruitmentAdmin(c *gin.Context) bool {
	return model.ContainsAnyRole(middleware.GetUserRoles(c), model.RoleAdmin, model.RoleSuperAdmin)
}

func recruitmentID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeParamError, "无效 ID")
		return 0, false
	}
	return uint(id), true
}

func (h *RecruitmentHandler) PublicContext(c *gin.Context) {
	result, err := h.svc.GetPublicContext(c.Param("token"))
	if err != nil {
		response.Fail(c, response.CodeNotFound, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *RecruitmentHandler) Submit(c *gin.Context) {
	var input service.SubmitRecruitmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	result, err := h.svc.Submit(middleware.GetUserID(c), c.Param("token"), input)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *RecruitmentHandler) CreateQuestionnaire(c *gin.Context) {
	var input service.RecruitmentQuestionnaireInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	result, err := h.svc.CreateQuestionnaire(middleware.GetUserID(c), input)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *RecruitmentHandler) UpdateQuestionnaire(c *gin.Context) {
	id, ok := recruitmentID(c)
	if !ok {
		return
	}
	var input service.RecruitmentQuestionnaireInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	result, err := h.svc.UpdateQuestionnaire(middleware.GetUserID(c), id, input, c.Query("publish") == "true", recruitmentAdmin(c))
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *RecruitmentHandler) GetQuestionnaire(c *gin.Context) {
	id, ok := recruitmentID(c)
	if !ok {
		return
	}
	result, err := h.svc.GetQuestionnaire(middleware.GetUserID(c), id, recruitmentAdmin(c))
	if err != nil {
		response.Fail(c, response.CodeForbidden, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *RecruitmentHandler) ListQuestionnaires(c *gin.Context) {
	result, err := h.svc.ListQuestionnaires(middleware.GetUserID(c), recruitmentAdmin(c))
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *RecruitmentHandler) CreateLink(c *gin.Context) {
	var input service.CreateRecruitmentLinkInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	result, err := h.svc.CreateLink(middleware.GetUserID(c), input, recruitmentAdmin(c))
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *RecruitmentHandler) ListLinks(c *gin.Context) {
	result, err := h.svc.ListLinks(middleware.GetUserID(c), recruitmentAdmin(c))
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *RecruitmentHandler) SetLinkActive(c *gin.Context) {
	id, ok := recruitmentID(c)
	if !ok {
		return
	}
	var input struct {
		Active bool `json:"active"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	if err := h.svc.SetLinkActive(middleware.GetUserID(c), id, input.Active, recruitmentAdmin(c)); err != nil {
		response.Fail(c, response.CodeForbidden, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *RecruitmentHandler) ListApplications(c *gin.Context) {
	result, err := h.svc.ListApplications(middleware.GetUserID(c), recruitmentAdmin(c))
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *RecruitmentHandler) ListMyApplications(c *gin.Context) {
	result, err := h.svc.ListMyApplications(middleware.GetUserID(c))
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *RecruitmentHandler) Withdraw(c *gin.Context) {
	id, ok := recruitmentID(c)
	if !ok {
		return
	}
	if err := h.svc.Withdraw(middleware.GetUserID(c), id); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *RecruitmentHandler) GetApplication(c *gin.Context) {
	id, ok := recruitmentID(c)
	if !ok {
		return
	}
	result, err := h.svc.GetApplication(middleware.GetUserID(c), id, recruitmentAdmin(c))
	if err != nil {
		response.Fail(c, response.CodeForbidden, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *RecruitmentHandler) Review(c *gin.Context) {
	id, ok := recruitmentID(c)
	if !ok {
		return
	}
	var input struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	if err := h.svc.Review(middleware.GetUserID(c), id, input.Status, input.Reason, recruitmentAdmin(c)); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *RecruitmentHandler) Cleanup(c *gin.Context) {
	var input struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	count, err := h.svc.Cleanup(middleware.GetUserID(c), input.Reason)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, gin.H{"applications_cleaned": count})
}
