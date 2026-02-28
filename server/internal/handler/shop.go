package handler

import (
	"amiya-eden/internal/middleware"
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"amiya-eden/internal/service"
	"amiya-eden/pkg/response"

	"github.com/gin-gonic/gin"
)

// ShopHandler 商店 HTTP 处理器
type ShopHandler struct {
	svc *service.ShopService
}

func NewShopHandler() *ShopHandler {
	return &ShopHandler{svc: service.NewShopService()}
}

// ─────────────────────────────────────────────
//  用户端（全部 POST）
// ─────────────────────────────────────────────

// shopListRequest 通用分页请求
type shopListRequest struct {
	Current int    `json:"current"`
	Size    int    `json:"size"`
	Type    string `json:"type"`
	Status  string `json:"status"`
}

// ListProducts POST /shop/products
// 获取上架商品列表
func (h *ShopHandler) ListProducts(c *gin.Context) {
	var req shopListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Current = 1
		req.Size = 20
	}

	list, total, err := h.svc.ListOnSaleProducts(req.Current, req.Size, req.Type)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OKWithPage(c, list, total, req.Current, req.Size)
}

// productDetailRequest 商品详情请求
type productDetailRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
}

// GetProductDetail POST /shop/product/detail
// 获取商品详情
func (h *ShopHandler) GetProductDetail(c *gin.Context) {
	var req productDetailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}

	product, err := h.svc.GetProductDetail(req.ProductID)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, product)
}

// BuyProduct POST /shop/buy
// 购买商品
func (h *ShopHandler) BuyProduct(c *gin.Context) {
	var req service.BuyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	order, err := h.svc.BuyProduct(userID, &req)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, order)
}

// GetMyOrders POST /shop/orders
// 获取我的订单
func (h *ShopHandler) GetMyOrders(c *gin.Context) {
	var req shopListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Current = 1
		req.Size = 20
	}

	userID := middleware.GetUserID(c)
	list, total, err := h.svc.GetMyOrders(userID, req.Current, req.Size, req.Status)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OKWithPage(c, list, total, req.Current, req.Size)
}

// GetMyRedeemCodes POST /shop/redeem/list
// 获取我的兑换码
func (h *ShopHandler) GetMyRedeemCodes(c *gin.Context) {
	var req shopListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Current = 1
		req.Size = 20
	}

	userID := middleware.GetUserID(c)
	list, total, err := h.svc.GetMyRedeemCodes(userID, req.Current, req.Size)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OKWithPage(c, list, total, req.Current, req.Size)
}

// ─────────────────────────────────────────────
//  管理员端（全部 POST）
// ─────────────────────────────────────────────

// adminProductCreateRequest 创建商品请求
type adminProductCreateRequest struct {
	Name         string  `json:"name" binding:"required"`
	Description  string  `json:"description"`
	Image        string  `json:"image"`
	Price        float64 `json:"price" binding:"required,gt=0"`
	Stock        int     `json:"stock"`        // -1=无限，>=0 有限
	MaxPerUser   int     `json:"max_per_user"` // 0=不限购
	Type         string  `json:"type" binding:"required,oneof=normal redeem"`
	NeedApproval bool    `json:"need_approval"`
	Status       int8    `json:"status"`
	SortOrder    int     `json:"sort_order"`
}

// AdminCreateProduct POST /system/shop/product/add
// 管理员创建商品
func (h *ShopHandler) AdminCreateProduct(c *gin.Context) {
	var req adminProductCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}

	product := &model.ShopProduct{
		Name:         req.Name,
		Description:  req.Description,
		Image:        req.Image,
		Price:        req.Price,
		Stock:        req.Stock,
		MaxPerUser:   req.MaxPerUser,
		Type:         req.Type,
		NeedApproval: req.NeedApproval,
		Status:       req.Status,
		SortOrder:    req.SortOrder,
	}

	if err := h.svc.AdminCreateProduct(product); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, product)
}

// adminProductUpdateRequest 更新商品请求
type adminProductUpdateRequest struct {
	ID uint `json:"id" binding:"required"`
	service.AdminProductUpdateRequest
}

