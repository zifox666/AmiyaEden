package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/model"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
)

// LotteryHandler 抽奖 HTTP 处理器
type LotteryHandler struct {
	svc *service.LotteryService
}

func NewLotteryHandler() *LotteryHandler {
	return &LotteryHandler{svc: service.NewLotteryService()}
}

// ─────────────────────────────────────────────
//  用户端
// ─────────────────────────────────────────────

type lotteryListRequest struct {
	Current int `json:"current"`
	Size    int `json:"size"`
}

// ListActivities POST /shop/lottery/list
func (h *LotteryHandler) ListActivities(c *gin.Context) {
	var req lotteryListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Current = 1
		req.Size = 20
	}
	list, total, err := h.svc.ListActiveActivities(req.Current, req.Size)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OKWithPage(c, list, total, req.Current, req.Size)
}

type drawRequest struct {
	ActivityID uint `json:"activity_id" binding:"required"`
}

// Draw POST /shop/lottery/draw
func (h *LotteryHandler) Draw(c *gin.Context) {
	var req drawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "参数错误: "+err.Error())
		return
	}
	userID := middleware.GetUserID(c)
	result, err := h.svc.Draw(userID, req.ActivityID)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, result)
}

// GetMyRecords POST /shop/lottery/records
func (h *LotteryHandler) GetMyRecords(c *gin.Context) {
	var req lotteryListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Current = 1
		req.Size = 20
	}
	userID := middleware.GetUserID(c)
	list, total, err := h.svc.GetMyLotteryRecords(userID, req.Current, req.Size)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OKWithPage(c, list, total, req.Current, req.Size)
}

// ─────────────────────────────────────────────
//  管理员端
// ─────────────────────────────────────────────

// AdminListActivities POST /system/shop/lottery/list
func (h *LotteryHandler) AdminListActivities(c *gin.Context) {
	var req lotteryListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Current = 1
		req.Size = 20
	}
	list, total, err := h.svc.AdminListActivities(req.Current, req.Size)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OKWithPage(c, list, total, req.Current, req.Size)
}

type adminActivityCreateRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Image       string  `json:"image"`
	CostPerDraw float64 `json:"cost_per_draw"`
	Status      int8    `json:"status"`
	StartAt     *string `json:"start_at"` // ISO 8601 or null
	EndAt       *string `json:"end_at"`
	SortOrder   int     `json:"sort_order"`
}

// AdminCreateActivity POST /system/shop/lottery/add
func (h *LotteryHandler) AdminCreateActivity(c *gin.Context) {
	var req adminActivityCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "参数错误: "+err.Error())
		return
	}
	activity := &model.ShopLotteryActivity{
		Name:        req.Name,
		Description: req.Description,
		Image:       req.Image,
		CostPerDraw: req.CostPerDraw,
		Status:      req.Status,
		SortOrder:   req.SortOrder,
	}
	activity.StartAt = parseTimePtr(req.StartAt)
	activity.EndAt = parseTimePtr(req.EndAt)

	if err := h.svc.AdminCreateActivity(activity); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, activity)
}

type adminActivityEditRequest struct {
	ID          uint     `json:"id" binding:"required"`
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	Image       *string  `json:"image"`
	CostPerDraw *float64 `json:"cost_per_draw"`
	Status      *int8    `json:"status"`
	StartAt     *string  `json:"start_at"`
	EndAt       *string  `json:"end_at"`
	SortOrder   *int     `json:"sort_order"`
}

// AdminUpdateActivity POST /system/shop/lottery/edit
func (h *LotteryHandler) AdminUpdateActivity(c *gin.Context) {
	var req adminActivityEditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "参数错误: "+err.Error())
		return
	}
	updateReq := &service.AdminLotteryActivityUpdateRequest{
		Name:        req.Name,
		Description: req.Description,
		Image:       req.Image,
		CostPerDraw: req.CostPerDraw,
		Status:      req.Status,
		StartAt:     parseTimePtr(req.StartAt),
		EndAt:       parseTimePtr(req.EndAt),
		SortOrder:   req.SortOrder,
	}
	activity, err := h.svc.AdminUpdateActivity(req.ID, updateReq)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, activity)
}

type adminDeleteRequest struct {
	ID uint `json:"id" binding:"required"`
}

// AdminDeleteActivity POST /system/shop/lottery/delete
func (h *LotteryHandler) AdminDeleteActivity(c *gin.Context) {
	var req adminDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.AdminDeleteActivity(req.ID); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}

type adminPrizeCreateRequest struct {
	ActivityID        uint   `json:"activity_id" binding:"required"`
	Name              string `json:"name" binding:"required"`
	Image             string `json:"image"`
	Tier              string `json:"tier"`
	ProbabilityWeight int    `json:"probability_weight"`
	TotalStock        int    `json:"total_stock"`
}

// AdminCreatePrize POST /system/shop/lottery/prize/add
func (h *LotteryHandler) AdminCreatePrize(c *gin.Context) {
	var req adminPrizeCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "参数错误: "+err.Error())
		return
	}
	if req.Tier == "" {
		req.Tier = model.LotteryPrizeTierNormal
	}
	prize := &model.ShopLotteryPrize{
		ActivityID:        req.ActivityID,
		Name:              req.Name,
		Image:             req.Image,
		Tier:              req.Tier,
		ProbabilityWeight: req.ProbabilityWeight,
		TotalStock:        req.TotalStock,
	}
	if err := h.svc.AdminCreatePrize(prize); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, prize)
}

type adminPrizeEditRequest struct {
	ID uint `json:"id" binding:"required"`
	service.AdminLotteryPrizeUpdateRequest
}

// AdminUpdatePrize POST /system/shop/lottery/prize/edit
func (h *LotteryHandler) AdminUpdatePrize(c *gin.Context) {
	var req adminPrizeEditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "参数错误: "+err.Error())
		return
	}
	prize, err := h.svc.AdminUpdatePrize(req.ID, &req.AdminLotteryPrizeUpdateRequest)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, prize)
}

// AdminDeletePrize POST /system/shop/lottery/prize/delete
func (h *LotteryHandler) AdminDeletePrize(c *gin.Context) {
	var req adminDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.AdminDeletePrize(req.ID); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}

type adminLotteryRecordListRequest struct {
	Current    int   `json:"current"`
	Size       int   `json:"size"`
	ActivityID *uint `json:"activity_id"`
}

// AdminListRecords POST /system/shop/lottery/records
func (h *LotteryHandler) AdminListRecords(c *gin.Context) {
	var req adminLotteryRecordListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Current = 1
		req.Size = 20
	}
	list, total, err := h.svc.AdminListRecords(req.Current, req.Size, req.ActivityID)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OKWithPage(c, list, total, req.Current, req.Size)
}

type adminUpdateRecordDeliveryRequest struct {
	ID             uint   `json:"id" binding:"required"`
	DeliveryStatus string `json:"delivery_status" binding:"required"`
}

// AdminUpdateRecordDelivery POST /system/shop/lottery/records/deliver
func (h *LotteryHandler) AdminUpdateRecordDelivery(c *gin.Context) {
	var req adminUpdateRecordDeliveryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.AdminUpdateRecordDelivery(req.ID, req.DeliveryStatus); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}

// parseTimePtr 将 ISO 8601 字符串指针转为 *time.Time
func parseTimePtr(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	formats := []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02 15:04:05", "2006-01-02"}
	for _, f := range formats {
		if t, err := time.ParseInLocation(f, *s, time.Local); err == nil {
			return &t
		}
	}
	return nil
}
