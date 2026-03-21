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
