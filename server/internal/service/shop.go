package service

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"
)

// ShopService 商店业务逻辑层
type ShopService struct {
	repo      *repository.ShopRepository
	walletSvc *SysWalletService
}

func NewShopService() *ShopService {
	return &ShopService{
		repo:      repository.NewShopRepository(),
		walletSvc: NewSysWalletService(),
	}
}

// ─────────────────────────────────────────────
//  用户端
// ─────────────────────────────────────────────

// ListOnSaleProducts 获取上架商品列表
func (s *ShopService) ListOnSaleProducts(page, pageSize int, productType string) ([]model.ShopProduct, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	status := model.ProductStatusOnSale
	filter := repository.ProductFilter{Status: &status, Type: productType}
	return s.repo.ListProducts(page, pageSize, filter)
}

// GetProductDetail 获取商品详情
func (s *ShopService) GetProductDetail(productID uint) (*model.ShopProduct, error) {
	p, err := s.repo.GetProductByID(productID)
	if err != nil {
		return nil, errors.New("商品不存在")
	}
	if p.Status != model.ProductStatusOnSale {
		return nil, errors.New("商品已下架")
	}
	return p, nil
}

// BuyRequest 购买请求
type BuyRequest struct {
	ProductID uint   `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
	Remark    string `json:"remark"`
}

// BuyProduct 购买商品
func (s *ShopService) BuyProduct(userID uint, req *BuyRequest) (*model.ShopOrder, error) {
	// 1. 获取商品
	product, err := s.repo.GetProductByID(req.ProductID)
	if err != nil {
		return nil, errors.New("商品不存在")
	}
	if product.Status != model.ProductStatusOnSale {
		return nil, errors.New("商品已下架")
	}

	// 2. 检查库存
	if product.Stock >= 0 && product.Stock < req.Quantity {
		return nil, errors.New("库存不足")
	}

	// 3. 检查限购
	if product.MaxPerUser > 0 {
		purchased, err := s.repo.CountUserProductPurchased(userID, product.ID)
		if err != nil {
			return nil, fmt.Errorf("查询购买记录失败: %w", err)
		}
		if int(purchased)+req.Quantity > product.MaxPerUser {
			remaining := product.MaxPerUser - int(purchased)
			if remaining <= 0 {
				return nil, errors.New("已达到限购数量")
			}
			return nil, fmt.Errorf("超出限购数量，还可购买 %d 件", remaining)
		}
	}

	totalPrice := product.Price * float64(req.Quantity)

	// 4. 检查余额（即使需要审批也先检查余额，给用户即时反馈）
	wallet, err := s.walletSvc.GetMyWallet(userID)
	if err != nil {
		return nil, fmt.Errorf("获取钱包失败: %w", err)
	}
	if wallet.Balance < totalPrice {
		return nil, errors.New("余额不足")
	}

	// 5. 生成订单号
	orderNo := generateOrderNo()

	// 6. 开启事务（库存扣减 + 创建订单）
	tx := global.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 7. 扣减库存（有限库存才扣）
	if product.Stock >= 0 {
		if err := s.repo.DecrStockTx(tx, product.ID, req.Quantity); err != nil {
			tx.Rollback()
			return nil, errors.New("库存不足")
		}
	}

	// 8. 创建订单
	order := &model.ShopOrder{
		OrderNo:     orderNo,
		UserID:      userID,
		ProductID:   product.ID,
		ProductName: product.Name,
		ProductType: product.Type,
		Quantity:    req.Quantity,
		UnitPrice:   product.Price,
		TotalPrice:  totalPrice,
		Remark:      req.Remark,
	}

	if product.NeedApproval {
		// 需要审批：订单状态为 pending，不扣款
		order.Status = model.OrderStatusPending
	} else {
		// 即时购买：先标记为 paid
		order.Status = model.OrderStatusPaid
	}

	if err := s.repo.CreateOrderTx(tx, order); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建订单失败: %w", err)
	}

	// 提交事务（库存 + 订单已落库）
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	// 9. 即时购买 — 扣款并完成（在事务外执行，避免死锁）
	if !product.NeedApproval {
		if err := s.debitAndComplete(order, userID, product); err != nil {
			// 扣款失败，将订单标记为余额不足
			order.Status = model.OrderStatusInsufficientFund
			_ = s.repo.UpdateOrder(order)
			return nil, err
		}
	}

	return order, nil
}

// debitAndComplete 扣款 + 生成兑换码（如需）+ 标记完成
func (s *ShopService) debitAndComplete(order *model.ShopOrder, userID uint, product *model.ShopProduct) error {
	// 使用 walletSvc.DebitUser 扣款
	reason := fmt.Sprintf("购买商品: %s x%d", product.Name, order.Quantity)
	refID := fmt.Sprintf("order:%s", order.OrderNo)
	if err := s.walletSvc.DebitUser(userID, order.TotalPrice, reason, model.WalletRefShopBuy, refID); err != nil {
		return fmt.Errorf("扣款失败: %w", err)
	}

	// 兑换码类商品 — 生成兑换码
	if product.Type == model.ProductTypeRedeem {
		for i := 0; i < order.Quantity; i++ {
			code := &model.ShopRedeemCode{
				OrderID:   order.ID,
				ProductID: product.ID,
				UserID:    userID,
				Code:      generateRedeemCode(),
				Status:    model.RedeemStatusUnused,
			}
			if err := s.repo.CreateRedeemCode(code); err != nil {
				return fmt.Errorf("生成兑换码失败: %w", err)
			}
		}
	}

	order.Status = model.OrderStatusCompleted
	if err := s.repo.UpdateOrder(order); err != nil {
		return fmt.Errorf("更新订单状态失败: %w", err)
	}

	return nil
}

// GetMyOrders 获取我的订单
func (s *ShopService) GetMyOrders(userID uint, page, pageSize int, status string) ([]model.ShopOrder, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	filter := repository.OrderFilter{UserID: &userID, Status: status}
	return s.repo.ListOrders(page, pageSize, filter)
}

// GetMyRedeemCodes 获取我的兑换码
func (s *ShopService) GetMyRedeemCodes(userID uint, page, pageSize int) ([]model.ShopRedeemCode, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListRedeemCodesByUser(userID, page, pageSize)
}

// ─────────────────────────────────────────────
//  管理员端
// ─────────────────────────────────────────────

// AdminCreateProduct 创建商品
func (s *ShopService) AdminCreateProduct(req *model.ShopProduct) error {
	return s.repo.CreateProduct(req)
}

// AdminUpdateProduct 更新商品
func (s *ShopService) AdminUpdateProduct(id uint, req *AdminProductUpdateRequest) (*model.ShopProduct, error) {
	product, err := s.repo.GetProductByID(id)
	if err != nil {
		return nil, errors.New("商品不存在")
	}

	if req.Name != "" {
		product.Name = req.Name
	}
	if req.Description != nil {
		product.Description = *req.Description
	}
	if req.Image != nil {
		product.Image = *req.Image
	}
	if req.Price != nil {
		product.Price = *req.Price
	}
	if req.Stock != nil {
		product.Stock = *req.Stock
	}
	if req.MaxPerUser != nil {
		product.MaxPerUser = *req.MaxPerUser
	}
	if req.Type != "" {
		product.Type = req.Type
	}
	if req.NeedApproval != nil {
		product.NeedApproval = *req.NeedApproval
	}
	if req.Status != nil {
		product.Status = *req.Status
	}
	if req.SortOrder != nil {
		product.SortOrder = *req.SortOrder
	}

	if err := s.repo.UpdateProduct(product); err != nil {
		return nil, err
	}
	return product, nil
}

// AdminProductUpdateRequest 商品更新请求
type AdminProductUpdateRequest struct {
	Name         string   `json:"name"`
	Description  *string  `json:"description"`
	Image        *string  `json:"image"`
	Price        *float64 `json:"price"`
	Stock        *int     `json:"stock"`
	MaxPerUser   *int     `json:"max_per_user"`
	Type         string   `json:"type"`
	NeedApproval *bool    `json:"need_approval"`
	Status       *int8    `json:"status"`
	SortOrder    *int     `json:"sort_order"`
}

// AdminDeleteProduct 删除商品
func (s *ShopService) AdminDeleteProduct(id uint) error {
	return s.repo.DeleteProduct(id)
}

// AdminListProducts 管理员查询商品（包含下架）
func (s *ShopService) AdminListProducts(page, pageSize int, filter repository.ProductFilter) ([]model.ShopProduct, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListProducts(page, pageSize, filter)
}

// AdminListOrders 管理员查询订单
func (s *ShopService) AdminListOrders(page, pageSize int, filter repository.OrderFilter) ([]model.ShopOrder, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListOrders(page, pageSize, filter)
}

// AdminApproveOrder 审批通过订单
func (s *ShopService) AdminApproveOrder(orderID uint, operatorID uint, remark string) (*model.ShopOrder, error) {
	order, err := s.repo.GetOrderByID(orderID)
	if err != nil {
		return nil, errors.New("订单不存在")
	}
	if order.Status != model.OrderStatusPending {
		return nil, fmt.Errorf("订单状态为 %s，无法审批", order.Status)
	}

	// 获取商品信息
	product, err := s.repo.GetProductByID(order.ProductID)
	if err != nil {
		return nil, errors.New("关联商品不存在")
	}

	// 检查用户余额
	wallet, err := s.walletSvc.GetMyWallet(order.UserID)
	if err != nil {
		return nil, fmt.Errorf("获取用户钱包失败: %w", err)
	}
	if wallet.Balance < order.TotalPrice {
		order.Status = model.OrderStatusInsufficientFund
		now := time.Now()
		order.ReviewedBy = &operatorID
		order.ReviewedAt = &now
		order.ReviewRemark = "审批时用户余额不足"
		_ = s.repo.UpdateOrder(order)
		return nil, errors.New("用户余额不足")
	}

	// 扣款
	reason := fmt.Sprintf("购买商品: %s x%d（审批通过）", order.ProductName, order.Quantity)
	refID := fmt.Sprintf("order:%s", order.OrderNo)
	if err := s.walletSvc.DebitUser(order.UserID, order.TotalPrice, reason, model.WalletRefShopBuy, refID); err != nil {
		return nil, fmt.Errorf("扣款失败: %w", err)
	}

	// 兑换码类商品 — 生成兑换码
	if product.Type == model.ProductTypeRedeem {
		for i := 0; i < order.Quantity; i++ {
			code := &model.ShopRedeemCode{
				OrderID:   order.ID,
				ProductID: product.ID,
				UserID:    order.UserID,
				Code:      generateRedeemCode(),
				Status:    model.RedeemStatusUnused,
			}
			if err := s.repo.CreateRedeemCode(code); err != nil {
				return nil, fmt.Errorf("生成兑换码失败: %w", err)
			}
		}
	}

	now := time.Now()
	order.Status = model.OrderStatusCompleted
	order.ReviewedBy = &operatorID
	order.ReviewedAt = &now
	order.ReviewRemark = remark
	if err := s.repo.UpdateOrder(order); err != nil {
		return nil, fmt.Errorf("更新订单失败: %w", err)
	}

	return order, nil
}

// AdminRejectOrder 拒绝订单
func (s *ShopService) AdminRejectOrder(orderID uint, operatorID uint, remark string) (*model.ShopOrder, error) {
	order, err := s.repo.GetOrderByID(orderID)
	if err != nil {
		return nil, errors.New("订单不存在")
	}
	if order.Status != model.OrderStatusPending {
		return nil, fmt.Errorf("订单状态为 %s，无法拒绝", order.Status)
	}

	// 恢复库存
	product, err := s.repo.GetProductByID(order.ProductID)
	if err == nil && product.Stock >= 0 {
		product.Stock += order.Quantity
		_ = s.repo.UpdateProduct(product)
	}

	now := time.Now()
	order.Status = model.OrderStatusRejected
	order.ReviewedBy = &operatorID
	order.ReviewedAt = &now
	order.ReviewRemark = remark
	if err := s.repo.UpdateOrder(order); err != nil {
		return nil, fmt.Errorf("更新订单失败: %w", err)
	}

	return order, nil
}

// AdminListRedeemCodes 管理员查询兑换码
func (s *ShopService) AdminListRedeemCodes(page, pageSize int, productID *uint, status string) ([]model.ShopRedeemCode, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.AdminListRedeemCodes(page, pageSize, productID, status)
}

// ─────────────────────────────────────────────
//  工具函数
// ─────────────────────────────────────────────

// generateOrderNo 生成订单号: SH + 时间戳 + 4位随机数
func generateOrderNo() string {
	ts := time.Now().Format("20060102150405")
	n, _ := rand.Int(rand.Reader, big.NewInt(10000))
	return fmt.Sprintf("SH%s%04d", ts, n.Int64())
}

// generateRedeemCode 生成兑换码: 16位大写字母+数字
func generateRedeemCode() string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // 去掉易混淆字符
	code := make([]byte, 16)
	for i := range code {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		code[i] = charset[n.Int64()]
	}
	return string(code)
}
