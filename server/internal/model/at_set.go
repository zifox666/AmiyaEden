package model

import "encoding/json"

type ATSet struct {
	BaseModel
	Name          string `gorm:"size:200;not null" json:"name"`
	Format        string `gorm:"size:8;not null;index" json:"format"`
	RuleVersionID uint   `gorm:"not null;index" json:"rule_version_id"`
	CreatedBy     uint   `gorm:"not null;index" json:"created_by"`
	Status        string `gorm:"size:20;not null;default:'draft'" json:"status"`
}

func (ATSet) TableName() string { return "at_set" }

// ATFit is a reusable tournament fitting. Base tags describe capabilities
// already present in BaseEFT; replacement rules describe optional tag variants.
// Set slots reference a fit and only store the final tags selected for the slot.
type ATFit struct {
	BaseModel
	Name          string          `gorm:"size:240;not null;index" json:"name"`
	ShipTypeID    int64           `gorm:"not null;index" json:"ship_type_id"`
	ShipName      string          `gorm:"size:128;not null;index" json:"ship_name"`
	BaseEFT       string          `gorm:"type:text;not null" json:"base_eft"`
	BaseEWarTags  json.RawMessage `gorm:"type:jsonb;not null" json:"base_ewar_tags"`
	BaseBonusTags json.RawMessage `gorm:"type:jsonb;not null" json:"base_bonus_tags"`
	CreatedBy     uint            `gorm:"not null;index" json:"created_by"`
	Status        string          `gorm:"size:20;not null;default:'active';index" json:"status"`
}

func (ATFit) TableName() string { return "at_fit" }

type ATSetSlot struct {
	BaseModel
	SetID      uint            `gorm:"not null;index;uniqueIndex:idx_at_set_slot_order" json:"set_id"`
	SlotOrder  int             `gorm:"not null;uniqueIndex:idx_at_set_slot_order" json:"slot_order"`
	FitID      *uint           `gorm:"index" json:"fit_id"`
	ShipTypeID int64           `gorm:"not null" json:"ship_type_id"`
	BaseEFT    string          `gorm:"type:text;not null" json:"base_eft"`
	EWarTags   json.RawMessage `gorm:"type:jsonb;not null" json:"ewar_tags"`
	BonusTags  json.RawMessage `gorm:"type:jsonb;not null" json:"bonus_tags"`
}

func (ATSetSlot) TableName() string { return "at_set_slot" }

type ATFitReplacement struct {
	BaseModel
	SetSlotID         uint   `gorm:"index" json:"set_slot_id"`
	FitID             *uint  `gorm:"index" json:"fit_id"`
	TagType           string `gorm:"size:16;not null" json:"tag_type"`
	Tag               string `gorm:"size:64;not null" json:"tag"`
	Operation         string `gorm:"size:16;not null" json:"operation"`
	TargetTypeID      int64  `gorm:"not null;default:0" json:"target_type_id"`
	ReplacementTypeID int64  `gorm:"not null;default:0" json:"replacement_type_id"`
	Quantity          int    `gorm:"not null;default:1" json:"quantity"`
}

func (ATFitReplacement) TableName() string { return "at_fit_replacement" }
