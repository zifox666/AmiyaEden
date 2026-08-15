package model

import (
	"encoding/json"
	"time"
)

const (
	RecruitmentDraft     = "draft"
	RecruitmentSubmitted = "submitted"
	RecruitmentReviewing = "reviewing"
	RecruitmentApproved  = "approved"
	RecruitmentRejected  = "rejected"
	RecruitmentWithdrawn = "withdrawn"
)

type RecruitmentQuestionnaire struct {
	BaseModel
	OwnerUserID uint   `gorm:"not null;index" json:"owner_user_id"`
	Title       string `gorm:"size:200;not null" json:"title"`
	Description string `gorm:"type:text" json:"description"`
	Version     int    `gorm:"not null;default:1" json:"version"`
	Published   bool   `gorm:"not null;default:false" json:"published"`
}

func (RecruitmentQuestionnaire) TableName() string { return "recruitment_questionnaire" }

type RecruitmentField struct {
	BaseModel
	QuestionnaireID uint            `gorm:"not null;index;uniqueIndex:idx_recruitment_questionnaire_field_key" json:"questionnaire_id"`
	FieldKey        string          `gorm:"size:64;not null;uniqueIndex:idx_recruitment_questionnaire_field_key" json:"field_key"`
	Type            string          `gorm:"size:20;not null" json:"type"`
	Label           string          `gorm:"size:200;not null" json:"label"`
	Description     string          `gorm:"size:500" json:"description"`
	Required        bool            `gorm:"not null;default:false" json:"required"`
	Sort            int             `gorm:"not null;default:0" json:"sort"`
	Options         json.RawMessage `gorm:"type:jsonb" json:"options"`
}

func (RecruitmentField) TableName() string { return "recruitment_field" }

type RecruitmentLink struct {
	BaseModel
	QuestionnaireID      uint            `gorm:"not null;index" json:"questionnaire_id"`
	OwnerUserID          uint            `gorm:"not null;index" json:"owner_user_id"`
	TokenHash            string          `gorm:"size:64;not null;uniqueIndex" json:"-"`
	ExpiresAt            *time.Time      `json:"expires_at"`
	MaxSubmissions       int             `gorm:"not null;default:0" json:"max_submissions"`
	SubmissionCount      int             `gorm:"not null;default:0" json:"submission_count"`
	Active               bool            `gorm:"not null;default:true" json:"active"`
	QuestionnaireVersion int             `gorm:"not null" json:"questionnaire_version"`
	FieldsSnapshot       json.RawMessage `gorm:"type:jsonb;not null" json:"fields"`
}

func (RecruitmentLink) TableName() string { return "recruitment_link" }

type RecruitmentApplication struct {
	BaseModel
	LinkID            uint       `gorm:"not null;index;uniqueIndex:idx_recruitment_application_user" json:"link_id"`
	QuestionnaireID   uint       `gorm:"not null;index" json:"questionnaire_id"`
	UserID            uint       `gorm:"not null;index;uniqueIndex:idx_recruitment_application_user" json:"user_id"`
	Status            string     `gorm:"size:20;not null;index" json:"status"`
	MainCharacterID   int64      `gorm:"not null" json:"main_character_id"`
	MainCharacterName string     `gorm:"size:128;not null" json:"main_character_name"`
	PortraitURL       string     `gorm:"size:512" json:"portrait_url"`
	CorporationID     int64      `json:"corporation_id"`
	AllianceID        *int64     `json:"alliance_id"`
	FactionID         *int64     `json:"faction_id"`
	SubmittedAt       *time.Time `json:"submitted_at"`
}

func (RecruitmentApplication) TableName() string { return "recruitment_application" }

type RecruitmentAnswer struct {
	BaseModel
	ApplicationID uint            `gorm:"not null;index;uniqueIndex:idx_recruitment_answer_key" json:"application_id"`
	FieldKey      string          `gorm:"size:64;not null;uniqueIndex:idx_recruitment_answer_key" json:"field_key"`
	Value         json.RawMessage `gorm:"type:jsonb;not null" json:"value"`
}

func (RecruitmentAnswer) TableName() string { return "recruitment_answer" }

type RecruitmentSkillSnapshot struct {
	BaseModel
	ApplicationID      uint  `gorm:"not null;index;uniqueIndex:idx_recruitment_skill" json:"application_id"`
	SkillID            int   `gorm:"not null;uniqueIndex:idx_recruitment_skill" json:"skill_id"`
	ActiveLevel        int   `json:"active_level"`
	TrainedLevel       int   `json:"trained_level"`
	SkillpointsInSkill int64 `json:"skillpoints_in_skill"`
}

func (RecruitmentSkillSnapshot) TableName() string { return "recruitment_skill_snapshot" }

type RecruitmentReviewEvent struct {
	BaseModel
	ApplicationID uint   `gorm:"not null;index" json:"application_id"`
	ActorUserID   uint   `gorm:"not null;index" json:"actor_user_id"`
	Action        string `gorm:"size:32;not null" json:"action"`
	Reason        string `gorm:"type:text" json:"reason"`
}

func (RecruitmentReviewEvent) TableName() string { return "recruitment_review_event" }
