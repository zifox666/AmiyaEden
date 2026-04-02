package model

import (
	"encoding/json"
	"testing"
)

func TestNewUserListItem(t *testing.T) {
	t.Run("omits legacy role field from json", func(t *testing.T) {
		item := NewUserListItem(User{
			BaseModel: BaseModel{ID: 7},
			Nickname:  "Capsuleer",
			Role:      RoleGuest,
		}, []string{RoleSuperAdmin, RoleAdmin}, []UserListCharacter{{
			CharacterID:   9001,
			CharacterName: "Amiya Prime",
			PortraitURL:   "portrait.png",
			TotalSP:       123456,
			TokenInvalid:  true,
		}})

		payload, err := json.Marshal(item)
		if err != nil {
			t.Fatalf("marshal item: %v", err)
		}

		var got map[string]any
		if err := json.Unmarshal(payload, &got); err != nil {
			t.Fatalf("unmarshal payload: %v", err)
		}

		if _, exists := got["role"]; exists {
			t.Fatalf("expected user list payload to omit legacy role field, got %v", got["role"])
		}

		rawRoles, ok := got["roles"].([]any)
		if !ok {
			t.Fatalf("expected roles array in payload, got %#v", got["roles"])
		}
		if len(rawRoles) != 2 || rawRoles[0] != RoleSuperAdmin || rawRoles[1] != RoleAdmin {
			t.Fatalf("unexpected roles payload: %#v", rawRoles)
		}

		rawCharacters, ok := got["characters"].([]any)
		if !ok {
			t.Fatalf("expected characters array in payload, got %#v", got["characters"])
		}
		if len(rawCharacters) != 1 {
			t.Fatalf("expected one character in payload, got %#v", rawCharacters)
		}

		firstCharacter, ok := rawCharacters[0].(map[string]any)
		if !ok {
			t.Fatalf("expected character object, got %#v", rawCharacters[0])
		}
		if firstCharacter["character_id"] != float64(9001) {
			t.Fatalf("unexpected character id payload: %#v", firstCharacter)
		}
		if firstCharacter["total_sp"] != float64(123456) {
			t.Fatalf("unexpected character total_sp payload: %#v", firstCharacter)
		}
		if firstCharacter["token_invalid"] != true {
			t.Fatalf("expected token_invalid to be serialized, got %#v", firstCharacter)
		}
	})

	t.Run("falls back to guest when no active roles exist", func(t *testing.T) {
		item := NewUserListItem(User{}, nil, nil)
		if len(item.Roles) != 1 || item.Roles[0] != RoleGuest {
			t.Fatalf("expected guest fallback, got %#v", item.Roles)
		}
		if item.Characters == nil || len(item.Characters) != 0 {
			t.Fatalf("expected characters to default to an empty slice, got %#v", item.Characters)
		}
	})
}
