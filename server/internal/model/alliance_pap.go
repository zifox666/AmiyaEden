package model

import "time"

// AlliancePAPRecord 联盟 PAP 舰队参与明细记录（以 fleet_id + character_id 联合唯一）
type AlliancePAPRecord struct {
	ID            uint       `gorm:"primarykey"                                         json:"id"`
	MainCharacter string     `gorm:"size:128;not null;index:idx_apr_main_ym"            json:"main_character"`
	CharacterID   string     `gorm:"size:32;not null;uniqueIndex:idx_apr_fleet_char"    json:"character_id"`
	CharacterName string     `gorm:"size:128"                                           json:"character_name"`
	FleetID       string     `gorm:"size:32;not null;uniqueIndex:idx_apr_fleet_char"    json:"fleet_id"`
	Year          int        `gorm:"not null;index:idx_apr_main_ym"                     json:"year"`
	Month         int        `gorm:"not null;index:idx_apr_main_ym"                     json:"month"`
	StartAt       time.Time  `gorm:"not null"                                           json:"start_at"`
	EndAt         *time.Time `gorm:""                                                  json:"end_at,omitempty"`
	Title         string     `gorm:"size:256"                                           json:"title"`
	Level         string     `gorm:"size:64"                                            json:"level"`
	Pap           float64    `gorm:"type:decimal(10,2);not null;default:0"              json:"pap"`
	ShipGroupID   string     `gorm:"size:32"                                            json:"ship_group_id"`
	ShipGroupName string     `gorm:"size:128"                                           json:"ship_group_name"`
	ShipTypeID    string     `gorm:"size:32"                                            json:"ship_type_id"`
	ShipTypeName  string     `gorm:"size:256"                                           json:"ship_type_name"`
	IsArchived    bool       `gorm:"not null;default:false"                             json:"is_archived"`
	CreatedAt     time.Time  `gorm:"autoCreateTime"                                     json:"created_at"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime"                                     json:"updated_at"`
}

func (AlliancePAPRecord) TableName() string { return "alliance_pap_record" }

// AlliancePAPSummary 联盟 PAP 月度汇总（主角色 + 年 + 月 唯一）
type AlliancePAPSummary struct {
	ID                uint      `gorm:"primarykey"                                                   json:"id"`
	MainCharacter     string    `gorm:"size:128;not null;uniqueIndex:idx_aps_main_ym"              json:"main_character"`
	Year              int       `gorm:"not null;uniqueIndex:idx_aps_main_ym"                       json:"year"`
	Month             int       `gorm:"not null;uniqueIndex:idx_aps_main_ym"                       json:"month"`
	CorporationID     string    `gorm:"size:32"                                                     json:"corporation_id"`
	TotalPap          float64   `gorm:"type:decimal(10,2);not null;default:0"                       json:"total_pap"`
	YearlyTotalPap    float64   `gorm:"type:decimal(10,2);not null;default:0"                       json:"yearly_total_pap"`
	MonthlyRank       int       `gorm:"default:0"                                                   json:"monthly_rank"`
	YearlyRank        int       `gorm:"default:0"                                                   json:"yearly_rank"`
	GlobalMonthlyRank int       `gorm:"default:0"                                                   json:"global_monthly_rank"`
	GlobalYearlyRank  int       `gorm:"default:0"                                                   json:"global_yearly_rank"`
	TotalInCorp       int       `gorm:"default:0"                                                   json:"total_in_corp"`
	TotalGlobal       int       `gorm:"default:0"                                                   json:"total_global"`
	CalculatedAt      time.Time `gorm:""                                                             json:"calculated_at"`
	IsArchived        bool      `gorm:"not null;default:false"                                       json:"is_archived"`
	CreatedAt         time.Time `gorm:"autoCreateTime"                                               json:"created_at"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime"                                               json:"updated_at"`
}

func (AlliancePAPSummary) TableName() string { return "alliance_pap_summary" }
