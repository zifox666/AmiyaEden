package model

import "time"

// SystemConfig 系统配置（key/value 键值对）
type SystemConfig struct {
	Key       string    `gorm:"primarykey;size:128"    json:"key"`
	Value     string    `gorm:"type:text;not null"     json:"value"`
	Desc      string    `gorm:"size:256"               json:"desc"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"         json:"updated_at"`
}

func (SystemConfig) TableName() string { return "system_config" }

// ─── 已知配置 Key ───

const (
	SysConfigPAPWalletPerPAP    = "pap.wallet_per_pap"   // 每 1 PAP 兑换多少系统钱包（float）
	SysConfigPAPExchangeEnabled = "pap.exchange_enabled" // PAP 兑换是否开启（bool）

	SysConfigWebhookURL           = "webhook.url"            // Webhook URL
	SysConfigWebhookEnabled       = "webhook.enabled"        // 是否启用（bool）
	SysConfigWebhookType          = "webhook.type"           // discord | feishu | dingtalk | onebot
	SysConfigWebhookFleetTemplate = "webhook.fleet_template" // 舰队行动通知模板

	SysConfigWebhookOBTargetType = "webhook.ob_target_type" // OneBot 目标类型 group | private
	SysConfigWebhookOBTargetID   = "webhook.ob_target_id"   // 目标群号或用户 QQ
	SysConfigWebhookOBToken      = "webhook.ob_token"       // access token（可空）
)
