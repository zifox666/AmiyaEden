package service

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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

func TestConvertSrpAmountToFuxiCoin(t *testing.T) {
	tests := []struct {
		name   string
		isk    float64
		expect float64
	}{
		{name: "rounds down to two decimals", isk: 1_234_567, expect: 1.23},
		{name: "rounds up to two decimals", isk: 1_235_678, expect: 1.24},
		{name: "whole million maps to whole coin", isk: 20_000_000, expect: 20},
		{name: "small amount keeps two decimals", isk: 499_000, expect: 0.50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertSrpAmountToFuxiCoin(tt.isk); got != tt.expect {
				t.Fatalf("convertSrpAmountToFuxiCoin(%v) = %v, want %v", tt.isk, got, tt.expect)
			}
		})
	}
}

func TestBuildSrpPayoutWalletReason(t *testing.T) {
	app := &model.SrpApplication{ID: 42, ShipName: "Guardian"}

	withFleet := buildSrpPayoutWalletReason(app, "钛舰集结")
	if withFleet != "SRP#42 Guardian | 钛舰集结" {
		t.Fatalf("buildSrpPayoutWalletReason() with fleet = %q", withFleet)
	}

	withoutFleet := buildSrpPayoutWalletReason(app, "")
	if withoutFleet != "SRP#42 Guardian" {
		t.Fatalf("buildSrpPayoutWalletReason() without fleet = %q", withoutFleet)
	}
}

func TestPayoutFuxiCoinModeCreditsWalletAndMarksApplicationPaid(t *testing.T) {
	db := newSrpServiceTestDB(t)
	oldDB := global.DB
	global.DB = db
	t.Cleanup(func() { global.DB = oldDB })

	fleetID := "fleet-fuxi"
	if err := db.Create(&model.Fleet{
		ID:              fleetID,
		Title:           "海燕护航",
		StartAt:         time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC),
		EndAt:           time.Date(2026, time.March, 1, 15, 0, 0, 0, time.UTC),
		Importance:      model.FleetImportanceCTA,
		FCUserID:        7,
		FCCharacterID:   90000001,
		FCCharacterName: "FC One",
		AutoSrpMode:     model.FleetAutoSrpDisabled,
	}).Error; err != nil {
		t.Fatalf("create fleet: %v", err)
	}

	app := &model.SrpApplication{
		UserID:            101,
		CharacterID:       90001001,
		CharacterName:     "Pilot One",
		KillmailID:        880001,
		FleetID:           &fleetID,
		ShipTypeID:        22436,
		ShipName:          "Guardian",
		SolarSystemID:     30000142,
		KillmailTime:      time.Date(2026, time.March, 2, 1, 0, 0, 0, time.UTC),
		RecommendedAmount: 2_350_000,
		FinalAmount:       2_350_000,
		ReviewStatus:      model.SrpReviewApproved,
		PayoutStatus:      model.SrpPayoutNotPaid,
	}
	if err := db.Create(app).Error; err != nil {
		t.Fatalf("create app: %v", err)
	}

	svc := newSrpServiceForTests()
	mailAttempted := false
	svc.payoutMailSender = func(ctx context.Context, payerID uint, apps []*model.SrpApplication) error {
		mailAttempted = true
		if payerID != 77 {
			t.Fatalf("payerID = %d, want 77", payerID)
		}
		if len(apps) != 1 || apps[0].ID != app.ID {
			t.Fatalf("unexpected apps payload: %#v", apps)
		}
		return errors.New("mail failed")
	}
	paidApp, err := svc.Payout(77, app.ID, &SrpPayoutRequest{Mode: SrpPayoutModeFuxiCoin})
	if err != nil {
		t.Fatalf("Payout() error = %v", err)
	}
	if !mailAttempted {
		t.Fatal("expected payout mail attempt after successful fuxi payout")
	}

	if paidApp.PayoutStatus != model.SrpPayoutPaid {
		t.Fatalf("payout_status = %q, want %q", paidApp.PayoutStatus, model.SrpPayoutPaid)
	}
	if paidApp.PaidBy == nil || *paidApp.PaidBy != 77 {
		t.Fatalf("paid_by = %v, want 77", paidApp.PaidBy)
	}

	var wallet model.SystemWallet
	if err := db.Where("user_id = ?", app.UserID).First(&wallet).Error; err != nil {
		t.Fatalf("load wallet: %v", err)
	}
	if wallet.Balance != 2.35 {
		t.Fatalf("wallet balance = %v, want 2.35", wallet.Balance)
	}

	var txs []model.WalletTransaction
	if err := db.Order("id ASC").Find(&txs).Error; err != nil {
		t.Fatalf("load wallet transactions: %v", err)
	}
	if len(txs) != 1 {
		t.Fatalf("wallet transaction count = %d, want 1", len(txs))
	}
	tx := txs[0]
	if tx.Amount != 2.35 {
		t.Fatalf("wallet tx amount = %v, want 2.35", tx.Amount)
	}
	if tx.RefType != model.WalletRefSrpPayout {
		t.Fatalf("wallet tx ref_type = %q, want %q", tx.RefType, model.WalletRefSrpPayout)
	}
	if tx.RefID != fmt.Sprintf("srp:%d", app.ID) {
		t.Fatalf("wallet tx ref_id = %q, want %q", tx.RefID, fmt.Sprintf("srp:%d", app.ID))
	}
	wantReason := fmt.Sprintf("SRP#%d Guardian | 海燕护航", app.ID)
	if tx.Reason != wantReason {
		t.Fatalf("wallet tx reason = %q, want %q", tx.Reason, wantReason)
	}
}

