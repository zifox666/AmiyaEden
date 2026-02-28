package model

import "time"

// Fleet 重要等级
const (
	FleetImportanceStratOp = "strat_op" // 战略行动
	FleetImportanceCTA     = "cta"      // 全面集结
	FleetImportanceOther   = "other"    // 其他
)

// Fleet 舰队记录
type Fleet struct {
	ID              string     `gorm:"primaryKey;size:36"         json:"id"`
	Title           string     `gorm:"size:256;not null"          json:"title"`
	Description     string     `gorm:"type:text"                  json:"description"`
	StartAt         time.Time  `gorm:"not null"                   json:"start_at"`
	EndAt           time.Time  `gorm:"not null"                   json:"end_at"`
	Importance      string     `gorm:"size:32;not null"           json:"importance"` // strat_op / cta / other
	PapCount        float64    `gorm:"default:0"                  json:"pap_count"`
	FCUserID        uint       `gorm:"not null;index"             json:"fc_user_id"`
	FCCharacterID   int64      `gorm:"not null"                   json:"fc_character_id"`
	FCCharacterName string     `gorm:"size:128"                   json:"fc_character_name"`
	ESIFleetID      *int64     `gorm:""                           json:"esi_fleet_id,omitempty"`
	CreatedAt       time.Time  `gorm:"autoCreateTime"             json:"created_at"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime"             json:"updated_at"`
	DeletedAt       *time.Time `gorm:"index"                      json:"deleted_at,omitempty"`
}

func (Fleet) TableName() string { return "fleet" }

// FleetMember 舰队成员记录（参与过的成员快照）
type FleetMember struct {
	ID            uint      `gorm:"primarykey"                                       json:"id"`
	FleetID       string    `gorm:"size:36;not null;index:idx_fleet_member,unique"   json:"fleet_id"`
	CharacterID   int64     `gorm:"not null;index:idx_fleet_member,unique"           json:"character_id"`
	CharacterName string    `gorm:"size:128"                                         json:"character_name"`
	UserID        uint      `gorm:"not null;index"                                   json:"user_id"`
	ShipTypeID    *int64    `gorm:""                                                 json:"ship_type_id,omitempty"`
	SolarSystemID *int64    `gorm:""                                                 json:"solar_system_id,omitempty"`
	JoinedAt      time.Time `gorm:"autoCreateTime"                                   json:"joined_at"`
}

func (FleetMember) TableName() string { return "fleet_member" }

// FleetPapLog PAP 发放记录
type FleetPapLog struct {
	ID          uint      `gorm:"primarykey"                                       json:"id"`
	FleetID     string    `gorm:"size:36;not null;index"                           json:"fleet_id"`
	CharacterID int64     `gorm:"not null;index"                                   json:"character_id"`
	UserID      uint      `gorm:"not null;index"                                   json:"user_id"`
	PapCount    float64   `gorm:"not null"                                         json:"pap_count"`
	IssuedBy    uint      `gorm:"not null"                                         json:"issued_by"` // 发放者 user_id
	IssuedAt    time.Time `gorm:"autoCreateTime"                                   json:"issued_at"`
}

func (FleetPapLog) TableName() string { return "fleet_pap_log" }

// FleetInvite 舰队邀请链接
type FleetInvite struct {
	ID        uint      `gorm:"primarykey"                 json:"id"`
	FleetID   string    `gorm:"size:36;not null;index"     json:"fleet_id"`
	Code      string    `gorm:"size:64;uniqueIndex"        json:"code"` // 邀请码
	Active    bool      `gorm:"default:true"               json:"active"`
	CreatedAt time.Time `gorm:"autoCreateTime"             json:"created_at"`
	ExpiresAt time.Time `gorm:"not null"                   json:"expires_at"`
}

func (FleetInvite) TableName() string { return "fleet_invite" }
