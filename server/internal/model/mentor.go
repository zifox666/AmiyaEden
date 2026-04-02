package model

import "time"

const (
	MentorRelationStatusPending   = "pending"
	MentorRelationStatusActive    = "active"
	MentorRelationStatusRejected  = "rejected"
	MentorRelationStatusRevoked   = "revoked"
	MentorRelationStatusGraduated = "graduated"
)

const (
	MentorConditionSkillPoints = "skill_points"
	MentorConditionPapCount    = "pap_count"
	MentorConditionDaysActive  = "days_active"
)

type MentorMenteeRelationship struct {
	BaseModel
	MenteeUserID                    uint       `gorm:"not null;index"                         json:"mentee_user_id"`
	MenteePrimaryCharacterIDAtStart int64      `gorm:"not null"                               json:"mentee_primary_character_id_at_start"`
	MentorUserID                    uint       `gorm:"not null;index"                         json:"mentor_user_id"`
	Status                          string     `gorm:"size:20;not null;default:pending;index" json:"status"`
	AppliedAt                       time.Time  `gorm:"not null;index"                         json:"applied_at"`
	RespondedAt                     *time.Time `json:"responded_at"`
	RevokedAt                       *time.Time `json:"revoked_at"`
	RevokedBy                       *uint      `json:"revoked_by"`
	GraduatedAt                     *time.Time `json:"graduated_at"`
}

func (MentorMenteeRelationship) TableName() string { return "mentor_mentee_relationship" }

type MentorRewardStage struct {
	BaseModel
	StageOrder    int     `gorm:"not null;uniqueIndex"      json:"stage_order"`
	Name          string  `gorm:"size:128;not null"         json:"name"`
	ConditionType string  `gorm:"size:32;not null"          json:"condition_type"`
	Threshold     float64 `gorm:"not null"                  json:"threshold"`
	RewardAmount  float64 `gorm:"not null"                  json:"reward_amount"`
}

func (MentorRewardStage) TableName() string { return "mentor_reward_stage" }

type MentorRewardDistribution struct {
	BaseModel
	RelationshipID      uint      `gorm:"not null;index;uniqueIndex:idx_mrd_rel_stage_order" json:"relationship_id"`
	StageID             uint      `gorm:"not null;index"                                      json:"stage_id"`
	StageOrder          int       `gorm:"not null;uniqueIndex:idx_mrd_rel_stage_order"       json:"stage_order"`
	MentorUserID        uint      `gorm:"not null;index"                                      json:"mentor_user_id"`
	MentorCharacterName string    `gorm:"size:255;not null;default:''"                       json:"mentor_character_name"`
	MentorNickname      string    `gorm:"size:255;not null;default:''"                       json:"mentor_nickname"`
	MenteeUserID        uint      `gorm:"not null;index"                                      json:"mentee_user_id"`
	MenteeCharacterName string    `gorm:"size:255;not null;default:''"                       json:"mentee_character_name"`
	MenteeNickname      string    `gorm:"size:255;not null;default:''"                       json:"mentee_nickname"`
	RewardAmount        float64   `gorm:"not null"                                            json:"reward_amount"`
	DistributedAt       time.Time `gorm:"not null;index"                                      json:"distributed_at"`
	WalletRefID         string    `gorm:"size:128;not null;uniqueIndex"                       json:"wallet_ref_id"`
}

func (MentorRewardDistribution) TableName() string { return "mentor_reward_distribution" }
