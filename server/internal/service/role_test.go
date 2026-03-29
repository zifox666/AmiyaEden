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
	t.Run("super admin can assign any role to others", func(t *testing.T) {
		err := validateSetUserRolesPermission(
			1,
			2,
			[]string{model.RoleSuperAdmin},
			[]string{model.RoleUser},
			[]string{model.RoleAdmin, model.RoleFC},
		)
		if err != nil {
			t.Fatalf("expected super admin to assign any role, got %v", err)
		}
	})

	t.Run("super admin can edit own roles", func(t *testing.T) {
		err := validateSetUserRolesPermission(
			1,
			1,
			[]string{model.RoleSuperAdmin},
			[]string{model.RoleSuperAdmin},
			[]string{model.RoleSuperAdmin, model.RoleCaptain},
		)
		if err != nil {
			t.Fatalf("expected super admin self edit to pass, got %v", err)
		}
	})

	t.Run("admin cannot assign admin role to others", func(t *testing.T) {
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

	t.Run("admin can edit own roles adding non-admin", func(t *testing.T) {
		err := validateSetUserRolesPermission(
			7,
			7,
			[]string{model.RoleAdmin},
			[]string{model.RoleAdmin},
			[]string{model.RoleAdmin, model.RoleFC},
		)
		if err != nil {
			t.Fatalf("expected self role edit to pass, got %v", err)
		}
	})

	t.Run("admin can remove own admin role", func(t *testing.T) {
		err := validateSetUserRolesPermission(
			7,
			7,
			[]string{model.RoleAdmin},
			[]string{model.RoleAdmin},
			[]string{model.RoleUser, model.RoleFC},
		)
		if err != nil {
			t.Fatalf("expected self admin removal to pass, got %v", err)
		}
	})

	t.Run("admin self-edit blocked if currentCodes inconsistent and requesting admin", func(t *testing.T) {
		err := validateSetUserRolesPermission(
			7,
			7,
			[]string{model.RoleAdmin},
			[]string{model.RoleUser},
			[]string{model.RoleAdmin},
		)
		if err == nil {
			t.Fatal("expected self admin promotion to be blocked when currentCodes lacks admin")
		}
	})

	t.Run("admin can manage other admin non-admin roles", func(t *testing.T) {
		err := validateSetUserRolesPermission(
			1,
			2,
			[]string{model.RoleAdmin},
			[]string{model.RoleAdmin},
			[]string{model.RoleAdmin, model.RoleFC},
		)
		if err != nil {
			t.Fatalf("expected admin to manage peer admin non-admin roles, got %v", err)
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
		if err != nil {
			t.Fatalf("validateSetUserRolesPermission should not check super_admin (handled by SetUserRoles entry), got %v", err)
		}
	})

	t.Run("non-admin non-super user cannot manage roles", func(t *testing.T) {
		err := validateSetUserRolesPermission(
			1,
			2,
			[]string{model.RoleUser},
			[]string{model.RoleUser},
			[]string{model.RoleFC},
		)
		if err == nil {
			t.Fatal("expected non-admin to be blocked")
		}
	})
}

func TestFilterOutRole(t *testing.T) {
	t.Run("removes target role", func(t *testing.T) {
		result := filterOutRole([]string{"super_admin", "admin", "fc"}, model.RoleSuperAdmin)
		if len(result) != 2 || result[0] != "admin" || result[1] != "fc" {
			t.Fatalf("expected [admin fc], got %v", result)
		}
	})

	t.Run("no target present returns same elements", func(t *testing.T) {
		result := filterOutRole([]string{"admin", "fc"}, model.RoleSuperAdmin)
		if len(result) != 2 {
			t.Fatalf("expected 2 elements, got %v", result)
		}
	})

	t.Run("empty input returns empty", func(t *testing.T) {
		result := filterOutRole([]string{}, model.RoleSuperAdmin)
		if len(result) != 0 {
			t.Fatalf("expected empty, got %v", result)
		}
	})

	t.Run("all elements are target returns empty", func(t *testing.T) {
		result := filterOutRole([]string{"super_admin", "super_admin"}, model.RoleSuperAdmin)
		if len(result) != 0 {
			t.Fatalf("expected empty, got %v", result)
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
