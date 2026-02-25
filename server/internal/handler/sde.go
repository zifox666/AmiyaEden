package handler

import (
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"
	"strconv"

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

// TriggerUpdate godoc
// POST /api/v1/sde/update
// 手动触发 SDE 更新（Admin 或以上权限）
func (h *SdeHandler) TriggerUpdate(c *gin.Context) {
	version, err := h.svc.TriggerManualUpdate()
	if err != nil {
		response.Fail(c, response.CodeBizError, "SDE 更新失败: "+err.Error())
		return
	}
	response.OK(c, gin.H{"version": version, "message": "SDE 导入成功"})
}

// GetTranslation godoc
// GET /api/v1/sde/translation?tc_id=8&key_id=34&language_id=zh
// 查询单条翻译
func (h *SdeHandler) GetTranslation(c *gin.Context) {
	tcID, err1 := strconv.Atoi(c.Query("tc_id"))
	keyID, err2 := strconv.Atoi(c.Query("key_id"))
	languageID := c.Query("language_id")

	if err1 != nil || err2 != nil || languageID == "" {
		response.Fail(c, response.CodeParamError, "参数错误：tc_id, key_id, language_id 均为必填")
		return
	}

	t, err := h.svc.GetTranslation(tcID, keyID, languageID)
	if err != nil {
		response.Fail(c, response.CodeNotFound, "翻译记录不存在")
		return
	}
	response.OK(c, t)
}

// GetTranslationsByKey godoc
// GET /api/v1/sde/translations?tc_id=8&key_id=34
// 查询一个 key 的所有语言翻译
func (h *SdeHandler) GetTranslationsByKey(c *gin.Context) {
	tcID, err1 := strconv.Atoi(c.Query("tc_id"))
	keyID, err2 := strconv.Atoi(c.Query("key_id"))

	if err1 != nil || err2 != nil {
		response.Fail(c, response.CodeParamError, "参数错误：tc_id 和 key_id 均为必填整数")
		return
	}

	list, err := h.svc.GetTranslationsByKey(tcID, keyID)
	if err != nil {
		response.Fail(c, response.CodeBizError, "查询失败: "+err.Error())
		return
	}
	response.OK(c, list)
}

// SearchByName godoc
// GET /api/v1/sde/search?keyword=Tritanium&limit=20
// 名称模糊搜索
func (h *SdeHandler) SearchByName(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		response.Fail(c, response.CodeParamError, "keyword 不能为空")
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	list, err := h.svc.FuzzySearchByName(keyword, limit)
	if err != nil {
		response.Fail(c, response.CodeBizError, "搜索失败: "+err.Error())
		return
	}
	response.OK(c, list)
}

// GetTypeDetail godoc
// GET /api/v1/sde/type/:type_id
// 查询 typeID 详情（含 group + category）
func (h *SdeHandler) GetTypeDetail(c *gin.Context) {
	typeID, err := strconv.Atoi(c.Param("type_id"))
	if err != nil {
		response.Fail(c, response.CodeParamError, "无效的 type_id")
		return
	}

	detail, err := h.svc.GetTypeDetail(typeID)
	if err != nil {
		response.Fail(c, response.CodeNotFound, "type_id 不存在")
		return
	}
	response.OK(c, detail)
}
