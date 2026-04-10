package model

import (
	"time"
)

// EveCharacter EVE 人物模型，一个 User 可绑定多个人物
type EveCharacter struct {
	BaseModel
	CharacterID   int64     `gorm:"uniqueIndex;not null"   json:"character_id"`
	CharacterName string    `gorm:"size:128;not null"      json:"character_name"`
	UserID        uint      `gorm:"not null;index"         json:"user_id"`
	AccessToken   string    `gorm:"type:text"              json:"-"`
	RefreshToken  string    `gorm:"type:text"              json:"-"`
	TokenExpiry   time.Time `gorm:""                       json:"token_expiry"`
	Scopes        string    `gorm:"type:text"              json:"scopes"` // 空格分隔的 scope 列表
	TokenInvalid  bool      `gorm:"not null;default:false" json:"token_invalid"`

	// ESI 公开信息
	Birthday             *time.Time `gorm:""                         json:"birthday,omitempty"`
	FuxiLegionTenureDays *int       `gorm:""                         json:"fuxi_legion_tenure_days,omitempty"`

	// ESI Affiliation 归属信息
	CorporationID int64  `gorm:"default:0;index"         json:"corporation_id"`
	AllianceID    *int64 `gorm:""                         json:"alliance_id,omitempty"`
	FactionID     *int64 `gorm:""                         json:"faction_id,omitempty"`
}

func (EveCharacter) TableName() string {
	return "eve_character"
}