func TestPayoutManualTransferIgnoresPayoutMailErrors(t *testing.T) {
	db := newSrpServiceTestDB(t)
	oldDB := global.DB
	global.DB = db
	t.Cleanup(func() { global.DB = oldDB })

	app := &model.SrpApplication{
		UserID:            101,
		CharacterID:       90001001,
		CharacterName:     "Pilot One",
		KillmailID:        880011,
		ShipTypeID:        22436,
		ShipName:          "Guardian",
		SolarSystemID:     30000142,
		KillmailTime:      time.Date(2026, time.March, 2, 1, 0, 0, 0, time.UTC),
		RecommendedAmount: 2_350_000,
		FinalAmount:       2_350_000,
		ReviewStatus:      model.SrpReviewApproved,
		PayoutStatus:      model.SrpPayoutNotPaid,
	}
	if err := db.Create(app).Error; err != nil {
		t.Fatalf("create app: %v", err)
	}

	svc := newSrpServiceForTests()
	mailAttempted := false
	svc.payoutMailSender = func(ctx context.Context, payerID uint, apps []*model.SrpApplication) error {
		mailAttempted = true
		if payerID != 77 {
			t.Fatalf("payerID = %d, want 77", payerID)
		}
		if len(apps) != 1 || apps[0].ID != app.ID {
			t.Fatalf("unexpected apps payload: %#v", apps)
		}
		return errors.New("mail failed")
	}

	paidApp, err := svc.Payout(77, app.ID, &SrpPayoutRequest{Mode: SrpPayoutModeManualTransfer})
	if err != nil {
		t.Fatalf("Payout() error = %v", err)
	}
	if !mailAttempted {
		t.Fatal("expected payout mail attempt after successful payout")
	}
	if paidApp.PayoutStatus != model.SrpPayoutPaid {
		t.Fatalf("payout_status = %q, want %q", paidApp.PayoutStatus, model.SrpPayoutPaid)
	}

	var stored model.SrpApplication
	if err := db.First(&stored, app.ID).Error; err != nil {
		t.Fatalf("reload app: %v", err)
	}
	if stored.PayoutStatus != model.SrpPayoutPaid {
		t.Fatalf("stored payout_status = %q, want %q", stored.PayoutStatus, model.SrpPayoutPaid)
	}
	if stored.PaidBy == nil || *stored.PaidBy != 77 {
		t.Fatalf("stored paid_by = %v, want 77", stored.PaidBy)
	}
}

