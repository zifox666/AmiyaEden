package model

import "time"

// SkillPlan 军团技能计划
type SkillPlan struct {
	ID          uint      `gorm:"primarykey"         json:"id"`
	Title       string    `gorm:"size:256;not null"  json:"title"`
	Description string    `gorm:"type:text"          json:"description"`
	ShipTypeID  *int      `gorm:"index"              json:"ship_type_id"`
	SortOrder   int       `gorm:"default:0"          json:"sort_order"`
	CreatedBy   uint      `gorm:"not null;index"     json:"created_by"`
	CreatedAt   time.Time `gorm:"autoCreateTime"     json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"     json:"updated_at"`
}

func (SkillPlan) TableName() string { return "skill_plan" }

// SkillPlanSkill 技能计划中的单条技能要求
type SkillPlanSkill struct {
	ID            uint `gorm:"primarykey"                             json:"id"`
	SkillPlanID   uint `gorm:"not null;index"                         json:"skill_plan_id"`
	SkillTypeID   int  `gorm:"not null;index"                         json:"skill_type_id"`
	RequiredLevel int  `gorm:"not null"                               json:"required_level"`
	Sort          int  `gorm:"not null;default:0"                     json:"sort"`
}

func (SkillPlanSkill) TableName() string { return "skill_plan_skill" }

// SkillPlanCheckCharacter 用户保存的技能完成度检查人物选择
type SkillPlanCheckCharacter struct {
	ID          uint  `gorm:"primarykey"                              json:"id"`
	UserID      uint  `gorm:"not null;uniqueIndex:udx_spcc_user_char" json:"user_id"`
	CharacterID int64 `gorm:"not null;uniqueIndex:udx_spcc_user_char" json:"character_id"`
}

func (SkillPlanCheckCharacter) TableName() string { return "skill_plan_check_character" }

// SkillPlanCheckPlan 用户保存的技能完成度检查计划选择
type SkillPlanCheckPlan struct {
	ID     uint `gorm:"primarykey"                              json:"id"`
	UserID uint `gorm:"not null;uniqueIndex:udx_spcp_user_plan" json:"user_id"`
	PlanID uint `gorm:"not null;uniqueIndex:udx_spcp_user_plan" json:"plan_id"`
}

func (SkillPlanCheckPlan) TableName() string { return "skill_plan_check_plan" }
