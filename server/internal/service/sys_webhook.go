package service

import (
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const defaultFleetTemplate = "@all 舰队行动通知\n行动: {title}\n指挥官: {fc_name}\n重要程度: {importance}\nPAP: {pap_count}\n时间: {start_at} ~ {end_at}\n{description}"

// WebhookConfig Webhook 配置
type WebhookConfig struct {
	URL           string `json:"url"`
	Enabled       bool   `json:"enabled"`
	Type          string `json:"type"`           // discord | feishu | dingtalk | onebot
	FleetTemplate string `json:"fleet_template"` // 舰队行动通知模板
	OBTargetType  string `json:"ob_target_type"` // group | private
	OBTargetID    int64  `json:"ob_target_id"`   // 目标群号或用户 QQ
	OBToken       string `json:"ob_token"`       // access token（可空）
}

// WebhookService Webhook 业务逻辑层
type WebhookService struct {
	repo *repository.SysConfigRepository
	http *http.Client
}

func NewWebhookService() *WebhookService {
	return &WebhookService{
		repo: repository.NewSysConfigRepository(),
		http: &http.Client{Timeout: 10 * time.Second},
	}
}

// GetConfig 获取 Webhook 配置
func (s *WebhookService) GetConfig() (*WebhookConfig, error) {
	url, _ := s.repo.Get(model.SysConfigWebhookURL, "")
	enabled := s.repo.GetBool(model.SysConfigWebhookEnabled, false)
	wtype, _ := s.repo.Get(model.SysConfigWebhookType, "discord")
	tmpl, _ := s.repo.Get(model.SysConfigWebhookFleetTemplate, defaultFleetTemplate)
	obTargetType, _ := s.repo.Get(model.SysConfigWebhookOBTargetType, "group")
	obTargetIDStr, _ := s.repo.Get(model.SysConfigWebhookOBTargetID, "0")
	obTargetID, _ := strconv.ParseInt(obTargetIDStr, 10, 64)
	obToken, _ := s.repo.Get(model.SysConfigWebhookOBToken, "")
	return &WebhookConfig{
		URL:           url,
		Enabled:       enabled,
		Type:          wtype,
		FleetTemplate: tmpl,
		OBTargetType:  obTargetType,
		OBTargetID:    obTargetID,
		OBToken:       obToken,
	}, nil
}

// SetConfig 保存 Webhook 配置
func (s *WebhookService) SetConfig(cfg *WebhookConfig) error {
	type kv struct{ k, v, d string }
	entries := []kv{
		{model.SysConfigWebhookURL, cfg.URL, "Webhook URL"},
		{model.SysConfigWebhookEnabled, fmt.Sprintf("%v", cfg.Enabled), "Webhook 是否启用"},
		{model.SysConfigWebhookType, cfg.Type, "Webhook 类型 (discord/feishu/dingtalk/onebot)"},
		{model.SysConfigWebhookFleetTemplate, cfg.FleetTemplate, "舰队行动通知模板"},
		{model.SysConfigWebhookOBTargetType, cfg.OBTargetType, "OneBot 目标类型 (group/private)"},
		{model.SysConfigWebhookOBTargetID, fmt.Sprintf("%d", cfg.OBTargetID), "OneBot 目标 ID"},
		{model.SysConfigWebhookOBToken, cfg.OBToken, "OneBot Access Token"},
	}
	for _, e := range entries {
		if err := s.repo.Set(e.k, e.v, e.d); err != nil {
			return err
		}
	}
	return nil
}

// SendFleetPing 发送舰队行动 Ping（若未启用则静默忽略）
func (s *WebhookService) SendFleetPing(fleet *model.Fleet) error {
	cfg, err := s.GetConfig()
	if err != nil || !cfg.Enabled || cfg.URL == "" {
		return nil
	}

	importanceLabel := map[string]string{
		model.FleetImportanceStratOp: "战略行动",
		model.FleetImportanceCTA:     "全面集结",
		model.FleetImportanceOther:   "其他行动",
	}[fleet.Importance]
	if importanceLabel == "" {
		importanceLabel = fleet.Importance
	}

	desc := fleet.Description
	if desc == "" {
		desc = "-"
	}

	content := cfg.FleetTemplate
	content = strings.ReplaceAll(content, "{title}", fleet.Title)
	content = strings.ReplaceAll(content, "{fc_name}", fleet.FCCharacterName)
	content = strings.ReplaceAll(content, "{importance}", importanceLabel)
	content = strings.ReplaceAll(content, "{pap_count}", fmt.Sprintf("%.0f", fleet.PapCount))
	content = strings.ReplaceAll(content, "{start_at}", fleet.StartAt.Local().Format("01/02 15:04"))
	content = strings.ReplaceAll(content, "{end_at}", fleet.EndAt.Local().Format("01/02 15:04"))
	content = strings.ReplaceAll(content, "{description}", desc)

	// 舰队配置信息
	fleetConfigInfo := ""
	if fleet.FleetConfigID != nil && *fleet.FleetConfigID > 0 {
		fcRepo := repository.NewFleetConfigRepository()
		if fc, fcErr := fcRepo.GetByID(*fleet.FleetConfigID); fcErr == nil {
			fittings, _ := fcRepo.ListFittingsByConfigID(fc.ID)
			fleetConfigInfo = fc.Name
			if len(fittings) > 0 {
				var names []string
				for _, f := range fittings {
					names = append(names, f.FittingName)
				}
				fleetConfigInfo += "\n  " + strings.Join(names, "\n  ")
			}
		}
	}
	content = strings.ReplaceAll(content, "{fleet_config}", fleetConfigInfo)

	return s.sendMessage(cfg, content)
}

// SendTest 发送测试消息
func (s *WebhookService) SendTest(cfg *WebhookConfig, content string) error {
	if cfg.Type == "" {
		cfg.Type = "discord"
	}
	if content == "" {
		content = "✅ Webhook 测试消息（来自 AmiyaEden）"
	}
	return s.sendMessage(cfg, content)
}

func (s *WebhookService) sendMessage(cfg *WebhookConfig, content string) error {
	var body []byte
	var err error
	var reqURL string

	switch cfg.Type {
	case "feishu":
		reqURL = cfg.URL
		body, err = json.Marshal(map[string]any{
			"msg_type": "text",
			"content":  map[string]string{"text": content},
		})
	case "dingtalk":
		reqURL = cfg.URL
		body, err = json.Marshal(map[string]any{
			"msgtype": "text",
			"text":    map[string]string{"content": content},
		})
	case "onebot":
		endpoint := "/send_group_msg"
		var postBody map[string]any
		if cfg.OBTargetType == "private" {
			endpoint = "/send_private_msg"
			postBody = map[string]any{
				"user_id": cfg.OBTargetID,
				"message": content,
			}
		} else {
			postBody = map[string]any{
				"group_id": cfg.OBTargetID,
				"message":  content,
			}
		}
		reqURL = strings.TrimRight(cfg.URL, "/") + endpoint
		body, err = json.Marshal(postBody)
	default: // discord
		reqURL = cfg.URL
		body, err = json.Marshal(map[string]string{"content": content})
	}
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("User-Agent", "AmiyaEden/1.0")
	if cfg.Type == "onebot" && cfg.OBToken != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.OBToken)
	}

	resp, err := s.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook 返回错误状态码: %d", resp.StatusCode)
	}
	return nil
}
