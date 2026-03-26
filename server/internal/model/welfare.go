package model

import "time"

// ─────────────────────────────────────────────
//  军团福利系统
// ─────────────────────────────────────────────

// ─── 发放模式 ───

const (
	WelfareDistModePerUser      = "per_user"      // 按自然人（User Account）发放
	WelfareDistModePerCharacter = "per_character"  // 按人物（EVE Character）发放
)

// ─── 状态 ───

const (
	WelfareStatusActive   int8 = 1 // 启用
	WelfareStatusDisabled int8 = 0 // 停用
)

// ─── 数据模型 ───

// Welfare 福利定义
type Welfare struct {
	BaseModel
	Name             string `gorm:"size:256;not null"           json:"name"`
	Description      string `gorm:"type:text"                   json:"description"`
	DistMode         string `gorm:"size:20;not null;default:'per_user'" json:"dist_mode"`
	RequireSkillPlan bool   `gorm:"default:false"               json:"require_skill_plan"`
	MaxCharAgeMonths *int   `gorm:""                            json:"max_char_age_months"`
	RequireEvidence  bool   `gorm:"default:false"               json:"require_evidence"`
	ExampleEvidence  string `gorm:"type:text"                   json:"example_evidence"`
	Status           int8   `gorm:"default:1"                   json:"status"`
	CreatedBy        uint   `gorm:"not null"                    json:"created_by"`

	// 虚拟字段，不存库，由业务层填充
	SkillPlanIDs []uint `gorm:"-" json:"skill_plan_ids"`
}

// WelfareSkillPlan 福利-技能计划关联表
type WelfareSkillPlan struct {
	WelfareID   uint `gorm:"primaryKey" json:"welfare_id"`
	SkillPlanID uint `gorm:"primaryKey" json:"skill_plan_id"`
}

func (WelfareSkillPlan) TableName() string { return "welfare_skill_plans" }

func (Welfare) TableName() string { return "welfare" }

// ─── 申请状态 ───

const (
	WelfareAppStatusRequested = "requested"
	WelfareAppStatusDelivered  = "delivered"
	WelfareAppStatusRejected   = "rejected"
)

// WelfareApplication 福利申请记录
type WelfareApplication struct {
	BaseModel
	WelfareID     uint       `gorm:"not null;index"            json:"welfare_id"`
	UserID        *uint      `gorm:"index"                     json:"user_id"`
	CharacterID   int64      `gorm:"not null"                  json:"character_id"`
	CharacterName string     `gorm:"size:128"                  json:"character_name"`
	QQ            string     `gorm:"size:20"                   json:"qq"`
	DiscordID     string     `gorm:"size:20"                   json:"discord_id"`
	EvidenceImage string     `gorm:"type:text"                 json:"evidence_image"`
	Status        string     `gorm:"size:20;default:'requested';index" json:"status"`
	ReviewedBy    uint       `gorm:"default:0"                 json:"reviewed_by"`
	ReviewedAt    *time.Time `gorm:""                          json:"reviewed_at"`
}

func (WelfareApplication) TableName() string { return "welfare_application" }
