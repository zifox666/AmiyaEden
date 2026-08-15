package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"
	"io"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ATScoreHandler struct{ svc *service.ATScoreService }

func NewATScoreHandler() *ATScoreHandler { return &ATScoreHandler{svc: service.NewATScoreService()} }

func atUintParam(c *gin.Context, name string) (uint, bool) {
	value, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeParamError, "无效 ID")
		return 0, false
	}
	return uint(value), true
}

func (h *ATScoreHandler) Import(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, response.CodeParamError, "请选择 XLSX 文件")
		return
	}
	if file.Size > 20<<20 {
		response.Fail(c, response.CodeParamError, "XLSX 文件不能超过 20MB")
		return
	}
	opened, err := file.Open()
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	defer opened.Close()
	data, err := io.ReadAll(opened)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	result, err := h.svc.Import(middleware.GetUserID(c), data, c.PostForm("source_url"))
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *ATScoreHandler) Publish(c *gin.Context) {
	id, ok := atUintParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Publish(id); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}
func (h *ATScoreHandler) List(c *gin.Context) {
	result, err := h.svc.List()
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}
func (h *ATScoreHandler) Get(c *gin.Context) {
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
func (h *ATScoreHandler) Validate(c *gin.Context) {
	var input struct {
		RuleVersionID uint    `json:"rule_version_id"`
		Format        string  `json:"format"`
		TypeIDs       []int64 `json:"type_ids"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	result, err := h.svc.ValidateRoster(input.RuleVersionID, input.Format, input.TypeIDs)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}
