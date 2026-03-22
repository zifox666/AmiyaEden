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
		}, []string{RoleSuperAdmin, RoleAdmin})

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
	})

	t.Run("falls back to guest when no active roles exist", func(t *testing.T) {
		item := NewUserListItem(User{}, nil)
		if len(item.Roles) != 1 || item.Roles[0] != RoleGuest {
			t.Fatalf("expected guest fallback, got %#v", item.Roles)
		}
	})
}
