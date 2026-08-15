package model

import "time"

// CombatLogReport stores the server-side analysis result for one uploaded EVE game log.
// The original log content is intentionally not persisted.
type CombatLogReport struct {
	BaseModel
	UserID           uint      `gorm:"index;uniqueIndex:idx_combat_log_user_hash,priority:1" json:"user_id"`
	OriginalFilename string    `gorm:"size:255" json:"original_filename"`
	FileHash         string    `gorm:"size:64;uniqueIndex:idx_combat_log_user_hash,priority:2" json:"-"`
	CharacterName    string    `gorm:"size:128;index" json:"character_name"`
	Language         string    `gorm:"size:16" json:"language"`
	StartedAt        time.Time `gorm:"index" json:"started_at"`
	EndedAt          time.Time `json:"ended_at"`
	DurationSeconds  int64     `json:"duration_seconds"`
	EventCount       int       `json:"event_count"`
	SourceBytes      int64     `json:"source_bytes"`
	DamageOut        int64     `json:"damage_out"`
	DamageIn         int64     `json:"damage_in"`
	LogisticsOut     int64     `json:"logistics_out"`
	LogisticsIn      int64     `json:"logistics_in"`
	CapTransferOut   int64     `json:"cap_transfer_out"`
	CapTransferIn    int64     `json:"cap_transfer_in"`
	CapDamageOut     int64     `json:"cap_damage_out"`
	CapDamageIn      int64     `json:"cap_damage_in"`
	MiningUnits      int64     `json:"mining_units"`
	MiningM3         float64   `json:"mining_m3"`
}

func (CombatLogReport) TableName() string { return "combat_log_report" }

// CombatLogBreakdown stores an aggregated category/pilot/ship/weapon row.
type CombatLogBreakdown struct {
	ID         uint    `gorm:"primarykey" json:"id"`
	ReportID   uint    `gorm:"index" json:"report_id"`
	Category   string  `gorm:"size:32;index" json:"category"`
	PilotName  string  `gorm:"size:256" json:"pilot_name"`
	ShipType   string  `gorm:"size:256" json:"ship_type"`
	WeaponType string  `gorm:"size:256" json:"weapon_type"`
	Amount     int64   `json:"amount"`
	VolumeM3   float64 `json:"volume_m3"`
	EventCount int     `json:"event_count"`
}

func (CombatLogBreakdown) TableName() string { return "combat_log_breakdown" }

// CombatLogTimeline stores per-second aggregated values used by the web chart.
type CombatLogTimeline struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	ReportID   uint      `gorm:"index" json:"report_id"`
	OccurredAt time.Time `gorm:"index" json:"occurred_at"`
	Category   string    `gorm:"size:32;index" json:"category"`
	Amount     int64     `json:"amount"`
}

func (CombatLogTimeline) TableName() string { return "combat_log_timeline" }

// CombatLogHitQuality stores the hit-result distribution for outgoing and
// incoming damage. Percentage is calculated when the detail is read.
type CombatLogHitQuality struct {
	ID           uint    `gorm:"primarykey" json:"id"`
	ReportID     uint    `gorm:"index;uniqueIndex:idx_combat_log_hit_quality,priority:1" json:"report_id"`
	Direction    string  `gorm:"size:16;index;uniqueIndex:idx_combat_log_hit_quality,priority:2" json:"direction"`
	Quality      string  `gorm:"size:32;uniqueIndex:idx_combat_log_hit_quality,priority:3" json:"quality"`
	HitCount     int     `json:"hit_count"`
	DamageAmount int64   `json:"damage_amount"`
	Percentage   float64 `gorm:"-" json:"percentage"`
}

func (CombatLogHitQuality) TableName() string { return "combat_log_hit_quality" }
