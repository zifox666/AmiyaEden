package model

import "time"

// SRP 申请审批状态
const (
	SrpReviewPending  = "pending"  // 待审批
	SrpReviewApproved = "approved" // 已批准
	SrpReviewRejected = "rejected" // 已拒绝
)

// SRP 发放状态
const (
	SrpPayoutPending = "pending" // 待发放
	SrpPayoutPaid    = "paid"    // 已发放
)

// SrpShipPrice 舰船标准补损金额表（可由 srp/admin 编辑）
type SrpShipPrice struct {
	ID         uint      `gorm:"primarykey"              json:"id"`
	ShipTypeID int64     `gorm:"uniqueIndex;not null"    json:"ship_type_id"`
	ShipName   string    `gorm:"size:256"                json:"ship_name"`
	Amount     float64   `gorm:"not null;default:0"      json:"amount"`
	CreatedBy  uint      `gorm:"not null"                json:"created_by"`
	UpdatedBy  uint      `gorm:"not null"                json:"updated_by"`
	CreatedAt  time.Time `gorm:"autoCreateTime"          json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime"          json:"updated_at"`
}

func (SrpShipPrice) TableName() string { return "srp_ship_price" }

// SrpApplication 补损申请
type SrpApplication struct {
	ID            uint    `gorm:"primarykey"                              json:"id"`
	UserID        uint    `gorm:"not null;index"                          json:"user_id"`
	CharacterID   int64   `gorm:"not null;index"                          json:"character_id"`
	CharacterName string  `gorm:"size:128"                                json:"character_name"`
	KillmailID    int64   `gorm:"not null;index"                          json:"killmail_id"`
	FleetID       *string `gorm:"size:36;index"                           json:"fleet_id,omitempty"`
	Note          string  `gorm:"size:512"                                json:"note"`
	// 冗余 KM 信息（提交时从 EveKillmail 快照）
	ShipTypeID      int64     `gorm:"not null"                               json:"ship_type_id"`
	ShipName        string    `gorm:"size:256"                               json:"ship_name"`
	SolarSystemID   int64     `gorm:"not null"                               json:"solar_system_id"`
	SolarSystemName string    `gorm:"size:128"                               json:"solar_system_name"`
	KillmailTime    time.Time `gorm:"not null"                               json:"killmail_time"`
	CorporationID   int64     `gorm:"default:0"                              json:"corporation_id"`
	CorporationName string    `gorm:"size:256"                               json:"corporation_name"`
	AllianceID      int64     `gorm:"default:0"                              json:"alliance_id"`
	AllianceName    string    `gorm:"size:256"                               json:"alliance_name"`
	// 金额
	RecommendedAmount float64 `gorm:"not null;default:0"                     json:"recommended_amount"`
	FinalAmount       float64 `gorm:"not null;default:0"                     json:"final_amount"`
	// 审批
	ReviewStatus string     `gorm:"size:32;not null;default:'pending';index" json:"review_status"`
	ReviewedBy   *uint      `gorm:""                                         json:"reviewed_by,omitempty"`
	ReviewedAt   *time.Time `gorm:""                                         json:"reviewed_at,omitempty"`
	ReviewNote   string     `gorm:"size:512"                                 json:"review_note"`
	// 发放
	PayoutStatus string     `gorm:"size:32;not null;default:'pending';index" json:"payout_status"`
	PaidBy       *uint      `gorm:""                                         json:"paid_by,omitempty"`
	PaidAt       *time.Time `gorm:""                                         json:"paid_at,omitempty"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (SrpApplication) TableName() string { return "srp_application" }
