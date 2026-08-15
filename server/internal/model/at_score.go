package model

import "time"

const (
	ATScoreRuleDraft     = "draft"
	ATScoreRulePublished = "published"
)

type ATScoreRuleVersion struct {
	BaseModel
	Version     string     `gorm:"size:32;not null;uniqueIndex" json:"version"`
	SourceURL   string     `gorm:"size:512;not null" json:"source_url"`
	FileHash    string     `gorm:"size:64;not null;uniqueIndex" json:"file_hash"`
	Status      string     `gorm:"size:20;not null;index" json:"status"`
	CreatedBy   uint       `gorm:"not null;index" json:"created_by"`
	PublishedAt *time.Time `json:"published_at"`
}

func (ATScoreRuleVersion) TableName() string { return "at_score_rule_version" }

type ATScoreShip struct {
	BaseModel
	RuleVersionID uint   `gorm:"not null;index;uniqueIndex:idx_at_score_ship_rule_type" json:"rule_version_id"`
	TypeID        int64  `gorm:"not null;uniqueIndex:idx_at_score_ship_rule_type" json:"type_id"`
	TypeName      string `gorm:"size:128;not null" json:"type_name"`
	BasePoints    int    `gorm:"not null" json:"base_points"`
	SizeCode      string `gorm:"size:8;not null" json:"size_code"`
}

func (ATScoreShip) TableName() string { return "at_score_ship" }

type ATScoreBannedShip struct {
	BaseModel
	RuleVersionID uint   `gorm:"not null;index;uniqueIndex:idx_at_score_banned_rule_type" json:"rule_version_id"`
	TypeID        int64  `gorm:"not null;uniqueIndex:idx_at_score_banned_rule_type" json:"type_id"`
	TypeName      string `gorm:"size:128;not null" json:"type_name"`
}

func (ATScoreBannedShip) TableName() string { return "at_score_banned_ship" }

type ATScoreDuplicatePenalty struct {
	BaseModel
	RuleVersionID uint   `gorm:"not null;index;uniqueIndex:idx_at_score_penalty_rule_size" json:"rule_version_id"`
	SizeCode      string `gorm:"size:8;not null;uniqueIndex:idx_at_score_penalty_rule_size" json:"size_code"`
	Penalty       int    `gorm:"not null" json:"penalty"`
}

func (ATScoreDuplicatePenalty) TableName() string { return "at_score_duplicate_penalty" }

type ATScoreRosterResult struct {
	Format      string              `json:"format"`
	TotalPoints int                 `json:"total_points"`
	MaxPoints   int                 `json:"max_points"`
	Valid       bool                `json:"valid"`
	Slots       []ATScoreRosterSlot `json:"slots"`
	Errors      []string            `json:"errors"`
}

type ATScoreRosterSlot struct {
	TypeID          int64  `json:"type_id"`
	TypeName        string `json:"type_name"`
	BasePoints      int    `json:"base_points"`
	DuplicateCount  int    `json:"duplicate_count"`
	EffectivePoints int    `json:"effective_points"`
}
