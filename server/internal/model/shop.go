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

// ─── 抽奖活动状态 ───

const (
	LotteryStatusActive   int8 = 1 // 进行中
	LotteryStatusInactive int8 = 0 // 已关闭
)

// ─── 抽奖奖品稀有度 ───

const (
	LotteryPrizeTierNormal    = "normal"    // 普通
	LotteryPrizeTierRare      = "rare"      // 稀有
	LotteryPrizeTierLegendary = "legendary" // 传说
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

// ─── 抽奖奖品发放状态 ───

const (
	LotteryDeliveryPending   = "pending"   // 待发放
	LotteryDeliveryDelivered = "delivered" // 已发放
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

// ─── 抽奖活动 ───

// ShopLotteryActivity 抽奖活动
type ShopLotteryActivity struct {
	BaseModel
	Name        string             `gorm:"size:200;not null"      json:"name"`
	Description string             `gorm:"type:text"              json:"description"`
	Image       string             `gorm:"size:500"               json:"image"`         // 活动封面图
	CostPerDraw float64            `gorm:"not null;default:0"     json:"cost_per_draw"` // 每次抽奖费用
	Status      int8               `gorm:"default:1;index"        json:"status"`        // 1=进行中 0=已关闭
	StartAt     *time.Time         `json:"start_at"`                                    // nil = 无限制
	EndAt       *time.Time         `json:"end_at"`                                      // nil = 无限制
	SortOrder   int                `gorm:"default:0"              json:"sort_order"`
	Prizes      []ShopLotteryPrize `gorm:"foreignKey:ActivityID" json:"prizes,omitempty"`
}

func (ShopLotteryActivity) TableName() string { return "shop_lottery_activity" }

// ShopLotteryPrize 抽奖奖品
type ShopLotteryPrize struct {
	BaseModel
	ActivityID        uint   `gorm:"index;not null"         json:"activity_id"`
	Name              string `gorm:"size:200;not null"      json:"name"`
	Image             string `gorm:"size:500"               json:"image"`              // 奖品图片
	Tier              string `gorm:"size:20;default:'normal'" json:"tier"`             // normal / rare / legendary
	ProbabilityWeight int    `gorm:"not null;default:1"     json:"probability_weight"` // 相对权重
	TotalStock        int    `gorm:"not null;default:0"     json:"total_stock"`        // 库存总量，0=无限
	DrawnCount        int    `gorm:"not null;default:0"     json:"drawn_count"`        // 已抽出数量
}

func (ShopLotteryPrize) TableName() string { return "shop_lottery_prize" }

// ShopLotteryRecord 抽奖记录
type ShopLotteryRecord struct {
	BaseModel
	UserID         uint    `gorm:"index;not null"              json:"user_id"`
	ActivityID     uint    `gorm:"index;not null"              json:"activity_id"`
	ActivityName   string  `gorm:"size:200"                    json:"activity_name"` // 快照
	PrizeID        uint    `gorm:"index;not null"              json:"prize_id"`
	PrizeName      string  `gorm:"size:200"                    json:"prize_name"`  // 快照
	PrizeTier      string  `gorm:"size:20"                     json:"prize_tier"`  // 快照
	PrizeImage     string  `gorm:"size:500"                    json:"prize_image"` // 快照
	Cost           float64 `gorm:"not null"                    json:"cost"`
	DeliveryStatus string  `gorm:"size:20;default:'pending'"   json:"delivery_status"` // pending / delivered
}

func (ShopLotteryRecord) TableName() string { return "shop_lottery_record" }
