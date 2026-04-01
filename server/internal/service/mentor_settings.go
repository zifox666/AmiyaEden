package service

import (
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"errors"
	"fmt"
)

type MentorSettings struct {
	MaxCharacterSP    int64 `json:"max_character_sp"`
	MaxAccountAgeDays int   `json:"max_account_age_days"`
}

func DefaultMentorSettings() MentorSettings {
	return MentorSettings{
		MaxCharacterSP:    model.SysConfigDefaultMenteeMaxCharacterSP,
		MaxAccountAgeDays: model.SysConfigDefaultMenteeMaxAccountAgeDays,
	}
}

func (s MentorSettings) Validate() error {
	switch {
	case s.MaxCharacterSP <= 0:
		return errors.New("学员人物技能点阈值必须大于 0")
	case s.MaxAccountAgeDays <= 0:
		return errors.New("学员账号注册天数阈值必须大于 0")
	default:
		return nil
	}
}

type MentorSettingsService struct {
	cfgRepo mentorSettingsConfigStore
}

type mentorSettingsConfigStore interface {
	GetInt64(key string, defaultVal int64) int64
	GetInt(key string, defaultVal int) int
	SetMany(items []repository.SysConfigUpsertItem) error
}

func NewMentorSettingsService() *MentorSettingsService {
	return &MentorSettingsService{cfgRepo: repository.NewSysConfigRepository()}
}

func (s *MentorSettingsService) GetSettings() MentorSettings {
	defaults := DefaultMentorSettings()
	return MentorSettings{
		MaxCharacterSP:    s.cfgRepo.GetInt64(model.SysConfigMenteeMaxCharacterSP, defaults.MaxCharacterSP),
		MaxAccountAgeDays: s.cfgRepo.GetInt(model.SysConfigMenteeMaxAccountAgeDays, defaults.MaxAccountAgeDays),
	}
}

func (s *MentorSettingsService) UpdateSettings(cfg MentorSettings) (MentorSettings, error) {
	if err := cfg.Validate(); err != nil {
		return MentorSettings{}, err
	}

	items := []repository.SysConfigUpsertItem{
		{
			Key:   model.SysConfigMenteeMaxCharacterSP,
			Value: fmt.Sprintf("%d", cfg.MaxCharacterSP),
			Desc:  "导师学员资格：人物技能点上限",
		},
		{
			Key:   model.SysConfigMenteeMaxAccountAgeDays,
			Value: fmt.Sprintf("%d", cfg.MaxAccountAgeDays),
			Desc:  "导师学员资格：账号注册天数上限",
		},
	}

	if err := s.cfgRepo.SetMany(items); err != nil {
		return MentorSettings{}, err
	}

	return cfg, nil
}
