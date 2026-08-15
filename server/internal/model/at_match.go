package model

import (
	"encoding/json"
	"time"
)

type ATSeason struct {
	BaseModel
	Name      string    `gorm:"size:128;not null" json:"name"`
	StartsAt  time.Time `gorm:"not null" json:"starts_at"`
	EndsAt    time.Time `gorm:"not null" json:"ends_at"`
	IsCurrent bool      `gorm:"not null;default:false;index" json:"is_current"`
}

func (ATSeason) TableName() string { return "at_season" }

type ATMatchType struct {
	BaseModel
	Name              string  `gorm:"size:128;not null;uniqueIndex" json:"name"`
	ContributionValue float64 `gorm:"not null;default:1" json:"contribution_value"`
	Active            bool    `gorm:"not null;default:true" json:"active"`
}

func (ATMatchType) TableName() string { return "at_match_type" }

type ATCleanupShip struct {
	BaseModel
	SeasonID uint  `gorm:"not null;index;uniqueIndex:idx_at_cleanup_season_type" json:"season_id"`
	TypeID   int64 `gorm:"not null;uniqueIndex:idx_at_cleanup_season_type" json:"type_id"`
}

func (ATCleanupShip) TableName() string { return "at_cleanup_ship" }

type ATMatch struct {
	BaseModel
	UUID                 string          `gorm:"size:36;uniqueIndex" json:"uuid"`
	SeasonID             uint            `gorm:"not null;index" json:"season_id"`
	MatchTypeID          uint            `gorm:"not null;index" json:"match_type_id"`
	Format               string          `gorm:"size:8;not null" json:"format"`
	Series               string          `gorm:"size:8;not null" json:"series"`
	TeamAName            string          `gorm:"size:128;not null" json:"team_a_name"`
	TeamBName            string          `gorm:"size:128;not null" json:"team_b_name"`
	SetID                uint            `gorm:"not null;index" json:"set_id"`
	RuleVersionID        uint            `gorm:"not null;index" json:"rule_version_id"`
	RuleSnapshot         json.RawMessage `gorm:"type:jsonb;not null" json:"rule_snapshot"`
	CleanupShipsSnapshot json.RawMessage `gorm:"type:jsonb;not null" json:"cleanup_ships_snapshot"`
	ContributionValue    float64         `gorm:"not null" json:"contribution_value"`
	FCUserID             uint            `gorm:"not null;index" json:"fc_user_id"`
	AllowSelfAssign      bool            `gorm:"not null;default:false" json:"allow_self_assign"`
	Status               string          `gorm:"size:24;not null;index" json:"status"`
	TeamAWins            int             `gorm:"not null;default:0" json:"team_a_wins"`
	TeamBWins            int             `gorm:"not null;default:0" json:"team_b_wins"`
	EndReason            string          `gorm:"type:text" json:"end_reason"`
}

func (ATMatch) TableName() string { return "at_match" }

type ATMatchRegistration struct {
	BaseModel
	MatchID                 uint            `gorm:"not null;index;uniqueIndex:idx_at_match_registration" json:"match_id"`
	UserID                  uint            `gorm:"not null;index;uniqueIndex:idx_at_match_registration" json:"user_id"`
	Status                  string          `gorm:"size:20;not null" json:"status"`
	SlotOrder               int             `gorm:"not null;default:0;index" json:"slot_order"`
	EWarTags                json.RawMessage `gorm:"type:jsonb;not null;default:'[]'" json:"ewar_tags"`
	BonusTags               json.RawMessage `gorm:"type:jsonb;not null;default:'[]'" json:"bonus_tags"`
	UserNickname            string          `gorm:"-" json:"user_nickname"`
	TournamentCharacterName string          `gorm:"-" json:"tournament_character_name"`
}

func (ATMatchRegistration) TableName() string { return "at_match_registration" }

type ATMatchRound struct {
	BaseModel
	MatchID     uint       `gorm:"not null;index" json:"match_id"`
	RoundNo     int        `gorm:"not null;index" json:"round_no"`
	Attempt     int        `gorm:"not null;default:1" json:"attempt"`
	Status      string     `gorm:"size:24;not null;index" json:"status"`
	Result      string     `gorm:"size:16" json:"result"`
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
	VoidReason  string     `gorm:"type:text" json:"void_reason"`
}

func (ATMatchRound) TableName() string { return "at_match_round" }

type ATMatchRoundRoster struct {
	BaseModel
	RoundID       uint   `gorm:"not null;index;uniqueIndex:idx_at_round_roster_slot" json:"round_id"`
	Side          string `gorm:"size:1;not null;uniqueIndex:idx_at_round_roster_slot" json:"side"`
	SlotOrder     int    `gorm:"not null;uniqueIndex:idx_at_round_roster_slot" json:"slot_order"`
	UserID        *uint  `gorm:"index" json:"user_id"`
	CharacterName string `gorm:"size:128;not null" json:"character_name"`
	ShipTypeID    int64  `gorm:"not null" json:"ship_type_id"`
}

func (ATMatchRoundRoster) TableName() string { return "at_match_round_roster" }

type ATMatchStateEvent struct {
	BaseModel
	MatchID     uint   `gorm:"not null;index" json:"match_id"`
	RoundID     *uint  `gorm:"index" json:"round_id"`
	ActorUserID uint   `gorm:"not null;index" json:"actor_user_id"`
	FromStatus  string `gorm:"size:24" json:"from_status"`
	ToStatus    string `gorm:"size:24;not null" json:"to_status"`
	Reason      string `gorm:"type:text" json:"reason"`
}

func (ATMatchStateEvent) TableName() string { return "at_match_state_event" }

type ATContributionLedger struct {
	BaseModel
	SeasonID    uint    `gorm:"not null;index" json:"season_id"`
	UserID      uint    `gorm:"not null;index" json:"user_id"`
	Amount      float64 `gorm:"not null" json:"amount"`
	SourceType  string  `gorm:"size:24;not null;index" json:"source_type"`
	SourceKey   string  `gorm:"size:128;not null;uniqueIndex" json:"source_key"`
	MatchID     *uint   `gorm:"index" json:"match_id"`
	Reason      string  `gorm:"type:text;not null" json:"reason"`
	ActorUserID uint    `gorm:"not null;index" json:"actor_user_id"`
}

func (ATContributionLedger) TableName() string { return "at_contribution_ledger" }
