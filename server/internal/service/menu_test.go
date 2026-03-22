package service

import (
	"amiya-eden/internal/model"
	"testing"
)

func TestFilterMenusBySystemRoleRestrictions(t *testing.T) {
	menus := []model.Menu{
		{Name: "Operation"},
		{Name: "Fleets"},
		{Name: "FleetDetail"},
		{Name: "FleetConfigs"},
		{Name: "CorporationPap"},
	}

	t.Run("user loses fleet management menus but keeps fleet configs", func(t *testing.T) {
		filtered := filterMenusBySystemRoleRestrictions(menus, []string{model.RoleUser})
		got := make(map[string]struct{}, len(filtered))
		for _, menu := range filtered {
			got[menu.Name] = struct{}{}
		}

		if _, ok := got["Fleets"]; ok {
			t.Fatal("expected Fleets to be filtered for user role")
		}
		if _, ok := got["FleetDetail"]; ok {
			t.Fatal("expected FleetDetail to be filtered for user role")
		}
		if _, ok := got["FleetConfigs"]; !ok {
			t.Fatal("expected FleetConfigs to remain visible for user role")
		}
		if _, ok := got["CorporationPap"]; !ok {
			t.Fatal("expected CorporationPap to remain visible for user role")
		}
	})

	t.Run("fc keeps restricted fleet menus", func(t *testing.T) {
		filtered := filterMenusBySystemRoleRestrictions(menus, []string{model.RoleFC})
		got := make(map[string]struct{}, len(filtered))
		for _, menu := range filtered {
			got[menu.Name] = struct{}{}
		}

		for _, name := range []string{"Fleets", "FleetDetail", "FleetConfigs"} {
			if _, ok := got[name]; !ok {
				t.Fatalf("expected %s to remain visible for fc role", name)
			}
		}
	})
}
