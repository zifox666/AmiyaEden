package handler

import (
	"errors"
	"io"
	"net/http"
	"strconv"

	"amiya-eden/internal/middleware"
	"amiya-eden/internal/model"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"

	"github.com/gin-gonic/gin"
)

type CombatLogHandler struct{ svc *service.CombatLogService }

func NewCombatLogHandler() *CombatLogHandler {
	return &CombatLogHandler{svc: service.NewCombatLogService()}
}

func (h *CombatLogHandler) Upload(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, service.MaxCombatLogSize+(1<<20))
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Fail(c, response.CodeParamError, "获取日志文件失败")
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, service.MaxCombatLogSize+1))
	if err != nil {
		response.Fail(c, response.CodeParamError, "读取日志文件失败")
		return
	}
	if len(data) > service.MaxCombatLogSize {
		response.Fail(c, response.CodeParamError, "日志大小不能超过 20 MB")
		return
	}
	result, err := h.svc.Upload(middleware.GetUserID(c), header.Filename, data)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *CombatLogHandler) List(c *gin.Context) {
	page := positiveQueryInt(c, "page", 1)
	pageSize := positiveQueryInt(c, "page_size", 20)
	if pageSize > 100 {
		pageSize = 100
	}
	includeAll := c.Query("scope") == "all"
	result, err := h.svc.List(middleware.GetUserID(c), includeAll, combatLogAdmin(c), page, pageSize)
	if err != nil {
		code := response.CodeBizError
		if errors.Is(err, service.ErrCombatLogForbidden) {
			code = response.CodeForbidden
		}
		response.Fail(c, code, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *CombatLogHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Fail(c, response.CodeParamError, "日志分析 ID 无效")
		return
	}
	result, err := h.svc.Get(middleware.GetUserID(c), uint(id), combatLogAdmin(c))
	if err != nil {
		code := response.CodeBizError
		if errors.Is(err, service.ErrCombatLogForbidden) {
			code = response.CodeForbidden
		} else if errors.Is(err, service.ErrCombatLogNotFound) {
			code = response.CodeNotFound
		}
		response.Fail(c, code, err.Error())
		return
	}
	response.OK(c, result)
}

func positiveQueryInt(c *gin.Context, key string, fallback int) int {
	value, err := strconv.Atoi(c.Query(key))
	if err != nil || value < 1 {
		return fallback
	}
	return value
}

func combatLogAdmin(c *gin.Context) bool {
	return model.ContainsAnyRole(middleware.GetUserRoles(c), model.RoleAdmin, model.RoleSuperAdmin)
}
