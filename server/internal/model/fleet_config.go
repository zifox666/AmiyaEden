package model

import "time"

// FleetConfig 舰队配置（含多个舰船装配）
type FleetConfig struct {
	ID          uint      `gorm:"primarykey"                json:"id"`
	Name        string    `gorm:"size:256;not null"         json:"name"`
	Description string    `gorm:"type:text"                 json:"description"`
	CreatedBy   uint      `gorm:"not null;index"            json:"created_by"`
	CreatedAt   time.Time `gorm:"autoCreateTime"            json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"            json:"updated_at"`
}

func (FleetConfig) TableName() string { return "fleet_config" }

// FleetConfigFitting 舰队配置中的舰船装配元数据（不再存储 EFT 文本）
type FleetConfigFitting struct {
	ID            uint    `gorm:"primarykey"                               json:"id"`
	FleetConfigID uint    `gorm:"not null;index"                           json:"fleet_config_id"`
	ShipTypeID    int64   `gorm:"not null"                                 json:"ship_type_id"`
	FittingName   string  `gorm:"size:256;not null"                        json:"fitting_name"`
	SrpAmount     float64 `gorm:"not null;default:0"                       json:"srp_amount"`
}

func (FleetConfigFitting) TableName() string { return "fleet_config_fitting" }

// 装备重要性
const (
	FittingItemRequired    = "required"    // 必要装备
	FittingItemOptional    = "optional"    // 非必要装备
	FittingItemReplaceable = "replaceable" // 可替换装备
)

// 装备不符惩罚
const (
	FittingPenaltyHalf = "half" // 补损减半
	FittingPenaltyNone = "none" // 不补损
)

// FleetConfigFittingItem 装配物品明细（与 EveCharacterFittingItem 对应）
type FleetConfigFittingItem struct {
	ID                   uint   `gorm:"primarykey"                                json:"id"`
	FleetConfigFittingID uint   `gorm:"not null;index"                            json:"fleet_config_fitting_id"`
	TypeID               int64  `gorm:"not null"                                  json:"type_id"`
	Quantity             int    `gorm:"not null;default:1"                        json:"quantity"`
	Flag                 string `gorm:"size:64;not null"                          json:"flag"`
	Importance           string `gorm:"size:32;not null;default:'required'"       json:"importance"`          // required/optional/replaceable
	Penalty              string `gorm:"size:32;not null;default:'none'"           json:"penalty"`             // half/none — 缺失时的惩罚
	ReplacementPenalty   string `gorm:"size:32;not null;default:'none'"           json:"replacement_penalty"` // half/none — 使用替代品时的惩罚
}

func (FleetConfigFittingItem) TableName() string { return "fleet_config_fitting_item" }

// FleetConfigFittingItemReplacement 可替换装备的替代品
type FleetConfigFittingItemReplacement struct {
	ID                       uint  `gorm:"primarykey"     json:"id"`
	FleetConfigFittingItemID uint  `gorm:"not null;index" json:"fleet_config_fitting_item_id"`
	TypeID                   int64 `gorm:"not null"       json:"type_id"`
}

func (FleetConfigFittingItemReplacement) TableName() string {
	return "fleet_config_fitting_item_replacement"
}
