package service

import (
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"net/http"
	"testing"
)

type fakeWebhookConfigStore struct {
	setManyCalls int
	setManyItems []repository.SysConfigUpsertItem
	setManyErr   error
}

func (f *fakeWebhookConfigStore) Get(_ string, defaultVal string) (string, error) {
	return defaultVal, nil
}

func (f *fakeWebhookConfigStore) GetBool(_ string, defaultVal bool) bool {
	return defaultVal
}

func (f *fakeWebhookConfigStore) SetMany(items []repository.SysConfigUpsertItem) error {
	f.setManyCalls++
	f.setManyItems = append([]repository.SysConfigUpsertItem(nil), items...)
	return f.setManyErr
}

func TestWebhookSetConfigPersistsSingleBatch(t *testing.T) {
	store := &fakeWebhookConfigStore{}
	svc := &WebhookService{repo: store, http: &http.Client{}}

	err := svc.SetConfig(&WebhookConfig{
		URL:           "https://example.test/webhook",
		Enabled:       true,
		Type:          "discord",
		FleetTemplate: defaultFleetTemplate,
		OBTargetType:  "group",
		OBTargetID:    42,
		OBToken:       "token",
	})
	if err != nil {
		t.Fatalf("expected config update to succeed, got %v", err)
	}
	if store.setManyCalls != 1 {
		t.Fatalf("expected exactly one batch write, got %d", store.setManyCalls)
	}
	if len(store.setManyItems) != 7 {
		t.Fatalf("expected 7 config items, got %d", len(store.setManyItems))
	}

	wantKeys := []string{
		model.SysConfigWebhookURL,
		model.SysConfigWebhookEnabled,
		model.SysConfigWebhookType,
		model.SysConfigWebhookFleetTemplate,
		model.SysConfigWebhookOBTargetType,
		model.SysConfigWebhookOBTargetID,
		model.SysConfigWebhookOBToken,
	}
	for i, want := range wantKeys {
		if store.setManyItems[i].Key != want {
			t.Fatalf("unexpected key at index %d: got %q want %q", i, store.setManyItems[i].Key, want)
		}
	}
}
