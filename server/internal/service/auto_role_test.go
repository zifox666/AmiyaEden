package service

import (
	"amiya-eden/internal/model"
	"testing"
)

func TestHasAllowedPrimaryCharacter(t *testing.T) {
	t.Run("main character in allowed corporation grants baseline access", func(t *testing.T) {
		allowCorpSet := map[int64]struct{}{
			98000001: {},
			98000002: {},
		}
		chars := []model.EveCharacter{
			{CharacterID: 1001, CorporationID: 98000002},
			{CharacterID: 1002, CorporationID: 98000099},
		}
		if !hasAllowedPrimaryCharacter(1001, chars, allowCorpSet) {
			t.Fatal("expected primary character in allowed corporation to grant access")
		}
	})

	t.Run("alt in allowed corporation does not grant access when main is outside", func(t *testing.T) {
		allowCorpSet := map[int64]struct{}{
			98000001: {},
		}
		chars := []model.EveCharacter{
			{CharacterID: 1001, CorporationID: 98000099},
			{CharacterID: 1002, CorporationID: 98000001},
		}
		if hasAllowedPrimaryCharacter(1001, chars, allowCorpSet) {
			t.Fatal("expected only the main character to control baseline access")
		}
	})
}

func TestHasAnyAllowedCharacter(t *testing.T) {
	t.Run("any allowed character promotes a guest-only account", func(t *testing.T) {
		allowCorpSet := map[int64]struct{}{
			98000001: {},
		}
		chars := []model.EveCharacter{
			{CharacterID: 1001, CorporationID: 98000099},
			{CharacterID: 1002, CorporationID: 98000001},
		}
		if !hasAnyAllowedCharacter(chars, allowCorpSet) {
			t.Fatal("expected any allowed character to grant auto-role baseline access")
		}
	})

	t.Run("all characters outside allow list keep guest baseline", func(t *testing.T) {
		allowCorpSet := map[int64]struct{}{
			98000001: {},
		}
		chars := []model.EveCharacter{
			{CharacterID: 1001, CorporationID: 98000099},
			{CharacterID: 1002, CorporationID: 98000098},
		}
		if hasAnyAllowedCharacter(chars, allowCorpSet) {
			t.Fatal("expected no baseline access when every character is outside allow list")
		}
	})
}

func TestShouldAutoPromoteGuestToUser(t *testing.T) {
	allowCorpSet := map[int64]struct{}{
		98000001: {},
	}
	chars := []model.EveCharacter{
		{CharacterID: 1001, CorporationID: 98000099},
		{CharacterID: 1002, CorporationID: 98000001},
	}

	t.Run("guest-only account is promoted when any character is allowed", func(t *testing.T) {
		if !shouldAutoPromoteGuestToUser([]string{model.RoleGuest}, chars, allowCorpSet) {
			t.Fatal("expected guest-only account to be promoted to user")
		}
	})

	t.Run("real-role account is left untouched", func(t *testing.T) {
		if shouldAutoPromoteGuestToUser([]string{model.RoleFC}, chars, allowCorpSet) {
			t.Fatal("expected non-guest account to keep its existing real role baseline")
		}
	})
}

func TestShouldAutoAssignAdminFromDirector(t *testing.T) {
	tests := []struct {
		name          string
		corporationID int64
		corpRole      string
		want          bool
	}{
		{
			name:          "director in required corporation",
			corporationID: model.SystemCorporationID,
			corpRole:      "Director",
			want:          true,
		},
		{
			name:          "case insensitive director role in required corporation",
			corporationID: model.SystemCorporationID,
			corpRole:      "director",
			want:          true,
		},
		{
			name:          "director in non required corporation",
			corporationID: 98000001,
			corpRole:      "Director",
			want:          false,
		},
		{
			name:          "non director role in required corporation",
			corporationID: model.SystemCorporationID,
			corpRole:      "Accountant",
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldAutoAssignAdminFromDirector(tt.corporationID, tt.corpRole); got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}
