package service

import (
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"errors"
	"testing"
)

type fakeMentorSettingsConfigStore struct {
	maxCharacterSP    int64
	maxAccountAgeDays int
	setManyCalls      int
	setManyItems      []repository.SysConfigUpsertItem
	setManyErr        error
}

func (f *fakeMentorSettingsConfigStore) GetInt64(_ string, defaultVal int64) int64 {
	if f.maxCharacterSP == 0 {
		return defaultVal
	}
	return f.maxCharacterSP
}

func (f *fakeMentorSettingsConfigStore) GetInt(_ string, defaultVal int) int {
	if f.maxAccountAgeDays == 0 {
		return defaultVal
	}
	return f.maxAccountAgeDays
}

func (f *fakeMentorSettingsConfigStore) SetMany(items []repository.SysConfigUpsertItem) error {
	f.setManyCalls++
	f.setManyItems = append([]repository.SysConfigUpsertItem(nil), items...)
	return f.setManyErr
}

func TestDefaultMentorSettings(t *testing.T) {
	cfg := DefaultMentorSettings()

	if cfg.MaxCharacterSP != 4_000_000 {
		t.Fatalf("expected MaxCharacterSP 4000000, got %d", cfg.MaxCharacterSP)
	}
	if cfg.MaxAccountAgeDays != 7 {
		t.Fatalf("expected MaxAccountAgeDays 7, got %d", cfg.MaxAccountAgeDays)
	}
}

func TestMentorSettingsGetSettingsUsesConfigOverrides(t *testing.T) {
	store := &fakeMentorSettingsConfigStore{maxCharacterSP: 5_500_000, maxAccountAgeDays: 12}
	svc := &MentorSettingsService{cfgRepo: store}

	got := svc.GetSettings()

	if got.MaxCharacterSP != 5_500_000 {
		t.Fatalf("expected MaxCharacterSP 5500000, got %d", got.MaxCharacterSP)
	}
	if got.MaxAccountAgeDays != 12 {
		t.Fatalf("expected MaxAccountAgeDays 12, got %d", got.MaxAccountAgeDays)
	}
}

func TestValidateMentorSettings(t *testing.T) {
	if err := DefaultMentorSettings().Validate(); err != nil {
		t.Fatalf("expected valid defaults, got error %v", err)
	}

	invalidCases := []struct {
		name string
		cfg  MentorSettings
	}{
		{
			name: "max character sp must be positive",
			cfg: MentorSettings{
				MaxCharacterSP:    0,
				MaxAccountAgeDays: 7,
			},
		},
		{
			name: "max account age days must be positive",
			cfg: MentorSettings{
				MaxCharacterSP:    4_000_000,
				MaxAccountAgeDays: 0,
			},
		},
	}

	for _, tc := range invalidCases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.cfg.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestUpdateMentorSettingsPersistsAllKeysInSingleBatch(t *testing.T) {
	store := &fakeMentorSettingsConfigStore{}
	svc := &MentorSettingsService{cfgRepo: store}
	cfg := MentorSettings{
		MaxCharacterSP:    6_000_000,
		MaxAccountAgeDays: 10,
	}

	updated, err := svc.UpdateSettings(cfg)
	if err != nil {
		t.Fatalf("expected update to succeed, got %v", err)
	}
	if updated != cfg {
		t.Fatalf("expected updated settings %v, got %v", cfg, updated)
	}
	if store.setManyCalls != 1 {
		t.Fatalf("expected exactly one batch write, got %d", store.setManyCalls)
	}
	if len(store.setManyItems) != 2 {
		t.Fatalf("expected 2 settings entries, got %d", len(store.setManyItems))
	}

	gotKeys := []string{store.setManyItems[0].Key, store.setManyItems[1].Key}
	wantKeys := []string{model.SysConfigMenteeMaxCharacterSP, model.SysConfigMenteeMaxAccountAgeDays}
	for i := range wantKeys {
		if gotKeys[i] != wantKeys[i] {
			t.Fatalf("unexpected key at index %d: got %q want %q", i, gotKeys[i], wantKeys[i])
		}
	}
}

func TestUpdateMentorSettingsReturnsBatchWriteError(t *testing.T) {
	store := &fakeMentorSettingsConfigStore{setManyErr: errors.New("write failed")}
	svc := &MentorSettingsService{cfgRepo: store}

	_, err := svc.UpdateSettings(DefaultMentorSettings())
	if err == nil {
		t.Fatal("expected batch write error")
	}
	if store.setManyCalls != 1 {
		t.Fatalf("expected one batch write attempt, got %d", store.setManyCalls)
	}
}
