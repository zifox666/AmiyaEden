package service

import (
	"amiya-eden/internal/model"
	"testing"
	"time"
)

func TestPapImportanceToWalletRate(t *testing.T) {
	rateMap := map[string]float64{
		model.PAPTypeSkirmish: 10,
		model.PAPTypeStratOp:  30,
		model.PAPTypeCTA:      50,
	}

	tests := []struct {
		name       string
		importance string
		want       float64
	}{
		{name: "CTA maps to cta rate", importance: model.FleetImportanceCTA, want: 50},
		{name: "strat_op maps to strat_op rate", importance: model.FleetImportanceStratOp, want: 30},
		{name: "other maps to skirmish rate", importance: model.FleetImportanceOther, want: 10},
		{name: "unknown importance maps to skirmish rate", importance: "unknown", want: 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := papImportanceToWalletRate(tt.importance, rateMap); got != tt.want {
				t.Fatalf("papImportanceToWalletRate(%q) = %v, want %v", tt.importance, got, tt.want)
			}
		})
	}
}

func TestPapImportanceToWalletRateMissingKey(t *testing.T) {
	// When a pap_type is absent from the map (e.g. DB read failure), fall back to 1.
	if got := papImportanceToWalletRate(model.FleetImportanceCTA, map[string]float64{}); got != 1 {
		t.Fatalf("expected fallback rate 1, got %v", got)
	}
}

func TestBuildPapWalletByUser(t *testing.T) {
	entries := []papWalletEntry{
		{UserID: 1, PapCount: 12},
		{UserID: 1, PapCount: 8},
		{UserID: 2, PapCount: 5},
		{UserID: 3, PapCount: 7},
	}

	got := buildPapWalletByUser(entries, 10)

	want := map[uint]float64{
		1: 200,
		2: 50,
		3: 70,
	}

	if len(got) != len(want) {
		t.Fatalf("buildPapWalletByUser() len = %d, want %d", len(got), len(want))
	}
	for uid, wantAmount := range want {
		if got[uid] != wantAmount {
			t.Fatalf("buildPapWalletByUser()[%d] = %v, want %v", uid, got[uid], wantAmount)
		}
	}
}

