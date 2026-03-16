package model

import "time"

// ─────────────────────────────────────────────
//  系统钱包（与 EVE Wallet 无关，独立系统）
// ─────────────────────────────────────────────

// SystemWallet 用户系统钱包（用于发放/兑换奖励）
type SystemWallet struct {
	ID        uint      `gorm:"primarykey"                 json:"id"`
	UserID    uint      `gorm:"uniqueIndex;not null"       json:"user_id"`
	Balance   float64   `gorm:"default:0"                  json:"balance"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"             json:"updated_at"`
}

func (SystemWallet) TableName() string { return "system_wallet" }

// WalletWithCharacter 钱包信息 + 用户主角色名（用于管理列表展示）
type WalletWithCharacter struct {
	SystemWallet
	CharacterName string `json:"character_name"`
}

// TransactionWithCharacter 钱包流水 + 用户主角色名
type TransactionWithCharacter struct {
	WalletTransaction
	CharacterName string `json:"character_name"`
}

// LogWithCharacter 钱包操作日志 + 操作人/目标用户角色名
type LogWithCharacter struct {
	WalletLog
	TargetCharacterName   string `json:"target_character_name"`
	OperatorCharacterName string `json:"operator_character_name"`
}

// WalletTransaction 钱包流水
type WalletTransaction struct {
	ID           uint      `gorm:"primarykey"                 json:"id"`
	UserID       uint      `gorm:"not null;index"             json:"user_id"`
	Amount       float64   `gorm:"not null"                   json:"amount"` // 正数=收入 负数=支出
	Reason       string    `gorm:"size:256"                   json:"reason"`
	RefType      string    `gorm:"size:64;index"              json:"ref_type"` // pap_reward / manual / redeem / admin_adjust
	RefID        string    `gorm:"size:64"                    json:"ref_id"`   // 关联 ID（如 fleet_id）
	BalanceAfter float64   `gorm:"not null"                   json:"balance_after"`
	OperatorID   uint      `gorm:"default:0;index"            json:"operator_id"` // 操作人 user_id（系统操作为 0）
	CreatedAt    time.Time `gorm:"autoCreateTime;index"       json:"created_at"`
}

func (WalletTransaction) TableName() string { return "wallet_transaction" }

// WalletLog 钱包操作日志（管理员操作审计）
type WalletLog struct {
	ID         uint      `gorm:"primarykey"                 json:"id"`
	OperatorID uint      `gorm:"not null;index"             json:"operator_id"` // 操作管理员 user_id
	TargetUID  uint      `gorm:"not null;index"             json:"target_uid"`  // 被操作用户 user_id
	Action     string    `gorm:"size:32;not null;index"     json:"action"`      // add / deduct / set
	Amount     float64   `gorm:"not null"                   json:"amount"`      // 操作金额
	Before     float64   `gorm:"not null"                   json:"before"`      // 操作前余额
	After      float64   `gorm:"not null"                   json:"after"`       // 操作后余额
	Reason     string    `gorm:"size:512"                   json:"reason"`      // 操作原因
	CreatedAt  time.Time `gorm:"autoCreateTime;index"       json:"created_at"`
}

func (WalletLog) TableName() string { return "wallet_log" }

// 钱包流水类型常量
const (
	WalletRefPapReward   = "pap_reward"    // PAP 奖励
	WalletRefPapConvert  = "pap_convert"   // 联盟 PAP 月度兑换
	WalletRefManual      = "manual"        // 手动操作
	WalletRefRedeem      = "redeem"        // 兑换消费
	WalletRefAdminAdjust = "admin_adjust"  // 管理员调整
	WalletRefSrpPayout   = "srp_payout"    // SRP 补损发放
	WalletRefShopBuy     = "shop_purchase" // 商城购买
	WalletRefLotteryDraw = "lottery_draw"  // 抽奖消费
)

// 钱包操作日志动作
const (
	WalletActionAdd    = "add"    // 增加
	WalletActionDeduct = "deduct" // 扣减
	WalletActionSet    = "set"    // 设置
)
