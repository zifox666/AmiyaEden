package service

import (
	"amiya-eden/internal/model"
	"testing"
	"time"
)

func TestInitialWelfareApplicationStatusIsRequested(t *testing.T) {
	got := initialWelfareApplicationRequestedStatus()

	if got != model.WelfareAppStatusRequested {
		t.Fatalf("initialWelfareApplicationRequestedStatus() = %q, want %q", got, model.WelfareAppStatusRequested)
	}
}

func TestValidateReviewTransition(t *testing.T) {
	tests := []struct {
		name          string
		currentStatus string
		action        string
		wantStatus    string
		wantErr       bool
	}{
		{
			name:          "deliver from requested succeeds",
			currentStatus: model.WelfareAppStatusRequested,
			action:        "deliver",
			wantStatus:    model.WelfareAppStatusDelivered,
		},
		{
			name:          "reject from requested succeeds",
			currentStatus: model.WelfareAppStatusRequested,
			action:        "reject",
			wantStatus:    model.WelfareAppStatusRejected,
		},
		{
			name:          "deliver from delivered is rejected",
			currentStatus: model.WelfareAppStatusDelivered,
			action:        "deliver",
			wantErr:       true,
		},
		{
			name:          "reject from delivered is rejected",
			currentStatus: model.WelfareAppStatusDelivered,
			action:        "reject",
			wantErr:       true,
		},
		{
			name:          "deliver from rejected is rejected",
			currentStatus: model.WelfareAppStatusRejected,
			action:        "deliver",
			wantErr:       true,
		},
		{
			name:          "reject from rejected is rejected",
			currentStatus: model.WelfareAppStatusRejected,
			action:        "reject",
			wantErr:       true,
		},
		{
			name:          "invalid action is rejected",
			currentStatus: model.WelfareAppStatusRequested,
			action:        "approve",
			wantErr:       true,
		},
		{
			name:          "empty action is rejected",
			currentStatus: model.WelfareAppStatusRequested,
			action:        "",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStatus, err := validateReviewTransition(tt.currentStatus, tt.action)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got status=%q", gotStatus)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotStatus != tt.wantStatus {
				t.Fatalf("got status=%q, want %q", gotStatus, tt.wantStatus)
			}
		})
	}
}

func TestCharacterAgeTooOld(t *testing.T) {
	now := time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC)

	bday := func(y, m, d int) *time.Time {
		t := time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
		return &t
	}

	tests := []struct {
		name     string
		birthday *time.Time
		months   int
		want     bool
	}{
		{
			name:     "nil birthday is not too old",
			birthday: nil,
			months:   6,
			want:     false,
		},
		{
			name:     "character born 3 months ago with 6 month limit is ok",
			birthday: bday(2025, 12, 23),
			months:   6,
			want:     false,
		},
		{
			name:     "character born exactly at limit is not too old",
			birthday: bday(2025, 9, 23),
			months:   6,
			want:     false,
		},
		{
			name:     "character born 7 months ago with 6 month limit is too old",
			birthday: bday(2025, 8, 22),
			months:   6,
			want:     true,
		},
		{
			name:     "character born 2 years ago with 12 month limit is too old",
			birthday: bday(2024, 3, 1),
			months:   12,
			want:     true,
		},
		{
			name:     "character born 11 months ago with 12 month limit is ok",
			birthday: bday(2025, 4, 24),
			months:   12,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := characterAgeTooOld(tt.birthday, tt.months, now)
			if got != tt.want {
				t.Fatalf("characterAgeTooOld() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAnyCharacterTooOld(t *testing.T) {
	now := time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC)

	bday := func(y, m, d int) *time.Time {
		t := time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
		return &t
	}

	young := model.EveCharacter{Birthday: bday(2026, 1, 1)}
	old := model.EveCharacter{Birthday: bday(2024, 1, 1)}
	noBday := model.EveCharacter{Birthday: nil}

	tests := []struct {
		name       string
		characters []model.EveCharacter
		months     int
		want       bool
	}{
		{
			name:       "all young characters pass",
			characters: []model.EveCharacter{young, noBday},
			months:     6,
			want:       false,
		},
		{
			name:       "one old character fails the check",
			characters: []model.EveCharacter{young, old},
			months:     12,
			want:       true,
		},
		{
			name:       "empty character list passes",
			characters: []model.EveCharacter{},
			months:     6,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := anyCharacterTooOld(tt.characters, tt.months, now)
			if got != tt.want {
				t.Fatalf("anyCharacterTooOld() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseImportedWelfareApplicationsSupportsCommaAndTabSeparatedRows(t *testing.T) {
	apps, err := parseImportedWelfareApplications(7, "Alice, 12345\n\nBob\t67890\nCharlie")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(apps) != 3 {
		t.Fatalf("expected 3 parsed applications, got %d", len(apps))
	}

	if apps[0].WelfareID != 7 || apps[0].CharacterName != "Alice" || apps[0].QQ != "12345" {
		t.Fatalf("unexpected first application: %+v", apps[0])
	}
	if apps[0].Status != model.WelfareAppStatusDelivered {
		t.Fatalf("expected imported status %q, got %q", model.WelfareAppStatusDelivered, apps[0].Status)
	}
	if apps[0].UserID != nil {
		t.Fatalf("expected imported user ID to be nil, got %v", apps[0].UserID)
	}

	if apps[1].CharacterName != "Bob" || apps[1].QQ != "67890" {
		t.Fatalf("unexpected second application: %+v", apps[1])
	}

	if apps[2].CharacterName != "Charlie" || apps[2].QQ != "" {
		t.Fatalf("unexpected third application: %+v", apps[2])
	}
}

func TestParseImportedWelfareApplicationsRejectsEmptyResult(t *testing.T) {
	_, err := parseImportedWelfareApplications(7, "\n , \n\t")
	if err == nil {
		t.Fatal("expected error for empty parsed import result")
	}
}
