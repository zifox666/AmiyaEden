package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CorpStructureHandler struct {
	svc *service.CorpStructureService
}

func NewCorpStructureHandler() *CorpStructureHandler {
	return &CorpStructureHandler{
		svc: service.NewCorpStructureService(),
	}
}

// ListStructures POST /operation/corp-structures/list
func (h *CorpStructureHandler) ListStructures(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req service.CorpStructureListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "参数错误: "+err.Error())
		return
	}

	result, err := h.svc.ListCorpStructures(userID, &req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

// GetCorpIDs GET /operation/corp-structures/corps
func (h *CorpStructureHandler) GetCorpIDs(c *gin.Context) {
	userID := middleware.GetUserID(c)

	corpIDs, err := h.svc.GetUserCorpIDs(userID)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, corpIDs)
}

// GetFuelSetting GET /operation/corp-structures/fuel/settings
func (h *CorpStructureHandler) GetFuelSetting(c *gin.Context) {
	userID := middleware.GetUserID(c)
	corpID, _ := strconv.ParseInt(c.Query("corp_id"), 10, 64)
	if corpID == 0 {
		corpIDs, err := h.svc.GetUserCorpIDs(userID)
		if err != nil {
			response.Fail(c, response.CodeBizError, err.Error())
			return
		}
		if len(corpIDs) == 0 {
			response.Fail(c, response.CodeBizError, "未找到可用军团")
			return
		}
		corpID = corpIDs[0]
	}

	setting, err := h.svc.GetFuelSetting(corpID)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, setting)
}

// UpsertFuelSetting PUT /operation/corp-structures/fuel/settings
func (h *CorpStructureHandler) UpsertFuelSetting(c *gin.Context) {
	operatorID := middleware.GetUserID(c)
	var req service.FuelSettingUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.UpsertFuelSetting(operatorID, &req); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}

// ClaimFuelTask POST /operation/corp-structures/:id/fuel/claim
func (h *CorpStructureHandler) ClaimFuelTask(c *gin.Context) {
	userID := middleware.GetUserID(c)
	structureID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || structureID <= 0 {
		response.Fail(c, response.CodeParamError, "建筑ID无效")
		return
	}
	if err := h.svc.ClaimFuelTask(userID, structureID); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}

// SettleFuelTask POST /operation/corp-structures/:id/fuel/settle
func (h *CorpStructureHandler) SettleFuelTask(c *gin.Context) {
	userID := middleware.GetUserID(c)
	structureID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || structureID <= 0 {
		response.Fail(c, response.CodeParamError, "建筑ID无效")
		return
	}
	result, err := h.svc.SettleFuelTask(userID, structureID)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

type markFuelTaskIskPaidRequest struct {
	Note string `json:"note"`
}

// MarkFuelTaskIskPaid POST /operation/corp-structures/fuel-tasks/:task_id/isk/mark-paid
func (h *CorpStructureHandler) MarkFuelTaskIskPaid(c *gin.Context) {
	operatorID := middleware.GetUserID(c)
	taskID64, err := strconv.ParseUint(c.Param("task_id"), 10, 64)
	if err != nil || taskID64 == 0 {
		response.Fail(c, response.CodeParamError, "任务ID无效")
		return
	}
	var req markFuelTaskIskPaidRequest
	_ = c.ShouldBindJSON(&req)
	if err := h.svc.MarkIskPaid(uint(taskID64), operatorID, req.Note); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}
