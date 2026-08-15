package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/model"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"

	"github.com/gin-gonic/gin"
)

type ATSetHandler struct{ svc *service.ATSetService }

func NewATSetHandler() *ATSetHandler { return &ATSetHandler{svc: service.NewATSetService()} }
func atSetAdmin(c *gin.Context) bool {
	return model.ContainsAnyRole(middleware.GetUserRoles(c), model.RoleAdmin, model.RoleSuperAdmin)
}
func (h *ATSetHandler) Create(c *gin.Context) {
	var input service.ATSetInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	result, err := h.svc.Create(middleware.GetUserID(c), input)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}
func (h *ATSetHandler) Update(c *gin.Context) {
	id, ok := atUintParam(c, "id")
	if !ok {
		return
	}
	var input service.ATSetInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	result, err := h.svc.Update(middleware.GetUserID(c), id, input, atSetAdmin(c))
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}
func (h *ATSetHandler) Get(c *gin.Context) {
	id, ok := atUintParam(c, "id")
	if !ok {
		return
	}
	result, err := h.svc.Get(id)
	if err != nil {
		response.Fail(c, response.CodeNotFound, err.Error())
		return
	}
	response.OK(c, result)
}
func (h *ATSetHandler) List(c *gin.Context) {
	result, err := h.svc.List()
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}
func (h *ATSetHandler) Delete(c *gin.Context) {
	id, ok := atUintParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Delete(middleware.GetUserID(c), id, atSetAdmin(c)); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}
func (h *ATSetHandler) EFT(c *gin.Context) {
	id, ok := atUintParam(c, "id")
	if !ok {
		return
	}
	result, err := h.svc.GenerateEFT(id)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *ATSetHandler) ParseEFT(c *gin.Context) {
	var input struct {
		EFT string `json:"eft" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, response.CodeParamError, "请粘贴 EFT 配置")
		return
	}
	result, err := h.svc.ParseEFT(input.EFT)
	if err != nil {
		response.Fail(c, response.CodeParamError, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *ATSetHandler) ListFits(c *gin.Context) {
	result, err := h.svc.ListFits()
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *ATSetHandler) GetFit(c *gin.Context) {
	id, ok := atUintParam(c, "id")
	if !ok {
		return
	}
	result, err := h.svc.GetFit(id)
	if err != nil {
		response.Fail(c, response.CodeNotFound, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *ATSetHandler) SaveFit(c *gin.Context) {
	var input service.ATFitInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	var id uint
	if c.Param("id") != "" {
		value, ok := atUintParam(c, "id")
		if !ok {
			return
		}
		id = value
	}
	result, err := h.svc.SaveFit(middleware.GetUserID(c), id, input, atAdmin(c))
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *ATSetHandler) DeleteFit(c *gin.Context) {
	id, ok := atUintParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.DeleteFit(middleware.GetUserID(c), id, atAdmin(c)); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}
