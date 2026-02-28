package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/repository"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SysWalletHandler 系统钱包 HTTP 处理器
type SysWalletHandler struct {
	svc *service.SysWalletService
}

func NewSysWalletHandler() *SysWalletHandler {
	return &SysWalletHandler{svc: service.NewSysWalletService()}
}

// ─────────────────────────────────────────────
//  用户端（POST 接口）
// ─────────────────────────────────────────────

// GetMyWallet POST /wallet/my
// 获取当前用户钱包信息
func (h *SysWalletHandler) GetMyWallet(c *gin.Context) {
	userID := middleware.GetUserID(c)
	wallet, err := h.svc.GetMyWallet(userID)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, wallet)
}

// listRequest 通用分页请求
type walletListRequest struct {
	Current int `json:"current"`
	Size    int `json:"size"`
}

// GetMyTransactions POST /wallet/my/transactions
// 获取当前用户钱包流水
func (h *SysWalletHandler) GetMyTransactions(c *gin.Context) {
	var req walletListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Current = 1
		req.Size = 20
	}
	userID := middleware.GetUserID(c)

	records, total, err := h.svc.GetMyTransactions(userID, req.Current, req.Size)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OKWithPage(c, records, total, req.Current, req.Size)
}

// ─────────────────────────────────────────────
//  管理员端（POST 接口）
// ─────────────────────────────────────────────

// AdminListWallets POST /system/wallet/list
// 管理员查询所有用户钱包
func (h *SysWalletHandler) AdminListWallets(c *gin.Context) {
	var req walletListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Current = 1
		req.Size = 20
	}

	wallets, total, err := h.svc.AdminListWallets(req.Current, req.Size)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OKWithPage(c, wallets, total, req.Current, req.Size)
}

// adminGetWalletRequest 查看指定用户钱包请求
type adminGetWalletRequest struct {
	UserID uint `json:"user_id" binding:"required"`
}

// AdminGetWallet POST /system/wallet/detail
// 管理员查看指定用户钱包
func (h *SysWalletHandler) AdminGetWallet(c *gin.Context) {
	var req adminGetWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}

	wallet, err := h.svc.AdminGetWallet(req.UserID)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, wallet)
}

// AdminAdjust POST /system/wallet/adjust
// 管理员调整用户钱包余额
func (h *SysWalletHandler) AdminAdjust(c *gin.Context) {
	var req service.AdminAdjustRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}

	operatorID := middleware.GetUserID(c)
	wallet, err := h.svc.AdminAdjust(operatorID, &req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, wallet)
}

// adminTransactionListRequest 管理员查询流水请求
type adminTransactionListRequest struct {
	Current int    `json:"current"`
	Size    int    `json:"size"`
	UserID  *uint  `json:"user_id"`
	RefType string `json:"ref_type"`
}

// AdminListTransactions POST /system/wallet/transactions
// 管理员查询钱包流水（可按用户/类型筛选）
func (h *SysWalletHandler) AdminListTransactions(c *gin.Context) {
	var req adminTransactionListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Current = 1
		req.Size = 20
	}

	filter := repository.WalletTransactionFilter{
		UserID:  req.UserID,
		RefType: req.RefType,
	}

	records, total, err := h.svc.AdminListTransactions(req.Current, req.Size, filter)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OKWithPage(c, records, total, req.Current, req.Size)
}

// adminLogListRequest 管理员查询操作日志请求
type adminLogListRequest struct {
	Current    int    `json:"current"`
	Size       int    `json:"size"`
	OperatorID *uint  `json:"operator_id"`
	TargetUID  *uint  `json:"target_uid"`
	Action     string `json:"action"`
}

// AdminListLogs POST /system/wallet/logs
// 管理员查询操作日志
func (h *SysWalletHandler) AdminListLogs(c *gin.Context) {
	var req adminLogListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Current = 1
		req.Size = 20
	}

	filter := repository.WalletLogFilter{
		OperatorID: req.OperatorID,
		TargetUID:  req.TargetUID,
		Action:     req.Action,
	}

	records, total, err := h.svc.AdminListLogs(req.Current, req.Size, filter)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OKWithPage(c, records, total, req.Current, req.Size)
}

// ─────────────────────────────────────────────
//  兼容旧接口：供 fleet 路由复用
// ─────────────────────────────────────────────

// GetWallet 兼容旧的 GET /operation/wallet
func (h *SysWalletHandler) GetWallet(c *gin.Context) {
	userID := middleware.GetUserID(c)
	wallet, err := h.svc.GetMyWallet(userID)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, wallet)
}

// GetWalletTransactions 兼容旧的 GET /operation/wallet/transactions
func (h *SysWalletHandler) GetWalletTransactions(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("current", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	records, total, err := h.svc.GetMyTransactions(userID, page, size)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, gin.H{
		"records": records,
		"current": page,
		"size":    size,
		"total":   total,
	})
}