func TestBatchPayoutAsFuxiCoinCreditsApprovedRequestsAcrossUsers(t *testing.T) {
	db := newSrpServiceTestDB(t)
	oldDB := global.DB
	global.DB = db
	t.Cleanup(func() { global.DB = oldDB })

	fleetID := "fleet-batch"
	if err := db.Create(&model.Fleet{
		ID:              fleetID,
		Title:           "远征夜巡",
		StartAt:         time.Date(2026, time.March, 10, 12, 0, 0, 0, time.UTC),
		EndAt:           time.Date(2026, time.March, 10, 15, 0, 0, 0, time.UTC),
		Importance:      model.FleetImportanceCTA,
		FCUserID:        9,
		FCCharacterID:   90000002,
		FCCharacterName: "FC Two",
		AutoSrpMode:     model.FleetAutoSrpDisabled,
	}).Error; err != nil {
		t.Fatalf("create fleet: %v", err)
	}

	apps := []*model.SrpApplication{
		{
			UserID:            201,
			CharacterID:       90002001,
			CharacterName:     "Pilot A",
			KillmailID:        880101,
			FleetID:           &fleetID,
			ShipTypeID:        111,
			ShipName:          "Scimitar",
			SolarSystemID:     30000142,
			KillmailTime:      time.Date(2026, time.March, 10, 12, 30, 0, 0, time.UTC),
			RecommendedAmount: 1_250_000,
			FinalAmount:       1_250_000,
			ReviewStatus:      model.SrpReviewApproved,
			PayoutStatus:      model.SrpPayoutNotPaid,
		},
		{
			UserID:            202,
			CharacterID:       90002002,
			CharacterName:     "Pilot B",
			KillmailID:        880102,
			ShipTypeID:        222,
			ShipName:          "Basilisk",
			SolarSystemID:     30000142,
			KillmailTime:      time.Date(2026, time.March, 10, 12, 40, 0, 0, time.UTC),
			RecommendedAmount: 2_500_000,
			FinalAmount:       2_500_000,
			ReviewStatus:      model.SrpReviewApproved,
			PayoutStatus:      model.SrpPayoutNotPaid,
		},
		{
			UserID:            201,
			CharacterID:       90002003,
			CharacterName:     "Pilot C",
			KillmailID:        880103,
			ShipTypeID:        333,
			ShipName:          "Oneiros",
			SolarSystemID:     30000142,
			KillmailTime:      time.Date(2026, time.March, 10, 12, 50, 0, 0, time.UTC),
			RecommendedAmount: 500_000,
			FinalAmount:       500_000,
			ReviewStatus:      model.SrpReviewApproved,
			PayoutStatus:      model.SrpPayoutNotPaid,
		},
		{
			UserID:            203,
			CharacterID:       90002004,
			CharacterName:     "Pilot D",
			KillmailID:        880104,
			ShipTypeID:        444,
			ShipName:          "Guardian",
			SolarSystemID:     30000142,
			KillmailTime:      time.Date(2026, time.March, 10, 13, 0, 0, 0, time.UTC),
			RecommendedAmount: 3_000_000,
			FinalAmount:       3_000_000,
			ReviewStatus:      model.SrpReviewSubmitted,
			PayoutStatus:      model.SrpPayoutNotPaid,
		},
	}
	for _, app := range apps {
		if err := db.Create(app).Error; err != nil {
			t.Fatalf("create app %d: %v", app.KillmailID, err)
		}
	}

	svc := newSrpServiceForTests()
	mailAttempted := false
	svc.payoutMailSender = func(ctx context.Context, payerID uint, selectedApps []*model.SrpApplication) error {
		mailAttempted = true
		if payerID != 88 {
			t.Fatalf("payerID = %d, want 88", payerID)
		}
		if len(selectedApps) != 3 {
			t.Fatalf("mail app count = %d, want 3", len(selectedApps))
		}
		return errors.New("mail failed")
	}
	summary, err := svc.BatchPayoutAsFuxiCoin(88)
	if err != nil {
		t.Fatalf("BatchPayoutAsFuxiCoin() error = %v", err)
	}
	if !mailAttempted {
		t.Fatal("expected payout mail attempt after successful fuxi batch payout")
	}

	if summary.ApplicationCount != 3 {
		t.Fatalf("application_count = %d, want 3", summary.ApplicationCount)
	}
	if summary.UserCount != 2 {
		t.Fatalf("user_count = %d, want 2", summary.UserCount)
	}
	if summary.TotalFuxiCoin != 4.25 {
		t.Fatalf("total_fuxi_coin = %v, want 4.25", summary.TotalFuxiCoin)
	}

	var paidCount int64
	if err := db.Model(&model.SrpApplication{}).
		Where("payout_status = ?", model.SrpPayoutPaid).
		Count(&paidCount).Error; err != nil {
		t.Fatalf("count paid apps: %v", err)
	}
	if paidCount != 3 {
		t.Fatalf("paid app count = %d, want 3", paidCount)
	}

	var wallet201 model.SystemWallet
	if err := db.Where("user_id = ?", 201).First(&wallet201).Error; err != nil {
		t.Fatalf("load wallet 201: %v", err)
	}
	if wallet201.Balance != 1.75 {
		t.Fatalf("wallet 201 balance = %v, want 1.75", wallet201.Balance)
	}

	var wallet202 model.SystemWallet
	if err := db.Where("user_id = ?", 202).First(&wallet202).Error; err != nil {
		t.Fatalf("load wallet 202: %v", err)
	}
	if wallet202.Balance != 2.50 {
		t.Fatalf("wallet 202 balance = %v, want 2.50", wallet202.Balance)
	}

	var txCount int64
	if err := db.Model(&model.WalletTransaction{}).Count(&txCount).Error; err != nil {
		t.Fatalf("count wallet txs: %v", err)
	}
	if txCount != 3 {
		t.Fatalf("wallet transaction count = %d, want 3", txCount)
	}
}

