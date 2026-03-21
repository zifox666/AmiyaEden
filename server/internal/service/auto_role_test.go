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

func TestHasDirectorCorpRole(t *testing.T) {
	tests := []struct {
		name      string
		corpRoles map[string]struct{}
		want      bool
	}{
		{
			name: "exact director role",
			corpRoles: map[string]struct{}{
				"Director": {},
			},
			want: true,
		},
		{
			name: "case insensitive director role",
			corpRoles: map[string]struct{}{
				"director": {},
			},
			want: true,
		},
		{
			name: "other corp roles only",
			corpRoles: map[string]struct{}{
				"Accountant":      {},
				"Station_Manager": {},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasDirectorCorpRole(tt.corpRoles); got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}
