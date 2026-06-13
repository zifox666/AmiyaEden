package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

const (
	FuelClaimModeAll       = "all"
	FuelClaimModeManual    = "manual"
	FuelClaimModeCondition = "condition"
	FuelClaimModeMixed     = "mixed"
)

const (
	FuelContributionUnitHour = "hour"
)

const (
	FuelCalcModeFixed   = "fixed"
	FuelCalcModePerHour = "per_hour"
)

const (
	FuelTaskStatusClaimed   = "claimed"
	FuelTaskStatusCompleted = "completed"
	FuelTaskStatusCancelled = "cancelled"
	FuelTaskStatusExpired   = "expired"
)

const (
	IskPayoutStatusPending = "pending"
	IskPayoutStatusPaid    = "paid"
	IskPayoutStatusWaived  = "waived"
)

type Int64List []int64

func (l Int64List) Value() (driver.Value, error) {
	if l == nil {
		return "[]", nil
	}
	b, err := json.Marshal(l)
	return string(b), err
}

func (l *Int64List) Scan(val interface{}) error {
	if val == nil {
		*l = nil
		return nil
	}
	var bytes []byte
	switch v := val.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	}
	return json.Unmarshal(bytes, l)
}

type StringList []string

func (l StringList) Value() (driver.Value, error) {
	if l == nil {
		return "[]", nil
	}
	b, err := json.Marshal(l)
	return string(b), err
}

func (l *StringList) Scan(val interface{}) error {
	if val == nil {
		*l = nil
		return nil
	}
	var bytes []byte
	switch v := val.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	}
	return json.Unmarshal(bytes, l)
}

// CorpStructureFuelSetting 军团建筑加油承接与贡献规则配置（每军团一条）
type CorpStructureFuelSetting struct {
	ID                   uint       `gorm:"primarykey"                         json:"id"`
	CorporationID        int64      `gorm:"uniqueIndex;not null"               json:"corporation_id"`
	Enabled              bool       `gorm:"default:false"                      json:"enabled"`
	ClaimMode            string     `gorm:"size:20;default:'all'"              json:"claim_mode"`
	ManualStructureIDs   Int64List  `gorm:"type:text"                          json:"manual_structure_ids"`
	ConditionFuelHoursLE *float64   `gorm:""                                   json:"condition_fuel_hours_le,omitempty"`
	ConditionStates      StringList `gorm:"type:text"                         json:"condition_states"`
	ContributionUnit     string     `gorm:"size:20;default:'hour'"             json:"contribution_unit"`
	WalletEnabled        bool       `gorm:"default:false"                      json:"wallet_enabled"`
	WalletCalcMode       string     `gorm:"size:20;default:'per_hour'"         json:"wallet_calc_mode"`
	WalletValue          float64    `gorm:"default:0"                          json:"wallet_value"`
	IskEnabled           bool       `gorm:"default:false"                      json:"isk_enabled"`
	IskCalcMode          string     `gorm:"size:20;default:'per_hour'"         json:"isk_calc_mode"`
	IskValue             float64    `gorm:"default:0"                          json:"isk_value"`
	UpdatedBy            uint       `gorm:"default:0;index"                    json:"updated_by"`
	CreatedAt            time.Time  `gorm:"autoCreateTime"                     json:"created_at"`
	UpdatedAt            time.Time  `gorm:"autoUpdateTime"                     json:"updated_at"`
}

func (CorpStructureFuelSetting) TableName() string { return "corp_structure_fuel_setting" }

// CorpStructureFuelTask 建筑加油承接任务
type CorpStructureFuelTask struct {
	ID                uint       `gorm:"primarykey"                                        json:"id"`
	CorporationID     int64      `gorm:"not null;index"                                    json:"corporation_id"`
	StructureID       int64      `gorm:"not null;index"                                    json:"structure_id"`
	ClaimerUserID     uint       `gorm:"not null;index"                                    json:"claimer_user_id"`
	Status            string     `gorm:"size:20;not null;index"                            json:"status"`
	BeforeFuelExpires time.Time  `gorm:"not null"                                          json:"before_fuel_expires"`
	AfterFuelExpires  *time.Time `gorm:""                                                  json:"after_fuel_expires,omitempty"`
	AddedHours        float64    `gorm:"default:0"                                         json:"added_hours"`
	WalletAmount      float64    `gorm:"default:0"                                         json:"wallet_amount"`
	IskAmount         float64    `gorm:"default:0"                                         json:"isk_amount"`
	IskPayoutStatus   string     `gorm:"size:20;default:'pending';index"                   json:"isk_payout_status"`
	IskPaidBy         *uint      `gorm:"index"                                             json:"isk_paid_by,omitempty"`
	IskPaidAt         *time.Time `gorm:""                                                  json:"isk_paid_at,omitempty"`
	IskPayoutNote     string     `gorm:"size:512"                                          json:"isk_payout_note"`
	ClaimedAt         time.Time  `gorm:"not null;index"                                    json:"claimed_at"`
	CompletedAt       *time.Time `gorm:"index"                                             json:"completed_at,omitempty"`
	CreatedAt         time.Time  `gorm:"autoCreateTime;index"                              json:"created_at"`
	UpdatedAt         time.Time  `gorm:"autoUpdateTime"                                    json:"updated_at"`
}

func (CorpStructureFuelTask) TableName() string { return "corp_structure_fuel_task" }

// CorpStructureFuelTaskListItem 建筑燃料贡献发放列表项
type CorpStructureFuelTaskListItem struct {
	ID              uint       `json:"id"`
	CorporationID   int64      `json:"corporation_id"`
	StructureID     int64      `json:"structure_id"`
	StructureName   string     `json:"structure_name"`
	ClaimerUserID   uint       `json:"claimer_user_id"`
	ClaimerName     string     `json:"claimer_name"`
	AddedHours      float64    `json:"added_hours"`
	WalletAmount    float64    `json:"wallet_amount"`
	IskAmount       float64    `json:"isk_amount"`
	IskPayoutStatus string     `json:"isk_payout_status"`
	ClaimedAt       time.Time  `json:"claimed_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	IskPaidAt       *time.Time `json:"isk_paid_at,omitempty"`
}