func TestBatchPayoutByUserIgnoresPayoutMailErrors(t *testing.T) {
	db := newSrpServiceTestDB(t)
	oldDB := global.DB
	global.DB = db
	t.Cleanup(func() { global.DB = oldDB })

	if err := db.Create(&model.User{
		BaseModel:          model.BaseModel{ID: 201},
		Nickname:           "Pilot A",
		PrimaryCharacterID: 90002001,
	}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := db.Create(&model.EveCharacter{
		CharacterID:   90002001,
		CharacterName: "Pilot A Main",
		UserID:        201,
		TokenExpiry:   time.Now().Add(time.Hour),
	}).Error; err != nil {
		t.Fatalf("create character: %v", err)
	}

	apps := []*model.SrpApplication{
		{
			UserID:            201,
			CharacterID:       90002011,
			CharacterName:     "Pilot A",
			KillmailID:        880301,
			ShipTypeID:        111,
			ShipName:          "Scimitar",
			SolarSystemID:     30000142,
			KillmailTime:      time.Date(2026, time.March, 10, 12, 30, 0, 0, time.UTC),
			RecommendedAmount: 1_250_000,
			FinalAmount:       1_250_000,
			ReviewStatus:      model.SrpReviewApproved,
			PayoutStatus:      model.SrpPayoutNotPaid,
		},
		{
			UserID:            201,
			CharacterID:       90002012,
			CharacterName:     "Pilot A Alt",
			KillmailID:        880302,
			ShipTypeID:        222,
			ShipName:          "Basilisk",
			SolarSystemID:     30000142,
			KillmailTime:      time.Date(2026, time.March, 10, 12, 40, 0, 0, time.UTC),
			RecommendedAmount: 750_000,
			FinalAmount:       750_000,
			ReviewStatus:      model.SrpReviewApproved,
			PayoutStatus:      model.SrpPayoutNotPaid,
		},
		{
			UserID:            201,
			CharacterID:       90002013,
			CharacterName:     "Pilot A Pending",
			KillmailID:        880303,
			ShipTypeID:        333,
			ShipName:          "Oneiros",
			SolarSystemID:     30000142,
			KillmailTime:      time.Date(2026, time.March, 10, 12, 50, 0, 0, time.UTC),
			RecommendedAmount: 500_000,
			FinalAmount:       500_000,
			ReviewStatus:      model.SrpReviewSubmitted,
			PayoutStatus:      model.SrpPayoutNotPaid,
		},
		{
			UserID:            202,
			CharacterID:       90002014,
			CharacterName:     "Pilot B",
			KillmailID:        880304,
			ShipTypeID:        444,
			ShipName:          "Guardian",
			SolarSystemID:     30000142,
			KillmailTime:      time.Date(2026, time.March, 10, 13, 0, 0, 0, time.UTC),
			RecommendedAmount: 900_000,
			FinalAmount:       900_000,
			ReviewStatus:      model.SrpReviewApproved,
			PayoutStatus:      model.SrpPayoutNotPaid,
		},
	}
	for _, app := range apps {
		if err := db.Create(app).Error; err != nil {
			t.Fatalf("create app %d: %v", app.KillmailID, err)
		}
	}

	svc := newSrpServiceForTests()
	mailAttempted := false
	svc.payoutMailSender = func(ctx context.Context, payerID uint, selectedApps []*model.SrpApplication) error {
		mailAttempted = true
		if payerID != 88 {
			t.Fatalf("payerID = %d, want 88", payerID)
		}
		if len(selectedApps) != 2 {
			t.Fatalf("mail app count = %d, want 2", len(selectedApps))
		}
		return errors.New("mail failed")
	}

	summary, err := svc.BatchPayoutByUser(88, 201)
	if err != nil {
		t.Fatalf("BatchPayoutByUser() error = %v", err)
	}
	if !mailAttempted {
		t.Fatal("expected payout mail attempt after successful batch payout")
	}
	if summary.UserID != 201 {
		t.Fatalf("summary user_id = %d, want 201", summary.UserID)
	}
	if summary.ApplicationCount != 2 {
		t.Fatalf("summary application_count = %d, want 2", summary.ApplicationCount)
	}
	if summary.TotalAmount != 2_000_000 {
		t.Fatalf("summary total_amount = %v, want 2000000", summary.TotalAmount)
	}
	if summary.MainCharacterName != "Pilot A Main" {
		t.Fatalf("summary main_character_name = %q, want %q", summary.MainCharacterName, "Pilot A Main")
	}

	var paidCount int64
	if err := db.Model(&model.SrpApplication{}).
		Where("user_id = ? AND payout_status = ?", 201, model.SrpPayoutPaid).
		Count(&paidCount).Error; err != nil {
		t.Fatalf("count paid apps: %v", err)
	}
	if paidCount != 2 {
		t.Fatalf("paid app count = %d, want 2", paidCount)
	}

	var otherUserPaidCount int64
	if err := db.Model(&model.SrpApplication{}).
		Where("user_id = ? AND payout_status = ?", 202, model.SrpPayoutPaid).
		Count(&otherUserPaidCount).Error; err != nil {
		t.Fatalf("count other user paid apps: %v", err)
	}
	if otherUserPaidCount != 0 {
		t.Fatalf("other user paid app count = %d, want 0", otherUserPaidCount)
	}
}

func TestPayoutAsFuxiCoinStoresPayerAsWalletOperator(t *testing.T) {
	db := newSrpServiceTestDB(t)
	originalDB := global.DB
	global.DB = db
	defer func() { global.DB = originalDB }()

	app := &model.SrpApplication{
		UserID:            201,
		CharacterID:       90002001,
		CharacterName:     "Pilot A",
		KillmailID:        880201,
		ShipTypeID:        111,
		ShipName:          "Scimitar",
		SolarSystemID:     30000142,
		KillmailTime:      time.Date(2026, time.March, 10, 12, 30, 0, 0, time.UTC),
		RecommendedAmount: 1_250_000,
		FinalAmount:       1_250_000,
		ReviewStatus:      model.SrpReviewApproved,
		PayoutStatus:      model.SrpPayoutNotPaid,
	}
	if err := db.Create(app).Error; err != nil {
		t.Fatalf("create app: %v", err)
	}

	svc := newSrpServiceForTests()
	if _, err := svc.Payout(77, app.ID, &SrpPayoutRequest{Mode: SrpPayoutModeFuxiCoin}); err != nil {
		t.Fatalf("Payout() error = %v", err)
	}

	var txs []model.WalletTransaction
	if err := db.Order("id ASC").Find(&txs).Error; err != nil {
		t.Fatalf("load wallet transactions: %v", err)
	}
	if len(txs) != 1 {
		t.Fatalf("wallet transaction count = %d, want 1", len(txs))
	}
	if txs[0].OperatorID != 77 {
		t.Fatalf("wallet transaction operator_id = %d, want 77", txs[0].OperatorID)
	}
}

func newSrpServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:srp_service_test_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.User{},
		&model.EveCharacter{},
		&model.SystemWallet{},
		&model.WalletTransaction{},
		&model.SrpApplication{},
		&model.Fleet{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}

func newSrpServiceForTests() *SrpService {
	return &SrpService{
		repo:      repository.NewSrpRepository(),
		fleetRepo: repository.NewFleetRepository(),
		charRepo:  repository.NewEveCharacterRepository(),
		userRepo:  repository.NewUserRepository(),
		sdeRepo:   repository.NewSdeRepository(),
		walletSvc: NewSysWalletService(),
	}
}

func strPtr(value string) *string {
	return &value
}

func uintPtr(value uint) *uint {
	return &value
}
