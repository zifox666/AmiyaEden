package model

import (
	"encoding/json"
	"time"
)

type ATKillmail struct {
	KMID                 string    `gorm:"size:36;primaryKey" json:"km_id"`
	CreatedAt            time.Time `gorm:"autoCreateTime" json:"created_at"`
	SubmittedBy          uint      `gorm:"not null;index" json:"submitted_by"`
	RawText              string    `gorm:"type:text;not null" json:"raw_text"`
	NormalizedText       string    `gorm:"type:text;not null" json:"normalized_text"`
	ContentHash          string    `gorm:"size:64;not null;uniqueIndex" json:"content_hash"`
	KillmailTime         time.Time `gorm:"not null;index" json:"killmail_time"`
	VictimName           string    `gorm:"size:128;not null;index" json:"victim_name"`
	VictimNameNormalized string    `gorm:"size:128;not null;index" json:"victim_name_normalized"`
	VictimCorporation    string    `gorm:"size:128" json:"victim_corporation"`
	VictimAlliance       string    `gorm:"size:128" json:"victim_alliance"`
	VictimFaction        string    `gorm:"size:128" json:"victim_faction"`
	VictimShipName       string    `gorm:"size:128;not null" json:"victim_ship_name"`
	VictimShipTypeID     *int64    `gorm:"index" json:"victim_ship_type_id"`
	SolarSystem          string    `gorm:"size:128" json:"solar_system"`
	SecurityStatus       string    `gorm:"size:32" json:"security_status"`
	DamageTaken          int       `gorm:"not null" json:"damage_taken"`
	ResolutionStatus     string    `gorm:"size:24;not null;index" json:"resolution_status"`
	AssignedRoundID      *uint     `gorm:"index" json:"assigned_round_id"`
}

func (ATKillmail) TableName() string { return "at_km_killmail" }

type ATKillmailAttacker struct {
	BaseModel
	KMID            string `gorm:"size:36;not null;index" json:"km_id"`
	CharacterName   string `gorm:"size:128;not null;index" json:"character_name"`
	CorporationName string `gorm:"size:128" json:"corporation_name"`
	AllianceName    string `gorm:"size:128" json:"alliance_name"`
	FactionName     string `gorm:"size:128" json:"faction_name"`
	ShipName        string `gorm:"size:128" json:"ship_name"`
	ShipTypeID      *int64 `json:"ship_type_id"`
	WeaponName      string `gorm:"size:128" json:"weapon_name"`
	WeaponTypeID    *int64 `json:"weapon_type_id"`
	DamageDone      int    `gorm:"not null" json:"damage_done"`
	FinalBlow       bool   `gorm:"not null" json:"final_blow"`
}

func (ATKillmailAttacker) TableName() string { return "at_km_attacker" }

type ATKillmailAssignmentEvent struct {
	BaseModel
	KMID        string `gorm:"size:36;not null;index" json:"km_id"`
	MatchID     uint   `gorm:"not null;index" json:"match_id"`
	RoundID     uint   `gorm:"not null;index" json:"round_id"`
	ActorUserID uint   `gorm:"not null;index" json:"actor_user_id"`
	Action      string `gorm:"size:24;not null" json:"action"`
	Reason      string `gorm:"type:text" json:"reason"`
}

func (ATKillmailAssignmentEvent) TableName() string { return "at_km_assignment_event" }

type ATBRVersion struct {
	BaseModel
	RoundID         uint            `gorm:"not null;index" json:"round_id"`
	Version         int             `gorm:"not null;uniqueIndex:idx_at_br_round_version" json:"version"`
	Status          string          `gorm:"size:20;not null;index" json:"status"`
	CreatedBy       uint            `gorm:"not null;index" json:"created_by"`
	RuleSnapshot    json.RawMessage `gorm:"type:jsonb;not null" json:"rule_snapshot"`
	CleanupSnapshot json.RawMessage `gorm:"type:jsonb;not null" json:"cleanup_snapshot"`
	RevisionReason  string          `gorm:"type:text" json:"revision_reason"`
}

func (ATBRVersion) TableName() string { return "at_br_version" }

type ATBRTeamSummary struct {
	BaseModel
	BRVersionID     uint   `gorm:"not null;index;uniqueIndex:idx_at_br_team" json:"br_version_id"`
	Side            string `gorm:"size:1;not null;uniqueIndex:idx_at_br_team" json:"side"`
	DestroyedPoints int    `gorm:"not null" json:"destroyed_points"`
	HostileDamage   int    `gorm:"not null" json:"hostile_damage"`
	FriendlyDamage  int    `gorm:"not null" json:"friendly_damage"`
	EffectiveDeaths int    `gorm:"not null" json:"effective_deaths"`
}

func (ATBRTeamSummary) TableName() string { return "at_br_team_summary" }

type ATBRParticipantSnapshot struct {
	BaseModel
	BRVersionID      uint    `gorm:"not null;index" json:"br_version_id"`
	Side             string  `gorm:"size:1;not null;index" json:"side"`
	SlotOrder        int     `gorm:"not null" json:"slot_order"`
	CharacterName    string  `gorm:"size:128;not null" json:"character_name"`
	ShipTypeID       int64   `gorm:"not null" json:"ship_type_id"`
	ShipName         string  `gorm:"size:128;not null" json:"ship_name"`
	ShipPoints       int     `gorm:"not null" json:"ship_points"`
	Signatures       int     `gorm:"not null" json:"signatures"`
	FinalBlows       int     `gorm:"not null" json:"final_blows"`
	HostileDamage    int     `gorm:"not null" json:"hostile_damage"`
	FriendlyDamage   int     `gorm:"not null" json:"friendly_damage"`
	DamagePercentage float64 `gorm:"not null" json:"damage_percentage"`
	Dead             bool    `gorm:"not null" json:"dead"`
	ExcludedCleanup  bool    `gorm:"not null" json:"excluded_cleanup"`
}

func (ATBRParticipantSnapshot) TableName() string { return "at_br_participant_snapshot" }
