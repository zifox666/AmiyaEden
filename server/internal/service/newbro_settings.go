package service

import (
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"errors"
	"time"
)

type NewbroSettings struct {
	MaxCharacterSP          int64   `json:"max_character_sp"`
	MultiCharacterSP        int64   `json:"multi_character_sp"`
	MultiCharacterThreshold int     `json:"multi_character_threshold"`
	RefreshIntervalDays     int     `json:"refresh_interval_days"`
	BonusRate               float64 `json:"bonus_rate"`
}

func DefaultNewbroSettings() NewbroSettings {
	return NewbroSettings{
		MaxCharacterSP:          model.SysConfigDefaultNewbroMaxCharacterSP,
		MultiCharacterSP:        model.SysConfigDefaultNewbroMultiCharacterSP,
		MultiCharacterThreshold: model.SysConfigDefaultNewbroMultiCharacterThreshold,
		RefreshIntervalDays:     model.SysConfigDefaultNewbroRefreshIntervalDays,
		BonusRate:               model.SysConfigDefaultNewbroBonusRate,
	}
}

func (s NewbroSettings) Validate() error {
	switch {
	case s.MaxCharacterSP <= 0:
		return errors.New("单人物技能点阈值必须大于 0")
	case s.MultiCharacterSP <= 0:
		return errors.New("多人物技能点阈值必须大于 0")
	case s.MultiCharacterThreshold <= 0:
		return errors.New("多人物计数阈值必须大于 0")
	case s.RefreshIntervalDays <= 0:
		return errors.New("资格快照刷新间隔必须大于 0")
	case s.BonusRate < 0:
		return errors.New("队长奖励比例不能小于 0")
	default:
		return nil
	}
}

func (s NewbroSettings) ToEligibilityRules() NewbroEligibilityRules {
	return NewbroEligibilityRules{
		MaxCharacterSP:          s.MaxCharacterSP,
		MultiCharacterSP:        s.MultiCharacterSP,
		MultiCharacterThreshold: s.MultiCharacterThreshold,
		AttributionLookbackDays: newbroAttributionLookbackDays,
	}
}

func (s NewbroSettings) RefreshInterval() time.Duration {
	return time.Duration(s.RefreshIntervalDays) * 24 * time.Hour
}

type NewbroSettingsService struct {
	cfgRepo newbroSettingsConfigStore
}

type newbroSettingsConfigStore interface {
	GetInt64(key string, defaultVal int64) int64
	GetInt(key string, defaultVal int) int
	GetFloat(key string, defaultVal float64) float64
	SetMany(items []repository.SysConfigUpsertItem) error
}

func NewNewbroSettingsService() *NewbroSettingsService {
	return &NewbroSettingsService{
		cfgRepo: repository.NewSysConfigRepository(),
	}
}

func (s *NewbroSettingsService) GetSettings() NewbroSettings {
	defaults := DefaultNewbroSettings()
	return NewbroSettings{
		MaxCharacterSP:          s.cfgRepo.GetInt64(model.SysConfigNewbroMaxCharacterSP, defaults.MaxCharacterSP),
		MultiCharacterSP:        s.cfgRepo.GetInt64(model.SysConfigNewbroMultiCharacterSP, defaults.MultiCharacterSP),
		MultiCharacterThreshold: s.cfgRepo.GetInt(model.SysConfigNewbroMultiCharacterThreshold, defaults.MultiCharacterThreshold),
		RefreshIntervalDays:     s.cfgRepo.GetInt(model.SysConfigNewbroRefreshIntervalDays, defaults.RefreshIntervalDays),
		BonusRate:               s.cfgRepo.GetFloat(model.SysConfigNewbroBonusRate, defaults.BonusRate),
	}
}

func (s *NewbroSettingsService) UpdateSettings(cfg NewbroSettings) (NewbroSettings, error) {
	if err := cfg.Validate(); err != nil {
		return NewbroSettings{}, err
	}

	items := newSysConfigBatch(5).
		AddInt64(model.SysConfigNewbroMaxCharacterSP, cfg.MaxCharacterSP, "新人资格：单人物技能点阈值").
		AddInt64(model.SysConfigNewbroMultiCharacterSP, cfg.MultiCharacterSP, "新人资格：多人物技能点阈值").
		AddInt(model.SysConfigNewbroMultiCharacterThreshold, cfg.MultiCharacterThreshold, "新人资格：达到多人物阈值的人物数量").
		AddInt(model.SysConfigNewbroRefreshIntervalDays, cfg.RefreshIntervalDays, "新人资格快照刷新间隔（天）").
		AddFloat64(model.SysConfigNewbroBonusRate, cfg.BonusRate, "队长奖励比例（百分比）").
		Items()

	if err := s.cfgRepo.SetMany(items); err != nil {
		return NewbroSettings{}, err
	}

	return cfg, nil
}
