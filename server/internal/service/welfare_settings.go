package service

import (
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"errors"
)

type WelfareSettings struct {
	AutoApproveFuxiCoinThreshold int `json:"auto_approve_fuxi_coin_threshold"`
}

func DefaultWelfareSettings() WelfareSettings {
	return WelfareSettings{
		AutoApproveFuxiCoinThreshold: model.SysConfigDefaultWelfareAutoApproveFuxiCoinThreshold,
	}
}

func (s WelfareSettings) Validate() error {
	if s.AutoApproveFuxiCoinThreshold < 0 {
		return errors.New("福利自动审批伏羲币阈值不能小于 0")
	}
	return nil
}

type welfareSettingsConfigStore interface {
	GetInt(key string, defaultVal int) int
	SetMany(items []repository.SysConfigUpsertItem) error
}

type WelfareSettingsService struct {
	cfgRepo welfareSettingsConfigStore
}

func NewWelfareSettingsService() *WelfareSettingsService {
	return &WelfareSettingsService{
		cfgRepo: repository.NewSysConfigRepository(),
	}
}

func (s *WelfareSettingsService) GetSettings() WelfareSettings {
	defaults := DefaultWelfareSettings()
	return WelfareSettings{
		AutoApproveFuxiCoinThreshold: s.cfgRepo.GetInt(
			model.SysConfigWelfareAutoApproveFuxiCoinThreshold,
			defaults.AutoApproveFuxiCoinThreshold,
		),
	}
}

func (s *WelfareSettingsService) UpdateSettings(cfg WelfareSettings) (WelfareSettings, error) {
	if err := cfg.Validate(); err != nil {
		return WelfareSettings{}, err
	}

	items := newSysConfigBatch(1).
		AddInt(
			model.SysConfigWelfareAutoApproveFuxiCoinThreshold,
			cfg.AutoApproveFuxiCoinThreshold,
			"福利自动审批伏羲币阈值",
		).
		Items()

	if err := s.cfgRepo.SetMany(items); err != nil {
		return WelfareSettings{}, err
	}

	return cfg, nil
}
