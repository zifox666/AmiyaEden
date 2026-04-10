package handler

import (
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

const hallOfFameBgMaxSize int64 = 5120 << 10 // 5 MB

// HallOfFameHandler 名人堂处理器
type HallOfFameHandler struct {
	svc *service.HallOfFameService
}

func NewHallOfFameHandler() *HallOfFameHandler {
	return &HallOfFameHandler{svc: service.NewHallOfFameService()}
}

// ─── Public ───

// GetTemple GET /api/v1/hall-of-fame/temple
func (h *HallOfFameHandler) GetTemple(c *gin.Context) {
	data, err := h.svc.GetTemple()
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, data)
}

// ─── Admin: Config ───

// GetConfig GET /api/v1/system/hall-of-fame/config
func (h *HallOfFameHandler) GetConfig(c *gin.Context) {
	cfg, err := h.svc.GetConfig()
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, cfg)
}

// UpdateConfig PUT /api/v1/system/hall-of-fame/config
func (h *HallOfFameHandler) UpdateConfig(c *gin.Context) {
	var req service.HofUpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}
	cfg, err := h.svc.UpdateConfig(&req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, cfg)
}

// UploadBackground POST /api/v1/system/hall-of-fame/upload-background
func (h *HallOfFameHandler) UploadBackground(c *gin.Context) {
	uploadImageAsDataURL(c, hallOfFameBgMaxSize, uploadAllowedMIME)
}

// ─── Admin: Cards ───

// ListCards GET /api/v1/system/hall-of-fame/cards
func (h *HallOfFameHandler) ListCards(c *gin.Context) {
	cards, err := h.svc.ListAllCards()
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, cards)
}

// CreateCard POST /api/v1/system/hall-of-fame/cards
func (h *HallOfFameHandler) CreateCard(c *gin.Context) {
	var req service.CreateCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}
	card, err := h.svc.CreateCard(&req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, card)
}

// UpdateCard PUT /api/v1/system/hall-of-fame/cards/:id
func (h *HallOfFameHandler) UpdateCard(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeParamError, "无效的卡片 ID")
		return
	}
	var req service.UpdateCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}
	card, err := h.svc.UpdateCard(uint(id), &req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, card)
}

// DeleteCard DELETE /api/v1/system/hall-of-fame/cards/:id
func (h *HallOfFameHandler) DeleteCard(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeParamError, "无效的卡片 ID")
		return
	}
	if err := h.svc.DeleteCard(uint(id)); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}

// BatchUpdateLayout PUT /api/v1/system/hall-of-fame/cards/batch-layout
func (h *HallOfFameHandler) BatchUpdateLayout(c *gin.Context) {
	var req []service.CardLayoutUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}
	if err := h.svc.BatchUpdateLayout(req); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}
