package router

import (
	"amiya-eden/internal/model"
	"testing"
)

func TestSrpManageRolesIncludeAdmin(t *testing.T) {
	if !containsRoleCode(srpManageRoles, model.RoleAdmin) {
		t.Fatal("expected srp manage roles to include admin")
	}
	if !containsRoleCode(srpPayoutRoles, model.RoleAdmin) {
		t.Fatal("expected srp payout roles to include admin")
	}
}

func containsRoleCode(codes []string, target string) bool {
	for _, code := range codes {
		if code == target {
			return true
		}
	}
	return false
}
