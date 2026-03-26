package model

import (
	"strings"
	"time"
)

// PAPTypeRate PAP 类型兑换汇率（每种 PAP 类型对应的系统钱包兑换比率）
type PAPTypeRate struct {
	PapType     string    `gorm:"primarykey;size:64"                    json:"pap_type"`
	DisplayName string    `gorm:"size:128;not null"                     json:"display_name"`
	Rate        float64   `gorm:"type:decimal(10,2);not null;default:1" json:"rate"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"                        json:"updated_at"`
}

func (PAPTypeRate) TableName() string { return "pap_type_rate" }

// PAP 类型常量
const (
	PAPTypeSkirmish = "skirmish" // 游击队
	PAPTypeStratOp  = "strat_op" // 战略行动
	PAPTypeCTA      = "cta"      // 全面集结
)

// DefaultPAPTypeRates 默认 PAP 类型兑换汇率
var DefaultPAPTypeRates = []PAPTypeRate{
	{PapType: PAPTypeSkirmish, DisplayName: "Skirmish（游击队）", Rate: 10},
	{PapType: PAPTypeStratOp, DisplayName: "Strategic（战略行动）", Rate: 30},
	{PapType: PAPTypeCTA, DisplayName: "CTA（全面集结）", Rate: 50},
}

// NormalizePAPLevel 将外部 PAP API 返回的 Level 字符串映射到系统内部 PAP 类型常量
func NormalizePAPLevel(level string) string {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "cta", "cta op":
		return PAPTypeCTA
	case "strat op", "strat_op", "strategic", "strategic op":
		return PAPTypeStratOp
	default:
		return PAPTypeSkirmish
	}
}
