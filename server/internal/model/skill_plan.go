package model

// SkillPlan 技能规划
type SkillPlan struct {
	BaseModel
	Name        string `gorm:"size:200;not null"  json:"name"`
	Description string `gorm:"type:text"          json:"description"`
	CreatedBy   uint   `gorm:"not null;index"     json:"created_by"`
}

func (SkillPlan) TableName() string { return "skill_plan" }

// SkillPlanItem 技能规划条目
type SkillPlanItem struct {
	ID            uint `gorm:"primarykey"          json:"id"`
	SkillPlanID   uint `gorm:"not null;index"      json:"skill_plan_id"`
	SkillTypeID   int  `gorm:"not null"            json:"skill_type_id"`
	RequiredLevel int  `gorm:"not null"            json:"required_level"`
}

func (SkillPlanItem) TableName() string { return "skill_plan_item" }
