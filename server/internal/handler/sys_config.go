package handler

import (
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"amiya-eden/pkg/response"
	"github.com/gin-gonic/gin"
	"strconv"
)

type SysConfigHandler struct {
	repo *repository.SysConfigRepository
}

func NewSysConfigHandler() *SysConfigHandler {
	return &SysConfigHandler{
		repo: repository.NewSysConfigRepository(),
	}
}

type BasicConfigResponse struct {
	CorpID    int64  `json:"corp_id"`
	SiteTitle string `json:"site_title"`
}

type UpdateBasicConfigRequest struct {
	CorpID    *int64  `json:"corp_id"`
	SiteTitle *string `json:"site_title"`
}

func (h *SysConfigHandler) GetBasicConfig(c *gin.Context) {
	defaultCorpID := strconv.FormatInt(model.SysConfigDefaultCorpID, 10)
	corpIDStr, _ := h.repo.Get(model.SysConfigCorpID, defaultCorpID)
	corpID, err := strconv.ParseInt(corpIDStr, 10, 64)
	if err != nil {
		corpID = model.SysConfigDefaultCorpID
	}
	siteTitle, _ := h.repo.Get(model.SysConfigSiteTitle, model.SysConfigDefaultSiteTitle)

	response.OK(c, BasicConfigResponse{
		CorpID:    corpID,
		SiteTitle: siteTitle,
	})
}

func (h *SysConfigHandler) UpdateBasicConfig(c *gin.Context) {
	var req UpdateBasicConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误")
		return
	}

	if req.CorpID != nil {
		if err := h.repo.Set(model.SysConfigCorpID, strconv.FormatInt(*req.CorpID, 10), "军团ID"); err != nil {
			response.Fail(c, response.CodeBizError, "更新军团ID失败")
			return
		}
	}

	if req.SiteTitle != nil {
		if err := h.repo.Set(model.SysConfigSiteTitle, *req.SiteTitle, "网站标题"); err != nil {
			response.Fail(c, response.CodeBizError, "更新网站标题失败")
			return
		}
	}

	response.OK(c, nil)
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
