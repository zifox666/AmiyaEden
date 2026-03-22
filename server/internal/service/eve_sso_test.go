package service

import (
	"amiya-eden/internal/model"
	"amiya-eden/pkg/eve/esi"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"
)

func newTestEveSSOService(baseURL string) *EveSSOService {
	return &EveSSOService{
		esiClient: esi.NewClientWithConfig(baseURL, ""),
	}
}

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
	user := buildDefaultSSOUser("https://example.com/avatar.png", 90000001, "127.0.0.1", now, model.RoleUser)

	if user.Role != model.RoleUser {
		t.Fatalf("expected role %q, got %q", model.RoleUser, user.Role)
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

func TestBuildDefaultSSOUserFallsBackToGuestRole(t *testing.T) {
	now := time.Date(2026, time.March, 22, 10, 11, 12, 0, time.UTC)
	user := buildDefaultSSOUser("https://example.com/avatar.png", 90000001, "127.0.0.1", now, "")

	if user.Role != model.RoleGuest {
		t.Fatalf("expected fallback role %q, got %q", model.RoleGuest, user.Role)
	}
}

func TestResolveInitialSSORole(t *testing.T) {
	t.Run("allowed corporation becomes user", func(t *testing.T) {
		role := resolveInitialSSORole(98185110, []int64{98185110, 12345})
		if role != model.RoleUser {
			t.Fatalf("expected role %q, got %q", model.RoleUser, role)
		}
	})

	t.Run("non allowed corporation stays guest", func(t *testing.T) {
		role := resolveInitialSSORole(55555, []int64{98185110, 12345})
		if role != model.RoleGuest {
			t.Fatalf("expected role %q, got %q", model.RoleGuest, role)
		}
	})

	t.Run("empty allow list stays guest", func(t *testing.T) {
		role := resolveInitialSSORole(98185110, nil)
		if role != model.RoleGuest {
			t.Fatalf("expected role %q, got %q", model.RoleGuest, role)
		}
	})
}

func TestFetchCharacterAffiliationRejectsOversizedResponse(t *testing.T) {
	const maxAffiliationResponseBytes = 1 << 20

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			t.Fatalf("expected POST request, got %s", req.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, strings.Repeat("x", maxAffiliationResponseBytes+1))
	}))
	t.Cleanup(server.Close)

	svc := newTestEveSSOService(server.URL)

	_, err := svc.fetchCharacterAffiliation(context.Background(), 90000001)
	if err == nil {
		t.Fatal("expected oversized affiliation response error")
	}
	if !strings.Contains(err.Error(), "response exceeds") {
		t.Fatalf("expected oversize error, got %v", err)
	}
}
