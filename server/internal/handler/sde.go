package handler

import (
	"amiya-eden/internal/service"
	"amiya-eden/pkg/eve/esi"
	"amiya-eden/pkg/response"
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
)

// SdeHandler SDE HTTP 处理器
type SdeHandler struct {
	svc *service.SdeService
}

func NewSdeHandler() *SdeHandler {
	return &SdeHandler{svc: service.NewSdeService()}
}

// GetVersion godoc
// GET /api/v1/sde/version
// 获取当前已导入的 SDE 版本信息
func (h *SdeHandler) GetVersion(c *gin.Context) {
	v, err := h.svc.GetCurrentVersion()
	if err != nil {
		response.Fail(c, response.CodeBizError, "查询版本失败: "+err.Error())
		return
	}
	if v == nil {
		response.OK(c, gin.H{"version": nil, "message": "尚未导入任何 SDE 版本"})
		return
	}
	response.OK(c, v)
}

// GetTypes godoc
// POST /api/v1/sde/types
// 批量查询物品信息（含 group + category + market_group 翻译）
type GetTypesRequest struct {
	TypeIDs    []int  `json:"type_ids"`
	Published  *bool  `json:"published"`
	LanguageID string `json:"language_id"`
}

func (h *SdeHandler) GetTypes(c *gin.Context) {
	var req GetTypesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "参数错误: "+err.Error())
		return
	}
	if len(req.TypeIDs) == 0 {
		response.Fail(c, response.CodeParamError, "type_ids 不能为空")
		return
	}
	if req.LanguageID == "" {
		req.LanguageID = "en"
	}

	list, err := h.svc.GetTypes(req.TypeIDs, req.Published, req.LanguageID)
	if err != nil {
		response.Fail(c, response.CodeBizError, "查询失败: "+err.Error())
		return
	}
	response.OK(c, list)
}

// GetNames godoc
// POST /api/v1/sde/names
// 批量查询 id -> name 映射
type GetNamesRequest struct {
	Language string           `json:"language"`
	IDs      map[string][]int `json:"ids"` // key 为 tcID 名称：type/group/category/region/constellation/solar_system/market_group/tech/description
	ESI      []int32          `json:"esi"` // character/corporation/alliance id，调用 ESI /universe/names
}

func (h *SdeHandler) GetNames(c *gin.Context) {
	var req GetNamesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "参数错误: "+err.Error())
		return
	}
	// language 优先 body，其次 Accept-Language header，其次 cookie，最后默认 en
	if req.Language == "" {
		req.Language = c.GetHeader("Accept-Language")
	}
	if req.Language == "" {
		if lang, err := c.Cookie("language"); err == nil && lang != "" {
			req.Language = lang
		}
	}
	if req.Language == "" {
		req.Language = "en"
	}
	if len(req.IDs) == 0 && len(req.ESI) == 0 {
		response.Fail(c, response.CodeParamError, "ids 和 esi 不能同时为空")
		return
	}

	result := make(map[int]string)

	// 查数据库翻译
	if len(req.IDs) > 0 {
		dbNames, err := h.svc.GetNames(req.IDs, req.Language)
		if err != nil {
			response.Fail(c, response.CodeBizError, "查询失败: "+err.Error())
			return
		}
		for k, v := range dbNames {
			result[k] = v
		}
	}

	// 调用 ESI /universe/names 查询 character/corporation/alliance
	if len(req.ESI) > 0 {
		type esiEntry struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}
		var esiResult []esiEntry
		client := esi.NewClient()
		if err := client.PostJSON(
			context.Background(),
			"/universe/names?datasource=tranquility",
			"",
			req.ESI,
			&esiResult,
		); err != nil {
			response.Fail(c, response.CodeBizError, fmt.Sprintf("ESI 查询失败: %v", err))
			return
		}
		for _, e := range esiResult {
			result[e.ID] = e.Name
		}
	}

	response.OK(c, result)
}
