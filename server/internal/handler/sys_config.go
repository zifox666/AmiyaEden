package handler

import (
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"amiya-eden/internal/utils"
	"amiya-eden/pkg/response"

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
