package model

import "time"

// HallOfFameConfig 名人堂画布配置（单行单例）
type HallOfFameConfig struct {
	ID              uint      `gorm:"primarykey"       json:"id"`
	BackgroundImage string    `gorm:"type:text"        json:"background_image"` // base64 data URL
	CanvasWidth     int       `gorm:"default:1920"     json:"canvas_width"`
	CanvasHeight    int       `gorm:"default:1080"     json:"canvas_height"`
	CreatedAt       time.Time `gorm:"autoCreateTime"   json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime"   json:"updated_at"`
}

func (HallOfFameConfig) TableName() string { return "hall_of_fame_config" }

// DefaultHallOfFameConfig returns the default config singleton.
func DefaultHallOfFameConfig() HallOfFameConfig {
	return HallOfFameConfig{
		ID:           1,
		CanvasWidth:  1920,
		CanvasHeight: 1080,
	}
}

// HallOfFameCard 名人堂英雄卡片
type HallOfFameCard struct {
	BaseModel
	Name              string  `gorm:"size:256;not null"         json:"name"`
	Title             string  `gorm:"size:512"                  json:"title"`
	Description       string  `gorm:"type:text"                 json:"description"`
	CharacterID       int64   `gorm:"default:0;index"           json:"character_id"` // EVE character ID for portrait URL
	BadgeImage        string  `gorm:"type:text"                 json:"badge_image"`
	PosX              float64 `gorm:"default:10"                json:"pos_x"`  // 0-100 percentage
	PosY              float64 `gorm:"default:10"                json:"pos_y"`  // 0-100 percentage
	Width             int     `gorm:"default:200"               json:"width"`  // logical px
	Height            int     `gorm:"default:0"                 json:"height"` // 0 = auto
	StylePreset       string  `gorm:"size:32;default:'gold'"    json:"style_preset"`
	CustomBgColor     string  `gorm:"size:32"                   json:"custom_bg_color"`
	CustomTextColor   string  `gorm:"size:32"                   json:"custom_text_color"`
	CustomBorderColor string  `gorm:"size:32"                   json:"custom_border_color"`
	BorderStyle       string  `gorm:"size:32;default:'none'"    json:"border_style"`
	TitleColor        string  `gorm:"size:32"                   json:"title_color"`
	FontSize          int     `gorm:"default:0"                 json:"font_size"` // 0 = default
	ZIndex            int     `gorm:"default:0"                 json:"z_index"`
	Visible           bool    `gorm:"default:true"              json:"visible"`
}

func (HallOfFameCard) TableName() string { return "hall_of_fame_card" }

// CardLayoutUpdate is used for batch position/z-index updates.
type CardLayoutUpdate struct {
	ID     uint    `json:"id"`
	PosX   float64 `json:"pos_x"`
	PosY   float64 `json:"pos_y"`
	Width  int     `json:"width"`
	Height int     `json:"height"`
	ZIndex int     `json:"z_index"`
}
