package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// FleetConfigHandler 舰队配置 HTTP 处理器
type FleetConfigHandler struct {
	svc *service.FleetConfigService
}

func NewFleetConfigHandler() *FleetConfigHandler {
	return &FleetConfigHandler{
		svc: service.NewFleetConfigService(),
	}
}

// CreateFleetConfig 创建舰队配置
func (h *FleetConfigHandler) CreateFleetConfig(c *gin.Context) {
	var req service.CreateFleetConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}
	userID := middleware.GetUserID(c)
	result, err := h.svc.CreateFleetConfig(userID, &req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

// ListFleetConfigs 分页查询舰队配置列表
func (h *FleetConfigHandler) ListFleetConfigs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("current", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	records, total, err := h.svc.ListFleetConfigs(page, size)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, gin.H{
		"list":     records,
		"page":     page,
		"pageSize": size,
		"total":    total,
	})
}

// GetFleetConfig 获取舰队配置详情
func (h *FleetConfigHandler) GetFleetConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeParamError, "无效的配置ID")
		return
	}
	result, err := h.svc.GetFleetConfig(uint(id))
	if err != nil {
		response.Fail(c, response.CodeNotFound, err.Error())
		return
	}
	response.OK(c, result)
}

// UpdateFleetConfig 更新舰队配置
func (h *FleetConfigHandler) UpdateFleetConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeParamError, "无效的配置ID")
		return
	}
	var req service.UpdateFleetConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	result, err := h.svc.UpdateFleetConfig(uint(id), userID, userRole, &req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

// DeleteFleetConfig 删除舰队配置
func (h *FleetConfigHandler) DeleteFleetConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeParamError, "无效的配置ID")
		return
	}
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	if err := h.svc.DeleteFleetConfig(uint(id), userID, userRole); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}

// ImportFromUserFitting 从用户装配导入为 EFT 格式
func (h *FleetConfigHandler) ImportFromUserFitting(c *gin.Context) {
	var req service.ImportFittingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}
	userID := middleware.GetUserID(c)
	result, err := h.svc.ImportFromUserFitting(userID, &req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

// ExportToESI 将配置中的装配导出到 ESI
func (h *FleetConfigHandler) ExportToESI(c *gin.Context) {
	var req service.ExportToESIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.svc.ExportToESI(userID, &req); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}

// GetFittingEFT 获取舰队配置中所有装配的本地化 EFT 文本
func (h *FleetConfigHandler) GetFittingEFT(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeParamError, "无效的配置ID")
		return
	}
	lang := c.DefaultQuery("lang", "zh")
	result, err := h.svc.GetFittingEFT(uint(id), lang)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

// GetFittingItems 获取装配物品详情（含重要性、惩罚、替代品）
func (h *FleetConfigHandler) GetFittingItems(c *gin.Context) {
	configID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeParamError, "无效的配置ID")
		return
	}
	fittingID, err := strconv.ParseUint(c.Param("fitting_id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeParamError, "无效的装配ID")
		return
	}
	lang := c.DefaultQuery("lang", "zh")
	result, err := h.svc.GetFittingItems(uint(configID), uint(fittingID), lang)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

// UpdateFittingItemsSettings 批量更新装配物品设置（重要性、惩罚、替代品）
func (h *FleetConfigHandler) UpdateFittingItemsSettings(c *gin.Context) {
	configID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeParamError, "无效的配置ID")
		return
	}
	fittingID, err := strconv.ParseUint(c.Param("fitting_id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeParamError, "无效的装配ID")
		return
	}
	var req service.UpdateFittingItemsSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	if err := h.svc.UpdateFittingItemsSettings(uint(configID), uint(fittingID), userID, userRole, &req); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}
