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
	repo        *repository.ShopRepository
	walletSvc   *SysWalletService
	userRepo    *repository.UserRepository
	charRepo    *repository.EveCharacterRepository
}

func NewShopService() *ShopService {
	return &ShopService{
		repo:      repository.NewShopRepository(),
		walletSvc: NewSysWalletService(),
		userRepo:  repository.NewUserRepository(),
		charRepo:  repository.NewEveCharacterRepository(),
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

// BuyProduct 购买商品：立即扣款，状态设为 requested
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
		limitPeriod := product.LimitPeriod
		if limitPeriod == "" {
			limitPeriod = model.LimitPeriodForever
		}
		purchased, err := s.repo.CountUserProductPurchased(userID, product.ID, limitPeriod)
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

	// 4. 检查余额
	wallet, err := s.walletSvc.GetMyWallet(userID)
	if err != nil {
		return nil, fmt.Errorf("获取钱包失败: %w", err)
	}
	if wallet.Balance < totalPrice {
		return nil, errors.New("余额不足")
	}

	// 5. 获取用户信息快照
	mainCharName, nickname, qq, discordID := s.getUserSnapshot(userID)

	// 6. 生成订单号
	orderNo := generateOrderNo()

	// 7. 开启事务（库存扣减 + 创建订单）
	tx := global.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 8. 扣减库存（有限库存才扣）
	if product.Stock >= 0 {
		if err := s.repo.DecrStockTx(tx, product.ID, req.Quantity); err != nil {
			tx.Rollback()
			return nil, errors.New("库存不足")
		}
	}

	// 9. 创建订单（状态 requested）
	order := &model.ShopOrder{
		OrderNo:           orderNo,
		UserID:            userID,
		MainCharacterName: mainCharName,
		Nickname:          nickname,
		QQ:                qq,
		DiscordID:         discordID,
		ProductID:         product.ID,
		ProductName:       product.Name,
		ProductType:       product.Type,
		Quantity:          req.Quantity,
		UnitPrice:         product.Price,
		TotalPrice:        totalPrice,
		Remark:            req.Remark,
		Status:            model.OrderStatusRequested,
	}

	if err := s.repo.CreateOrderTx(tx, order); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建订单失败: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	// 10. 立即扣款
	reason := fmt.Sprintf("购买商品: %s x%d", product.Name, req.Quantity)
	refID := fmt.Sprintf("order:%s", order.OrderNo)
	if err := s.walletSvc.DebitUser(userID, totalPrice, reason, model.WalletRefShopBuy, refID); err != nil {
		// 扣款失败，拒绝订单并恢复库存
		s.rollbackOrder(order, product)
		return nil, fmt.Errorf("扣款失败: %w", err)
	}

	return order, nil
}

// rollbackOrder 扣款失败时恢复库存并拒绝订单（不退款，因为扣款本就未成功）
func (s *ShopService) rollbackOrder(order *model.ShopOrder, product *model.ShopProduct) {
	if product != nil && product.Stock >= 0 {
		product.Stock += order.Quantity
		_ = s.repo.UpdateProduct(product)
	}
	order.Status = model.OrderStatusRejected
	_ = s.repo.UpdateOrder(order)
}

// getUserSnapshot 获取用户信息快照（主角色名、昵称、QQ、Discord）
func (s *ShopService) getUserSnapshot(userID uint) (mainCharName, nickname, qq, discordID string) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return
	}
	nickname = user.Nickname
	qq = user.QQ
	discordID = user.DiscordID
	if user.PrimaryCharacterID != 0 {
		char, err := s.charRepo.GetByCharacterID(user.PrimaryCharacterID)
		if err == nil {
			mainCharName = char.CharacterName
		}
	}
	return
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
	if req.LimitPeriod != nil {
		product.LimitPeriod = *req.LimitPeriod
	}
	if req.Type != "" {
		product.Type = req.Type
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
	Name        string   `json:"name"`
	Description *string  `json:"description"`
	Image       *string  `json:"image"`
	Price       *float64 `json:"price"`
	Stock       *int     `json:"stock"`
	MaxPerUser  *int     `json:"max_per_user"`
	LimitPeriod *string  `json:"limit_period"`
	Type        string   `json:"type"`
	Status      *int8    `json:"status"`
	SortOrder   *int     `json:"sort_order"`
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

// AdminDeliverOrder 发放订单
func (s *ShopService) AdminDeliverOrder(orderID uint, operatorID uint, remark string) (*model.ShopOrder, error) {
	order, err := s.repo.GetOrderByID(orderID)
	if err != nil {
		return nil, errors.New("订单不存在")
	}
	if order.Status != model.OrderStatusRequested {
		return nil, fmt.Errorf("订单状态为 %s，无法发放", order.Status)
	}

	now := time.Now()
	order.Status = model.OrderStatusDelivered
	order.ReviewedBy = &operatorID
	order.ReviewedAt = &now
	order.ReviewRemark = remark

	// 兑换码类商品 — 生成兑换码
	product, err := s.repo.GetProductByID(order.ProductID)
	if err == nil && product.Type == model.ProductTypeRedeem {
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

	if err := s.repo.UpdateOrder(order); err != nil {
		return nil, fmt.Errorf("更新订单失败: %w", err)
	}

	return order, nil
}

// AdminRejectOrder 拒绝订单（退款）
func (s *ShopService) AdminRejectOrder(orderID uint, operatorID uint, remark string) (*model.ShopOrder, error) {
	order, err := s.repo.GetOrderByID(orderID)
	if err != nil {
		return nil, errors.New("订单不存在")
	}
	if order.Status != model.OrderStatusRequested {
		return nil, fmt.Errorf("订单状态为 %s，无法拒绝", order.Status)
	}

	// 恢复库存
	product, err := s.repo.GetProductByID(order.ProductID)
	if err == nil && product.Stock >= 0 {
		product.Stock += order.Quantity
		_ = s.repo.UpdateProduct(product)
	}

	// 退款
	reason := fmt.Sprintf("商品订单退款: %s x%d", order.ProductName, order.Quantity)
	refID := fmt.Sprintf("order:%s", order.OrderNo)
	if err := s.walletSvc.CreditUser(order.UserID, order.TotalPrice, reason, model.WalletRefShopRefund, refID); err != nil {
		return nil, fmt.Errorf("退款失败: %w", err)
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

// generateOrderNo 生成短订单号: 8位随机大写字母+数字
func generateOrderNo() string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	code := make([]byte, 8)
	for i := range code {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		code[i] = charset[n.Int64()]
	}
	return string(code)
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
