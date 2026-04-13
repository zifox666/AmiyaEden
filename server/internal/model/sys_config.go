package model

import "time"

// ─── 准入名单 ───

const (
	AllowListAutoRole    = "auto_role"    // 允许自动权限的联盟/军团
	AllowListBasicAccess = "basic_access" // 允许基础 user 授权的联盟/军团
)

const (
	AllowEntityTypeAlliance    = "alliance"
	AllowEntityTypeCorporation = "corporation"
)

// AllowedEntity 准入名单实体（联盟或军团）
type AllowedEntity struct {
	ID         uint      `gorm:"primarykey"                                         json:"id"`
	ListType   string    `gorm:"size:32;not null;uniqueIndex:idx_allowed_entity"    json:"list_type"`
	EntityID   int64     `gorm:"not null;uniqueIndex:idx_allowed_entity"            json:"entity_id"`
	EntityType string    `gorm:"size:16;not null"                                   json:"entity_type"` // "alliance" | "corporation"
	EntityName string    `gorm:"size:256;not null"                                  json:"entity_name"`
	CreatedAt  time.Time `gorm:"autoCreateTime"                                     json:"created_at"`
}

func (AllowedEntity) TableName() string { return "allowed_entity" }

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

	SysConfigCorpID    = "corp.id"    // 军团ID (int64) - 用于获取Logo
	SysConfigSiteTitle = "site.title" // 网站标题 (string)

	// SeAT OAuth 配置
	SysConfigSeatEnabled      = "seat.enabled"       // 是否启用 SeAT 登录（bool）
	SysConfigSeatBaseURL      = "seat.base_url"      // SeAT 基础 URL，如 https://seat.winterco.org
	SysConfigSeatClientID     = "seat.client_id"     // OAuth Client ID
	SysConfigSeatClientSecret = "seat.client_secret" // OAuth Client Secret
	SysConfigSeatCallbackURL  = "seat.callback_url"  // OAuth 回调 URL
	SysConfigSeatScopes       = "seat.scopes"        // OAuth Scopes（空格分隔）

	SysConfigDefaultCorpID    int64  = 1
	SysConfigDefaultSiteTitle string = "Amiya eden"

	SysConfigDefaultSeatEnabled = "false"
	SysConfigDefaultSeatScopes  = "openid accounts groups contacts.qq passthrough"
)