// AdminUpdateProduct POST /system/shop/product/edit
// 管理员更新商品
func (h *ShopHandler) AdminUpdateProduct(c *gin.Context) {
	var req adminProductUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}

	product, err := h.svc.AdminUpdateProduct(req.ID, &req.AdminProductUpdateRequest)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, product)
}

// adminProductDeleteRequest 删除商品请求
type adminProductDeleteRequest struct {
	ID uint `json:"id" binding:"required"`
}

// AdminDeleteProduct POST /system/shop/product/delete
// 管理员删除商品
func (h *ShopHandler) AdminDeleteProduct(c *gin.Context) {
	var req adminProductDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}

	if err := h.svc.AdminDeleteProduct(req.ID); err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, nil)
}

// adminProductListRequest 管理员商品列表请求
type adminProductListRequest struct {
	Current int    `json:"current"`
	Size    int    `json:"size"`
	Status  *int8  `json:"status"`
	Type    string `json:"type"`
	Name    string `json:"name"`
}

// AdminListProducts POST /system/shop/product/list
// 管理员查询商品列表
func (h *ShopHandler) AdminListProducts(c *gin.Context) {
	var req adminProductListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Current = 1
		req.Size = 20
	}

	filter := repository.ProductFilter{
		Status: req.Status,
		Type:   req.Type,
		Name:   req.Name,
	}

	list, total, err := h.svc.AdminListProducts(req.Current, req.Size, filter)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OKWithPage(c, list, total, req.Current, req.Size)
}

// adminOrderListRequest 管理员订单列表请求
type adminOrderListRequest struct {
	Current   int    `json:"current"`
	Size      int    `json:"size"`
	UserID    *uint  `json:"user_id"`
	ProductID *uint  `json:"product_id"`
	Status    string `json:"status"`
}

// AdminListOrders POST /system/shop/order/list
// 管理员查询订单列表
func (h *ShopHandler) AdminListOrders(c *gin.Context) {
	var req adminOrderListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Current = 1
		req.Size = 20
	}

	filter := repository.OrderFilter{
		UserID:    req.UserID,
		ProductID: req.ProductID,
		Status:    req.Status,
	}

	list, total, err := h.svc.AdminListOrders(req.Current, req.Size, filter)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OKWithPage(c, list, total, req.Current, req.Size)
}

// adminOrderReviewRequest 审批/拒绝订单请求
type adminOrderReviewRequest struct {
	OrderID uint   `json:"order_id" binding:"required"`
	Remark  string `json:"remark"`
}

// AdminApproveOrder POST /system/shop/order/approve
// 管理员审批通过订单
func (h *ShopHandler) AdminApproveOrder(c *gin.Context) {
	var req adminOrderReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}

	operatorID := middleware.GetUserID(c)
	order, err := h.svc.AdminApproveOrder(req.OrderID, operatorID, req.Remark)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, order)
}

// AdminRejectOrder POST /system/shop/order/reject
// 管理员拒绝订单
func (h *ShopHandler) AdminRejectOrder(c *gin.Context) {
	var req adminOrderReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeParamError, "请求参数错误: "+err.Error())
		return
	}

	operatorID := middleware.GetUserID(c)
	order, err := h.svc.AdminRejectOrder(req.OrderID, operatorID, req.Remark)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OK(c, order)
}

// adminRedeemListRequest 兑换码列表请求
type adminRedeemListRequest struct {
	Current   int    `json:"current"`
	Size      int    `json:"size"`
	ProductID *uint  `json:"product_id"`
	Status    string `json:"status"`
}

// AdminListRedeemCodes POST /system/shop/redeem/list
// 管理员查询兑换码
func (h *ShopHandler) AdminListRedeemCodes(c *gin.Context) {
	var req adminRedeemListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Current = 1
		req.Size = 20
	}

	list, total, err := h.svc.AdminListRedeemCodes(req.Current, req.Size, req.ProductID, req.Status)
	if err != nil {
		response.Fail(c, response.CodeBizError, err.Error())
		return
	}
	response.OKWithPage(c, list, total, req.Current, req.Size)
}
