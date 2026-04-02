package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"amiya-eden/internal/utils"
	"amiya-eden/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SysConfigHandler struct {
	repo *repository.SysConfigRepository
}

func NewSysConfigHandler() *SysConfigHandler {
	return &SysConfigHandler{
		repo: repository.NewSysConfigRepository(),
	}
}

func (h *SysConfigHandler) GetBasicConfig(c *gin.Context) {
	response.OK(c, model.DefaultSystemIdentity())
}

// SDEConfigResponse SDE 配置响应
type SDEConfigResponse struct {
	APIKey      string `json:"api_key"`
	Proxy       string `json:"proxy"`
	DownloadURL string `json:"download_url"`
}

// UpdateSDEConfigRequest 更新 SDE 配置请求
type UpdateSDEConfigRequest struct {
	APIKey      *string `json:"api_key"`
	Proxy       *string `json:"proxy"`
	DownloadURL *string `json:"download_url"`
}

func (h *SysConfigHandler) GetSDEConfig(c *gin.Context) {
	apiKey, _ := h.repo.Get(model.SysConfigSDEAPIKey, model.SysConfigDefaultSDEAPIKey)
	proxy, _ := h.repo.Get(model.SysConfigSDEProxy, model.SysConfigDefaultSDEProxy)
	downloadURL, _ := h.repo.Get(model.SysConfigSDEDownloadURL, model.SysConfigDefaultSDEDownloadURL)

	response.OK(c, SDEConfigResponse{
		APIKey:      apiKey,
		Proxy:       proxy,
		DownloadURL: downloadURL,
	})
}

func (h *SysConfigHandler) UpdateSDEConfig(c *gin.Context) {
	var req UpdateSDEConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}

	if req.APIKey != nil {
		if err := h.repo.Set(model.SysConfigSDEAPIKey, *req.APIKey, "SDE 查询 API Key"); err != nil {
			response.Fail(c, response.CodeBizError, "更新 API Key 失败")
			return
		}
	}

	if req.Proxy != nil {
		if err := h.repo.Set(model.SysConfigSDEProxy, *req.Proxy, "SDE 下载代理"); err != nil {
			response.Fail(c, response.CodeBizError, "更新代理配置失败")
			return
		}
	}

	if req.DownloadURL != nil {
		if err := h.repo.Set(model.SysConfigSDEDownloadURL, *req.DownloadURL, "SDE 下载地址"); err != nil {
			response.Fail(c, response.CodeBizError, "更新下载地址失败")
			return
		}
	}

	response.OK(c, nil)
}

type AllowCorporationsResponse struct {
	AllowCorporations []int64 `json:"allow_corporations"`
}

type UpdateAllowCorporationsRequest struct {
	AllowCorporations []int64 `json:"allow_corporations"`
}

type CharacterESIRestrictionConfigResponse struct {
	EnforceCharacterESIRestriction bool `json:"enforce_character_esi_restriction"`
}

type UpdateCharacterESIRestrictionConfigRequest struct {
	EnforceCharacterESIRestriction *bool `json:"enforce_character_esi_restriction"`
}

func (h *SysConfigHandler) GetAllowCorporations(c *gin.Context) {
	response.OK(c, AllowCorporationsResponse{
		AllowCorporations: utils.GetAllowCorporations(),
	})
}

func (h *SysConfigHandler) UpdateAllowCorporations(c *gin.Context) {
	var req UpdateAllowCorporationsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	if err := utils.ValidateAllowCorporations(req.AllowCorporations); err != nil {
		response.Fail(c, response.CodeParamError, "军团 ID 必须为正整数")
		return
	}

	normalizedAllowCorporations := utils.NormalizeAllowCorporations(req.AllowCorporations)
	if err := h.repo.SetInt64Slice(model.SysConfigAllowCorporations, normalizedAllowCorporations, "允许访问的公司 ID 列表"); err != nil {
		response.Fail(c, response.CodeBizError, "更新允许的军团列表失败")
		return
	}

	utils.InvalidateAllowCorporationsCache()

	response.OK(c, nil)
}

func (h *SysConfigHandler) GetCharacterESIRestrictionConfig(c *gin.Context) {
	response.OK(c, CharacterESIRestrictionConfigResponse{
		EnforceCharacterESIRestriction: h.repo.GetBool(
			model.SysConfigEnforceCharacterESIRestriction,
			model.SysConfigDefaultEnforceCharacterESIRestriction,
		),
	})
}

func (h *SysConfigHandler) UpdateCharacterESIRestrictionConfig(c *gin.Context) {
	if !model.IsSuperAdmin(middleware.GetUserRoles(c)) {
		response.Fail(c, response.CodeForbidden, "仅超级管理员可修改该配置")
		return
	}

	var req UpdateCharacterESIRestrictionConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}
	if req.EnforceCharacterESIRestriction == nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}

	if err := h.repo.Set(
		model.SysConfigEnforceCharacterESIRestriction,
		strconv.FormatBool(*req.EnforceCharacterESIRestriction),
		"是否强制限制失效人物 ESI 停留在人物页面",
	); err != nil {
		response.Fail(c, response.CodeBizError, "更新人物 ESI 限制配置失败")
		return
	}

	response.OK(c, nil)
}
