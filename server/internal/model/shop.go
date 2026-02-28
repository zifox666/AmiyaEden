package model

import "time"

// ─────────────────────────────────────────────
//  商店系统
// ─────────────────────────────────────────────

// ─── 商品类型 ───

const (
	ProductTypeNormal = "normal" // 普通商品
	ProductTypeRedeem = "redeem" // 兑换码/服务类商品
)

// ─── 商品状态 ───

const (
	ProductStatusOnSale  int8 = 1 // 上架
	ProductStatusOffSale int8 = 0 // 下架
)

// ─── 订单状态 ───

const (
	OrderStatusPending          = "pending"            // 待审批
	OrderStatusPaid             = "paid"               // 已付款（即时购买）
	OrderStatusApproved         = "approved"           // 已审批（审批流程）
	OrderStatusRejected         = "rejected"           // 已拒绝
	OrderStatusCompleted        = "completed"          // 已完成
	OrderStatusCancelled        = "cancelled"          // 已取消
	OrderStatusInsufficientFund = "insufficient_funds" // 余额不足（审批时）
)

// ─── 兑换码状态 ───

const (
	RedeemStatusUnused  = "unused"
	RedeemStatusUsed    = "used"
	RedeemStatusExpired = "expired"
)

// ─── 数据模型 ───

// ShopProduct 商品
type ShopProduct struct {
	BaseModel
	Name         string  `gorm:"size:200;not null"              json:"name"`
	Description  string  `gorm:"type:text"                      json:"description"`
	Image        string  `gorm:"size:500"                       json:"image"`
	Price        float64 `gorm:"not null"                       json:"price"`
	Stock        int     `gorm:"default:-1"                     json:"stock"`         // -1 = 无限库存
	MaxPerUser   int     `gorm:"default:0"                      json:"max_per_user"`  // 0 = 不限购
	Type         string  `gorm:"size:20;default:'normal';index" json:"type"`          // normal / redeem
	NeedApproval bool    `gorm:"default:false"                  json:"need_approval"` // 是否需要管理员审批
	Status       int8    `gorm:"default:1;index"                json:"status"`        // 1=上架 0=下架
	SortOrder    int     `gorm:"default:0"                      json:"sort_order"`    // 排序（越大越靠前）
}

func (ShopProduct) TableName() string { return "shop_product" }

// ShopOrder 订单
type ShopOrder struct {
	BaseModel
	OrderNo       string     `gorm:"size:50;uniqueIndex"          json:"order_no"`
	UserID        uint       `gorm:"index;not null"               json:"user_id"`
	ProductID     uint       `gorm:"index;not null"               json:"product_id"`
	ProductName   string     `gorm:"size:200"                     json:"product_name"` // 商品名快照
	ProductType   string     `gorm:"size:20"                      json:"product_type"` // 商品类型快照
	Quantity      int        `gorm:"default:1"                    json:"quantity"`
	UnitPrice     float64    `gorm:"not null"                     json:"unit_price"` // 单价快照
	TotalPrice    float64    `gorm:"not null"                     json:"total_price"`
	Status        string     `gorm:"size:30;index;default:'pending'" json:"status"`
	TransactionID *uint      `gorm:"index"                        json:"transaction_id"` // 关联钱包流水 ID
	Remark        string     `gorm:"size:500"                     json:"remark"`         // 用户备注
	ReviewedBy    *uint      `gorm:"index"                        json:"reviewed_by"`    // 审批人
	ReviewedAt    *time.Time `json:"reviewed_at"`
	ReviewRemark  string     `gorm:"size:500"                     json:"review_remark"` // 审批备注
}

func (ShopOrder) TableName() string { return "shop_order" }

// ShopRedeemCode 兑换码
type ShopRedeemCode struct {
	BaseModel
	OrderID   uint       `gorm:"index;not null"              json:"order_id"`
	ProductID uint       `gorm:"index"                       json:"product_id"`
	UserID    uint       `gorm:"index"                       json:"user_id"`
	Code      string     `gorm:"size:50;uniqueIndex"         json:"code"`
	Status    string     `gorm:"size:20;default:'unused'"    json:"status"` // unused / used / expired
	UsedAt    *time.Time `json:"used_at"`
	ExpiresAt *time.Time `json:"expires_at"`
}

func (ShopRedeemCode) TableName() string { return "shop_redeem_code" }
