package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"

	"gorm.io/gorm"
)

// ShopRepository 商店数据访问层
type ShopRepository struct{}

func NewShopRepository() *ShopRepository {
	return &ShopRepository{}
}

// ─────────────────────────────────────────────
//  商品
// ─────────────────────────────────────────────

// CreateProduct 创建商品
func (r *ShopRepository) CreateProduct(p *model.ShopProduct) error {
	return global.DB.Create(p).Error
}

// UpdateProduct 更新商品
func (r *ShopRepository) UpdateProduct(p *model.ShopProduct) error {
	return global.DB.Save(p).Error
}

// DeleteProduct 删除商品（软删除）
func (r *ShopRepository) DeleteProduct(id uint) error {
	return global.DB.Delete(&model.ShopProduct{}, id).Error
}

// GetProductByID 根据 ID 获取商品
func (r *ShopRepository) GetProductByID(id uint) (*model.ShopProduct, error) {
	var p model.ShopProduct
	if err := global.DB.First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

// ProductFilter 商品查询筛选
type ProductFilter struct {
	Status *int8
	Type   string
	Name   string
}

// ListProducts 分页查询商品
func (r *ShopRepository) ListProducts(page, pageSize int, filter ProductFilter) ([]model.ShopProduct, int64, error) {
	var list []model.ShopProduct
	var total int64
	offset := (page - 1) * pageSize

	db := global.DB.Model(&model.ShopProduct{})
	if filter.Status != nil {
		db = db.Where("status = ?", *filter.Status)
	}
	if filter.Type != "" {
		db = db.Where("type = ?", filter.Type)
	}
	if filter.Name != "" {
		db = db.Where("name ILIKE ?", "%"+filter.Name+"%")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Order("sort_order DESC, id DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// DecrStock 扣减库存（事务中使用，stock > 0 才扣减）
func (r *ShopRepository) DecrStockTx(tx *gorm.DB, productID uint, qty int) error {
	result := tx.Model(&model.ShopProduct{}).
		Where("id = ? AND stock >= ?", productID, qty).
		Update("stock", gorm.Expr("stock - ?", qty))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound // 库存不足
	}
	return nil
}

// ─────────────────────────────────────────────
//  订单
// ─────────────────────────────────────────────

// CreateOrder 创建订单
func (r *ShopRepository) CreateOrder(o *model.ShopOrder) error {
	return global.DB.Create(o).Error
}

// CreateOrderTx 在事务中创建订单
func (r *ShopRepository) CreateOrderTx(tx *gorm.DB, o *model.ShopOrder) error {
	return tx.Create(o).Error
}

// UpdateOrder 更新订单
func (r *ShopRepository) UpdateOrder(o *model.ShopOrder) error {
	return global.DB.Save(o).Error
}

// UpdateOrderTx 在事务中更新订单
func (r *ShopRepository) UpdateOrderTx(tx *gorm.DB, o *model.ShopOrder) error {
	return tx.Save(o).Error
}

// GetOrderByID 根据 ID 获取订单
func (r *ShopRepository) GetOrderByID(id uint) (*model.ShopOrder, error) {
	var o model.ShopOrder
	if err := global.DB.First(&o, id).Error; err != nil {
		return nil, err
	}
	return &o, nil
}

// GetOrderByOrderNo 根据订单号获取订单
func (r *ShopRepository) GetOrderByOrderNo(orderNo string) (*model.ShopOrder, error) {
	var o model.ShopOrder
	if err := global.DB.Where("order_no = ?", orderNo).First(&o).Error; err != nil {
		return nil, err
	}
	return &o, nil
}

// OrderFilter 订单查询筛选
type OrderFilter struct {
	UserID    *uint
	ProductID *uint
	Status    string
}

// ListOrders 分页查询订单
func (r *ShopRepository) ListOrders(page, pageSize int, filter OrderFilter) ([]model.ShopOrder, int64, error) {
	var list []model.ShopOrder
	var total int64
	offset := (page - 1) * pageSize

	db := global.DB.Model(&model.ShopOrder{})
	if filter.UserID != nil {
		db = db.Where("user_id = ?", *filter.UserID)
	}
	if filter.ProductID != nil {
		db = db.Where("product_id = ?", *filter.ProductID)
	}
	if filter.Status != "" {
		db = db.Where("status = ?", filter.Status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// CountUserProductPurchased 统计用户对某商品的已购数量（pending + paid + approved + completed）
func (r *ShopRepository) CountUserProductPurchased(userID, productID uint) (int64, error) {
	var total int64
	err := global.DB.Model(&model.ShopOrder{}).
		Where("user_id = ? AND product_id = ? AND status IN ?", userID, productID,
			[]string{model.OrderStatusPending, model.OrderStatusPaid, model.OrderStatusApproved, model.OrderStatusCompleted}).
		Select("COALESCE(SUM(quantity), 0)").Scan(&total).Error
	return total, err
}

// ─────────────────────────────────────────────
//  兑换码
// ─────────────────────────────────────────────

// CreateRedeemCode 创建兑换码
func (r *ShopRepository) CreateRedeemCode(rc *model.ShopRedeemCode) error {
	return global.DB.Create(rc).Error
}

// CreateRedeemCodeTx 在事务中创建兑换码
func (r *ShopRepository) CreateRedeemCodeTx(tx *gorm.DB, rc *model.ShopRedeemCode) error {
	return tx.Create(rc).Error
}

// ListRedeemCodesByOrder 根据订单查询兑换码
func (r *ShopRepository) ListRedeemCodesByOrder(orderID uint) ([]model.ShopRedeemCode, error) {
	var list []model.ShopRedeemCode
	err := global.DB.Where("order_id = ?", orderID).Find(&list).Error
	return list, err
}

// ListRedeemCodesByUser 根据用户查询兑换码
func (r *ShopRepository) ListRedeemCodesByUser(userID uint, page, pageSize int) ([]model.ShopRedeemCode, int64, error) {
	var list []model.ShopRedeemCode
	var total int64
	offset := (page - 1) * pageSize

	db := global.DB.Model(&model.ShopRedeemCode{}).Where("user_id = ?", userID)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// GetRedeemCodeByCode 根据兑换码查询
func (r *ShopRepository) GetRedeemCodeByCode(code string) (*model.ShopRedeemCode, error) {
	var rc model.ShopRedeemCode
	if err := global.DB.Where("code = ?", code).First(&rc).Error; err != nil {
		return nil, err
	}
	return &rc, nil
}

// UpdateRedeemCode 更新兑换码
func (r *ShopRepository) UpdateRedeemCode(rc *model.ShopRedeemCode) error {
	return global.DB.Save(rc).Error
}

// AdminListRedeemCodes 管理员分页查询兑换码
func (r *ShopRepository) AdminListRedeemCodes(page, pageSize int, productID *uint, status string) ([]model.ShopRedeemCode, int64, error) {
	var list []model.ShopRedeemCode
	var total int64
	offset := (page - 1) * pageSize

	db := global.DB.Model(&model.ShopRedeemCode{})
	if productID != nil {
		db = db.Where("product_id = ?", *productID)
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
