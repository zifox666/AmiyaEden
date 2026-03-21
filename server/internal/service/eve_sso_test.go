package service

import (
	"amiya-eden/internal/model"
	"slices"
	"testing"
	"time"
)

func TestBuildLoginScopesIncludesPublicDataRegisteredAndExtraScopes(t *testing.T) {
	scopeMu.Lock()
	original := slices.Clone(registeredScopes)
	registeredScopes = []RegisteredScope{
		{Module: "killmail", Scope: "esi-killmails.read_killmails.v1"},
		{Module: "wallet", Scope: "  esi-wallet.read_character_wallet.v1  "},
		{Module: "empty", Scope: "   "},
	}
	scopeMu.Unlock()
	t.Cleanup(func() {
		scopeMu.Lock()
		registeredScopes = original
		scopeMu.Unlock()
	})

	scopes := buildLoginScopes([]string{
		"esi-location.read_location.v1",
		"esi-wallet.read_character_wallet.v1",
	})

	if !slices.Contains(scopes, "publicData") {
		t.Fatal("expected publicData to be included")
	}
	if !slices.Contains(scopes, "esi-killmails.read_killmails.v1") {
		t.Fatal("expected registered killmail scope to be included")
	}
	if !slices.Contains(scopes, "esi-wallet.read_character_wallet.v1") {
		t.Fatal("expected trimmed wallet scope to be included")
	}
	if !slices.Contains(scopes, "esi-location.read_location.v1") {
		t.Fatal("expected extra scope to be included")
	}

	seen := make(map[string]struct{}, len(scopes))
	for _, scope := range scopes {
		if _, exists := seen[scope]; exists {
			t.Fatalf("duplicate scope %q in result %v", scope, scopes)
		}
		seen[scope] = struct{}{}
	}
}

func TestGetRegisteredScopesReturnsCopy(t *testing.T) {
	scopeMu.Lock()
	original := slices.Clone(registeredScopes)
	registeredScopes = []RegisteredScope{
		{Module: "killmail", Scope: "esi-killmails.read_killmails.v1"},
	}
	scopeMu.Unlock()
	t.Cleanup(func() {
		scopeMu.Lock()
		registeredScopes = original
		scopeMu.Unlock()
	})

	got := GetRegisteredScopes()
	got[0].Scope = "mutated"

	scopeMu.RLock()
	defer scopeMu.RUnlock()
	if registeredScopes[0].Scope != "esi-killmails.read_killmails.v1" {
		t.Fatalf("expected registeredScopes to remain unchanged, got %q", registeredScopes[0].Scope)
	}
}

func TestBuildDefaultSSOUser(t *testing.T) {
	now := time.Date(2026, time.March, 22, 10, 11, 12, 0, time.UTC)
	user := buildDefaultSSOUser("https://example.com/avatar.png", 90000001, "127.0.0.1", now)

	if user.Role != model.RoleGuest {
		t.Fatalf("expected role %q, got %q", model.RoleGuest, user.Role)
	}
	if user.PrimaryCharacterID != 90000001 {
		t.Fatalf("expected primary character 90000001, got %d", user.PrimaryCharacterID)
	}
	if user.Avatar != "https://example.com/avatar.png" {
		t.Fatalf("expected avatar to be copied, got %q", user.Avatar)
	}
	if user.LastLoginIP != "127.0.0.1" {
		t.Fatalf("expected last login ip to be copied, got %q", user.LastLoginIP)
	}
	if user.LastLoginAt == nil || !user.LastLoginAt.Equal(now) {
		t.Fatalf("expected last login time %v, got %+v", now, user.LastLoginAt)
	}
	if user.Status != 1 {
		t.Fatalf("expected status 1, got %d", user.Status)
	}
}
