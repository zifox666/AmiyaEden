package service

import (
	"amiya-eden/internal/model"
	"testing"
)

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
			1,
			2,
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
			1,
			2,
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
			1,
			2,
			[]string{model.RoleAdmin},
			[]string{model.RoleUser},
			[]string{model.RoleUser, model.RoleFC},
		)
		if err != nil {
			t.Fatalf("expected normal role assignment to pass, got %v", err)
		}
	})

	t.Run("admin can edit own roles", func(t *testing.T) {
		err := validateSetUserRolesPermission(
			7,
			7,
			[]string{model.RoleAdmin},
			[]string{model.RoleAdmin},
			[]string{model.RoleUser, model.RoleFC},
		)
		if err != nil {
			t.Fatalf("expected self role edit to pass, got %v", err)
		}
	})

	t.Run("admin cannot add super admin to self", func(t *testing.T) {
		err := validateSetUserRolesPermission(
			7,
			7,
			[]string{model.RoleAdmin},
			[]string{model.RoleAdmin},
			[]string{model.RoleAdmin, model.RoleSuperAdmin},
		)
		if err == nil {
			t.Fatal("expected self super admin assignment to be blocked")
		}
	})
}

func TestNormalizeAssignedRoleCodes(t *testing.T) {
	t.Run("keeps guest when it is the only role", func(t *testing.T) {
		codes := normalizeAssignedRoleCodes([]string{model.RoleGuest})
		if len(codes) != 1 || codes[0] != model.RoleGuest {
			t.Fatalf("expected guest role code to remain, got %v", codes)
		}
	})

	t.Run("drops guest when a real role is present", func(t *testing.T) {
		codes := normalizeAssignedRoleCodes([]string{model.RoleGuest, model.RoleUser, model.RoleFC})
		if len(codes) != 2 {
			t.Fatalf("expected 2 non-guest codes, got %v", codes)
		}
		if model.ContainsRole(codes, model.RoleGuest) {
			t.Fatalf("expected guest to be dropped, got %v", codes)
		}
	})

	t.Run("deduplicates codes", func(t *testing.T) {
		codes := normalizeAssignedRoleCodes([]string{model.RoleUser, model.RoleUser, model.RoleFC})
		if len(codes) != 2 {
			t.Fatalf("expected deduplicated codes, got %v", codes)
		}
	})
}
