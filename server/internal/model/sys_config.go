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
	SysConfigPAPWalletPerPAP    = "pap.wallet_per_pap"   // 每 1 PAP 兑换多少伏羲币（float）
	SysConfigPAPExchangeEnabled = "pap.exchange_enabled" // PAP 兑换是否开启（bool）
	SysConfigPAPFCSalary        = "pap.fc_salary"        // FC 工资（float）
	SysConfigPAPFCSalaryLimit   = "pap.fc_salary_limit"  // FC 工资每月上限次数（int）

	SysConfigWebhookURL           = "webhook.url"            // Webhook URL
	SysConfigWebhookEnabled       = "webhook.enabled"        // 是否启用（bool）
	SysConfigWebhookType          = "webhook.type"           // discord | feishu | dingtalk | onebot
	SysConfigWebhookFleetTemplate = "webhook.fleet_template" // 舰队行动通知模板

	SysConfigWebhookOBTargetType = "webhook.ob_target_type" // OneBot 目标类型 group | private
	SysConfigWebhookOBTargetID   = "webhook.ob_target_id"   // 目标群号或用户 QQ
	SysConfigWebhookOBToken      = "webhook.ob_token"       // access token（可空）

	SysConfigSDEAPIKey      = "sde.api_key"      // SDE 查询 API Key
	SysConfigSDEProxy       = "sde.proxy"        // SDE 下载代理
	SysConfigSDEDownloadURL = "sde.download_url" // SDE 下载地址

	SysConfigAllowCorporations = "app.allow_corporations" // 允许访问的公司 ID 列表 (JSON 数组)

	SysConfigNewbroMaxCharacterSP          = "newbro.max_character_sp"
	SysConfigNewbroMultiCharacterSP        = "newbro.multi_character_sp"
	SysConfigNewbroMultiCharacterThreshold = "newbro.multi_character_threshold"
	SysConfigNewbroRefreshIntervalDays     = "newbro.refresh_interval_days"
	SysConfigNewbroBonusRate               = "newbro.bonus_rate"

	SysConfigDefaultSDEAPIKey      = "modify_your_api_key"
	SysConfigDefaultSDEProxy       = ""
	SysConfigDefaultSDEDownloadURL = "https://api.github.com/repos/garveen/eve-sde-converter/releases/latest"

	SysConfigDefaultNewbroMaxCharacterSP          int64   = 20_000_000
	SysConfigDefaultNewbroMultiCharacterSP        int64   = 10_000_000
	SysConfigDefaultNewbroMultiCharacterThreshold         = 3
	SysConfigDefaultNewbroRefreshIntervalDays             = 7
	SysConfigDefaultNewbroBonusRate               float64 = 20
	SysConfigDefaultPAPFCSalary                   float64 = 400
	SysConfigDefaultPAPFCSalaryLimit              int     = 5
)
