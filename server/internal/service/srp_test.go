package service

import (
	"amiya-eden/internal/model"
	"errors"
	"testing"
	"time"
)

func TestSlotCategory(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "high slot with digit", in: "HiSlot0", want: "HiSlot"},
		{name: "med slot with multiple digits", in: "MedSlot12", want: "MedSlot"},
		{name: "already normalized", in: "Cargo", want: "Cargo"},
		{name: "implant", in: "Implant", want: "Implant"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := slotCategory(tt.in); got != tt.want {
				t.Fatalf("slotCategory(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestSlotCategoryNamesContainRequiredLocales(t *testing.T) {
	requiredCategories := []string{"HiSlot", "MedSlot", "LoSlot", "Cargo"}

	for _, category := range requiredCategories {
		names, ok := slotCategoryNames[category]
		if !ok {
			t.Fatalf("missing slotCategoryNames entry for %q", category)
		}
		if names["zh"] == "" {
			t.Fatalf("missing zh name for %q", category)
		}
		if names["en"] == "" {
			t.Fatalf("missing en name for %q", category)
		}
	}
}

func TestCanManualAutoApproveApplication(t *testing.T) {
	tests := []struct {
		name          string
		selectedFleet string
		app           *model.SrpApplication
		fleet         *model.Fleet
		want          bool
	}{
		{
			name:          "eligible submitted linked app on selected auto approve fleet",
			selectedFleet: "fleet-1",
			app:           &model.SrpApplication{ReviewStatus: model.SrpReviewSubmitted, FleetID: strPtr("fleet-1")},
			fleet:         &model.Fleet{ID: "fleet-1", AutoSrpMode: model.FleetAutoSrpAutoApprove, FleetConfigID: uintPtr(5)},
			want:          true,
		},
		{
			name:  "skip when selected fleet id missing",
			app:   &model.SrpApplication{ReviewStatus: model.SrpReviewSubmitted, FleetID: strPtr("fleet-1")},
			fleet: &model.Fleet{ID: "fleet-1", AutoSrpMode: model.FleetAutoSrpAutoApprove, FleetConfigID: uintPtr(5)},
			want:  false,
		},
		{
			name:          "skip when app fleet does not match selected fleet",
			selectedFleet: "fleet-2",
			app:           &model.SrpApplication{ReviewStatus: model.SrpReviewSubmitted, FleetID: strPtr("fleet-1")},
			fleet:         &model.Fleet{ID: "fleet-1", AutoSrpMode: model.FleetAutoSrpAutoApprove, FleetConfigID: uintPtr(5)},
			want:          false,
		},
		{
			name:          "skip when app is not submitted status",
			selectedFleet: "fleet-1",
			app:           &model.SrpApplication{ReviewStatus: model.SrpReviewApproved, FleetID: strPtr("fleet-1")},
			fleet:         &model.Fleet{ID: "fleet-1", AutoSrpMode: model.FleetAutoSrpAutoApprove, FleetConfigID: uintPtr(5)},
			want:          false,
		},
		{
			name:          "skip when fleet id missing on app",
			selectedFleet: "fleet-1",
			app:           &model.SrpApplication{ReviewStatus: model.SrpReviewSubmitted},
			fleet:         &model.Fleet{ID: "fleet-1", AutoSrpMode: model.FleetAutoSrpAutoApprove, FleetConfigID: uintPtr(5)},
			want:          false,
		},
		{
			name:          "skip when fleet mode is not auto approve",
			selectedFleet: "fleet-1",
			app:           &model.SrpApplication{ReviewStatus: model.SrpReviewSubmitted, FleetID: strPtr("fleet-1")},
			fleet:         &model.Fleet{ID: "fleet-1", AutoSrpMode: model.FleetAutoSrpSubmitOnly, FleetConfigID: uintPtr(5)},
			want:          false,
		},
		{
			name:          "skip when fleet config missing",
			selectedFleet: "fleet-1",
			app:           &model.SrpApplication{ReviewStatus: model.SrpReviewSubmitted, FleetID: strPtr("fleet-1")},
			fleet:         &model.Fleet{ID: "fleet-1", AutoSrpMode: model.FleetAutoSrpAutoApprove},
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := canManualAutoApproveApplication(tt.app, tt.fleet, tt.selectedFleet); got != tt.want {
				t.Fatalf("canManualAutoApproveApplication() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApplyAutoApprovalToApplication(t *testing.T) {
	app := &model.SrpApplication{
		RecommendedAmount: 10_000_000,
		FinalAmount:       10_000_000,
		ReviewStatus:      model.SrpReviewSubmitted,
	}
	reviewerID := uint(42)
	reviewedAt := time.Date(2026, time.March, 22, 11, 30, 0, 0, time.UTC)

	applyAutoApprovalToApplication(app, reviewerID, 25_000_000, 12_500_000, reviewedAt)

	if app.RecommendedAmount != 25_000_000 {
		t.Fatalf("recommended_amount = %v, want %v", app.RecommendedAmount, 25_000_000)
	}
	if app.FinalAmount != 12_500_000 {
		t.Fatalf("final_amount = %v, want %v", app.FinalAmount, 12_500_000)
	}
	if app.ReviewStatus != model.SrpReviewApproved {
		t.Fatalf("review_status = %q, want %q", app.ReviewStatus, model.SrpReviewApproved)
	}
	if app.ReviewedBy == nil || *app.ReviewedBy != reviewerID {
		t.Fatalf("reviewed_by = %v, want %d", app.ReviewedBy, reviewerID)
	}
	if app.ReviewedAt == nil || !app.ReviewedAt.Equal(reviewedAt) {
		t.Fatalf("reviewed_at = %v, want %v", app.ReviewedAt, reviewedAt)
	}
	if app.ReviewNote != "补损根据舰队的自动补损设置，已由系统自动批准。" {
		t.Fatalf("review_note = %q, want %q", app.ReviewNote, "补损根据舰队的自动补损设置，已由系统自动批准。")
	}
}

func TestAutoApproveReviewNote(t *testing.T) {
	got := autoApproveReviewNote()
	want := "补损根据舰队的自动补损设置，已由系统自动批准。"
	if got != want {
		t.Fatalf("autoApproveReviewNote() = %q, want %q", got, want)
	}
}

func TestIsDuplicateSrpApplicationError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "postgres duplicate key", err: errors.New("duplicate key value violates unique constraint"), want: true},
		{name: "sqlite UNIQUE constraint", err: errors.New("UNIQUE constraint failed: srp_application.killmail_id"), want: true},
		{name: "mysql Duplicate entry", err: errors.New("Duplicate entry '123' for key 'idx_km_char'"), want: true},
		{name: "unrelated error", err: errors.New("connection refused"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isDuplicateSrpApplicationError(tt.err); got != tt.want {
				t.Fatalf("isDuplicateSrpApplicationError(%q) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func strPtr(value string) *string {
	return &value
}

func uintPtr(value uint) *uint {
	return &value
}