func TestCalculateFCSalaryAmount(t *testing.T) {
	tests := []struct {
		name                 string
		fcInMembers          bool
		existingSalaryAmount float64
		monthlyCount         int64
		monthlyLimit         int
		currentSalary        float64
		want                 float64
	}{
		{name: "not in members", fcInMembers: false, existingSalaryAmount: 400, monthlyCount: 0, monthlyLimit: 5, currentSalary: 400, want: 0},
		{name: "existing salary stays", fcInMembers: true, existingSalaryAmount: 200, monthlyCount: 5, monthlyLimit: 5, currentSalary: 400, want: 400},
		{name: "under limit", fcInMembers: true, existingSalaryAmount: 0, monthlyCount: 4, monthlyLimit: 5, currentSalary: 400, want: 400},
		{name: "at limit", fcInMembers: true, existingSalaryAmount: 0, monthlyCount: 5, monthlyLimit: 5, currentSalary: 400, want: 0},
		{name: "disabled limit", fcInMembers: true, existingSalaryAmount: 0, monthlyCount: 0, monthlyLimit: 0, currentSalary: 400, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateFCSalaryAmount(tt.fcInMembers, tt.existingSalaryAmount, tt.monthlyCount, tt.monthlyLimit, tt.currentSalary)
			if got != tt.want {
				t.Fatalf("calculateFCSalaryAmount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToPapWalletEntriesFromLogs(t *testing.T) {
	logs := []model.FleetPapLog{
		{UserID: 1, PapCount: 12},
		{UserID: 2, PapCount: 34},
	}

	got := toPapWalletEntriesFromLogs(logs)

	if len(got) != len(logs) {
		t.Fatalf("toPapWalletEntriesFromLogs() len = %d, want %d", len(got), len(logs))
	}
	for i := range logs {
		if got[i].UserID != logs[i].UserID || got[i].PapCount != logs[i].PapCount {
			t.Fatalf("entry %d = %+v, want %+v", i, got[i], logs[i])
		}
	}
}

func TestNormalizeAutoSrpMode(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "submit only", in: model.FleetAutoSrpSubmitOnly, want: model.FleetAutoSrpSubmitOnly},
		{name: "auto approve", in: model.FleetAutoSrpAutoApprove, want: model.FleetAutoSrpAutoApprove},
		{name: "empty defaults disabled", in: "", want: model.FleetAutoSrpDisabled},
		{name: "unknown defaults disabled", in: "surprise", want: model.FleetAutoSrpDisabled},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeAutoSrpMode(tt.in); got != tt.want {
				t.Fatalf("normalizeAutoSrpMode(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestFleetServiceCanManageFleet(t *testing.T) {
	svc := &FleetService{}
	fleet := &model.Fleet{FCUserID: 42}

	tests := []struct {
		name      string
		fleet     *model.Fleet
		userID    uint
		userRoles []string
		want      bool
	}{
		// privileged roles can manage any fleet regardless of ownership
		{name: "admin any fleet", fleet: fleet, userID: 7, userRoles: []string{model.RoleAdmin}, want: true},
		{name: "super admin any fleet", fleet: fleet, userID: 8, userRoles: []string{model.RoleSuperAdmin}, want: true},
		{name: "senior_fc any fleet", fleet: fleet, userID: 9, userRoles: []string{model.RoleSeniorFC}, want: true},
		// fc can only manage own fleet
		{name: "fc own fleet", fleet: fleet, userID: 42, userRoles: []string{model.RoleFC}, want: true},
		{name: "fc other fleet", fleet: fleet, userID: 99, userRoles: []string{model.RoleFC}, want: false},
		// fc with nil fleet (no fleet context, e.g. DeactivateInvite)
		{name: "fc nil fleet", fleet: nil, userID: 99, userRoles: []string{model.RoleFC}, want: true},
		// unprivileged roles are denied
		{name: "user denied", fleet: fleet, userID: 42, userRoles: []string{model.RoleUser}, want: false},
		{name: "other user denied", fleet: fleet, userID: 9, userRoles: []string{model.RoleUser}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := svc.canManageFleet(tt.fleet, tt.userID, tt.userRoles); got != tt.want {
				t.Fatalf("canManageFleet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFleetServiceCanDeleteFleet(t *testing.T) {
	svc := &FleetService{}

	tests := []struct {
		name      string
		userRoles []string
		want      bool
	}{
		{name: "admin", userRoles: []string{model.RoleAdmin}, want: true},
		{name: "super admin", userRoles: []string{model.RoleSuperAdmin}, want: true},
		{name: "fc", userRoles: []string{model.RoleFC}, want: false},
		{name: "user", userRoles: []string{model.RoleUser}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := svc.canDeleteFleet(tt.userRoles); got != tt.want {
				t.Fatalf("canDeleteFleet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizeCharacterNames(t *testing.T) {
	got := normalizeCharacterNames([]string{
		"  PlayerOne  ",
		"",
		"PlayerTwo",
		"PlayerOne",
		"   ",
		"PlayerThree",
	})

	want := []string{"PlayerOne", "PlayerTwo", "PlayerThree"}
	if len(got) != len(want) {
		t.Fatalf("normalizeCharacterNames() len = %d, want %d", len(got), len(want))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("normalizeCharacterNames()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestFleetServiceBuildCorporationPapFilter(t *testing.T) {
	svc := &FleetService{}
	location := time.FixedZone("UTC+8", 8*60*60)
	now := time.Date(2026, time.March, 21, 10, 30, 0, 0, location)

	t.Run("default last month", func(t *testing.T) {
		filter, period, year, err := svc.buildCorporationPapFilter("", 0, now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if period != CorporationPapPeriodLastMonth {
			t.Fatalf("period = %q, want %q", period, CorporationPapPeriodLastMonth)
		}
		if year != nil {
			t.Fatalf("year = %v, want nil", *year)
		}
		if filter.StartAt == nil || filter.StartAt.Format(time.DateOnly) != "2026-02-01" {
			t.Fatalf("start = %v, want 2026-02-01", filter.StartAt)
		}
		if filter.EndAt == nil || filter.EndAt.Format(time.DateOnly) != "2026-03-01" {
			t.Fatalf("end = %v, want 2026-03-01", filter.EndAt)
		}
	})

	t.Run("at year normalizes current year when missing", func(t *testing.T) {
		filter, period, year, err := svc.buildCorporationPapFilter(CorporationPapPeriodAtYear, 0, now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if period != CorporationPapPeriodAtYear {
			t.Fatalf("period = %q, want %q", period, CorporationPapPeriodAtYear)
		}
		if year == nil || *year != 2026 {
			t.Fatalf("year = %v, want 2026", year)
		}
		if filter.StartAt == nil || filter.StartAt.Format(time.DateOnly) != "2026-01-01" {
			t.Fatalf("start = %v, want 2026-01-01", filter.StartAt)
		}
		if filter.EndAt == nil || filter.EndAt.Format(time.DateOnly) != "2027-01-01" {
			t.Fatalf("end = %v, want 2027-01-01", filter.EndAt)
		}
	})

	t.Run("last year alias", func(t *testing.T) {
		_, period, year, err := svc.buildCorporationPapFilter("last_year", 0, now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if period != CorporationPapPeriodAtYear {
			t.Fatalf("period = %q, want %q", period, CorporationPapPeriodAtYear)
		}
		if year == nil || *year != 2025 {
			t.Fatalf("year = %v, want 2025", year)
		}
	})

	t.Run("invalid period", func(t *testing.T) {
		if _, _, _, err := svc.buildCorporationPapFilter("bad_period", 0, now); err == nil {
			t.Fatal("expected invalid period error")
		}
	})
}
