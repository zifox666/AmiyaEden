package model

import "testing"

func TestHasNonGuestRole(t *testing.T) {
	tests := []struct {
		name  string
		roles []string
		want  bool
	}{
		{name: "empty", roles: nil, want: false},
		{name: "guest only", roles: []string{RoleGuest}, want: false},
		{name: "user", roles: []string{RoleUser}, want: true},
		{name: "srp and guest", roles: []string{RoleGuest, RoleSRP}, want: true},
		{name: "super admin", roles: []string{RoleSuperAdmin}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasNonGuestRole(tt.roles); got != tt.want {
				t.Fatalf("HasNonGuestRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizeRoleCodes(t *testing.T) {
	tests := []struct {
		name     string
		roles    []string
		fallback string
		want     []string
	}{
		{name: "keep active roles", roles: []string{RoleAdmin, RoleGuest}, fallback: RoleGuest, want: []string{RoleAdmin, RoleGuest}},
		{name: "fallback to legacy role", roles: nil, fallback: RoleUser, want: []string{RoleUser}},
		{name: "fallback to guest", roles: nil, fallback: "", want: []string{RoleGuest}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeRoleCodes(tt.roles, tt.fallback)
			if len(got) != len(tt.want) {
				t.Fatalf("expected %d roles, got %d (%v)", len(tt.want), len(got), got)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Fatalf("expected role %q at %d, got %q", tt.want[i], i, got[i])
				}
			}
		})
	}
}

func TestSystemRoleSeedsIncludeSeniorFC(t *testing.T) {
	for _, role := range SystemRoleSeeds {
		if role.Code == RoleSeniorFC {
			if role.Name == "" {
				t.Fatal("expected senior_fc seed to have a name")
			}
			if role.Description == "" {
				t.Fatal("expected senior_fc seed to have a description")
			}
			return
		}
	}

	t.Fatal("expected senior_fc to be present in system role seeds")
}
