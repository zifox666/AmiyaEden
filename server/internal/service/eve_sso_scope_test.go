package service

import (
	"amiya-eden/internal/model"
	"testing"
)

func TestBuildLoginScopes_excludesNonRequired(t *testing.T) {
	// Reset global scope state
	scopeMu.Lock()
	orig := registeredScopes
	registeredScopes = nil
	scopeMu.Unlock()
	defer func() {
		scopeMu.Lock()
		registeredScopes = orig
		scopeMu.Unlock()
	}()

	RegisterScope("test", "esi-required.v1", "required scope", true)
	RegisterScope("test", "esi-optional.v1", "optional scope", false)

	scopes := buildLoginScopes(nil)

	scopeSet := make(map[string]bool, len(scopes))
	for _, s := range scopes {
		scopeSet[s] = true
	}

	if !scopeSet["esi-required.v1"] {
		t.Error("expected required scope to be included")
	}
	if scopeSet["esi-optional.v1"] {
		t.Error("expected optional scope to be excluded")
	}
	if !scopeSet["publicData"] {
		t.Error("expected publicData to always be included")
	}
}

func TestBuildLoginScopes_extraScopesOverrideOptional(t *testing.T) {
	scopeMu.Lock()
	orig := registeredScopes
	registeredScopes = nil
	scopeMu.Unlock()
	defer func() {
		scopeMu.Lock()
		registeredScopes = orig
		scopeMu.Unlock()
	}()

	RegisterScope("test", "esi-optional.v1", "optional scope", false)

	scopes := buildLoginScopes([]string{"esi-optional.v1"})

	scopeSet := make(map[string]bool, len(scopes))
	for _, s := range scopes {
		scopeSet[s] = true
	}

	if !scopeSet["esi-optional.v1"] {
		t.Error("expected optional scope to be included when passed as extra")
	}
}

func TestValidateExtraScopesRejectsCorpKillmailScopeForNonAdmin(t *testing.T) {
	err := validateExtraScopes([]string{"esi-killmails.read_corporation_killmails.v1"}, []string{model.RoleSRP})
	if err == nil {
		t.Fatal("expected non-admin user to be rejected for corporation killmail scope")
	}
}

func TestValidateExtraScopesRejectsCorpKillmailScopeForPublicLogin(t *testing.T) {
	err := validateExtraScopes([]string{"esi-killmails.read_corporation_killmails.v1"}, nil)
	if err == nil {
		t.Fatal("expected public login request to be rejected for corporation killmail scope")
	}
}

func TestValidateExtraScopesAllowsCorpKillmailScopeForAdmins(t *testing.T) {
	for _, roles := range [][]string{{model.RoleAdmin}, {model.RoleSuperAdmin}} {
		if err := validateExtraScopes([]string{"esi-killmails.read_corporation_killmails.v1"}, roles); err != nil {
			t.Fatalf("expected roles %v to be allowed, got %v", roles, err)
		}
	}
}

func TestValidateExtraScopesAllowsRegularScopesForLoggedInUsers(t *testing.T) {
	err := validateExtraScopes([]string{"esi-location.read_location.v1"}, []string{model.RoleUser})
	if err != nil {
		t.Fatalf("expected regular scope to be allowed, got %v", err)
	}
}
