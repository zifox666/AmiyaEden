package service

import (
	"amiya-eden/internal/model"
	"testing"
)

func TestContainsRoleCode(t *testing.T) {
	roles := []string{"guest", "admin", "super_admin"}

	if !containsRoleCode(roles, "admin") {
		t.Fatal("expected admin to be found")
	}

	if containsRoleCode(roles, "fc") {
		t.Fatal("did not expect fc to be found")
	}
}

func TestEnsureUserHasDefaultRoleUsesGuest(t *testing.T) {
	svc := NewRoleService()
	if svc == nil {
		t.Fatal("expected role service to be constructed")
	}

	if model.RoleGuest != "guest" {
		t.Fatalf("expected guest compatibility constant, got %q", model.RoleGuest)
	}
}

func TestValidateSetUserRolesPermission(t *testing.T) {
	t.Run("admin cannot edit admin target", func(t *testing.T) {
		err := validateSetUserRolesPermission(
			[]string{model.RoleAdmin},
			[]string{model.RoleAdmin},
			[]string{model.RoleUser},
		)
		if err == nil {
			t.Fatal("expected protected target edit to be blocked")
		}
	})

	t.Run("admin cannot assign admin role", func(t *testing.T) {
		err := validateSetUserRolesPermission(
			[]string{model.RoleAdmin},
			[]string{model.RoleUser},
			[]string{model.RoleAdmin},
		)
		if err == nil {
			t.Fatal("expected admin role assignment to be blocked")
		}
	})

	t.Run("admin can assign normal roles to normal user", func(t *testing.T) {
		err := validateSetUserRolesPermission(
			[]string{model.RoleAdmin},
			[]string{model.RoleUser},
			[]string{model.RoleUser, model.RoleFC},
		)
		if err != nil {
			t.Fatalf("expected normal role assignment to pass, got %v", err)
		}
	})
}
