package service

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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

func TestWelfareAgeRestrictionFailed(t *testing.T) {
	now := time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC)

	bday := func(y, m, d int) *time.Time {
		t := time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
		return &t
	}

	tests := []struct {
		name       string
		characters []model.EveCharacter
		maxMonths  *int
		want       bool
	}{
		{
			name:       "nil limit never blocks",
			characters: []model.EveCharacter{{Birthday: bday(2024, 1, 1)}},
			maxMonths:  nil,
			want:       false,
		},
		{
			name:       "zero limit never blocks",
			characters: []model.EveCharacter{{Birthday: bday(2024, 1, 1)}},
			maxMonths:  func() *int { v := 0; return &v }(),
			want:       false,
		},
		{
			name:       "old character blocks the welfare",
			characters: []model.EveCharacter{{Birthday: bday(2024, 1, 1)}},
			maxMonths:  func() *int { v := 12; return &v }(),
			want:       true,
		},
		{
			name:       "young characters pass the welfare",
			characters: []model.EveCharacter{{Birthday: bday(2026, 1, 1)}},
			maxMonths:  func() *int { v := 12; return &v }(),
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := welfareAgeRestrictionFailed(tt.characters, tt.maxMonths, now)
			if got != tt.want {
				t.Fatalf("welfareAgeRestrictionFailed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWelfareMinimumPapRestrictionFailed(t *testing.T) {
	tests := []struct {
		name      string
		minimum   *int
		totalPap  float64
		wantBlock bool
	}{
		{
			name:      "nil minimum never blocks",
			minimum:   nil,
			totalPap:  0,
			wantBlock: false,
		},
		{
			name:      "zero minimum never blocks",
			minimum:   func() *int { v := 0; return &v }(),
			totalPap:  0,
			wantBlock: false,
		},
		{
			name:      "total equal to minimum blocks (strictly greater required)",
			minimum:   func() *int { v := 10; return &v }(),
			totalPap:  10,
			wantBlock: true,
		},
		{
			name:      "total below minimum blocks",
			minimum:   func() *int { v := 10; return &v }(),
			totalPap:  9.9,
			wantBlock: true,
		},
		{
			name:      "total above minimum passes",
			minimum:   func() *int { v := 10; return &v }(),
			totalPap:  10.1,
			wantBlock: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := welfareMinimumPapRestrictionFailed(tt.minimum, tt.totalPap)
			if got != tt.wantBlock {
				t.Fatalf("welfareMinimumPapRestrictionFailed() = %v, want %v", got, tt.wantBlock)
			}
		})
	}
}

func TestBuildEligibleWelfareRespKeepsMinimumPapRestrictedWelfareVisible(t *testing.T) {
	svc := &WelfareService{}

	user := &model.User{QQ: "12345", DiscordID: "discord-1"}
	characters := []model.EveCharacter{
		{CharacterID: 1001, CharacterName: "Alpha"},
		{CharacterID: 1002, CharacterName: "Beta"},
	}

	t.Run("per user welfare stays visible but disabled", func(t *testing.T) {
		minimumPap := func() *int { v := 10; return &v }()
		welfare := model.Welfare{
			BaseModel:        model.BaseModel{ID: 12},
			Name:             "Per User Minimum PAP",
			DistMode:         model.WelfareDistModePerUser,
			MinimumPap:       minimumPap,
			RequireSkillPlan: false,
		}

		got, ok := svc.buildEligibleWelfareResp(user, characters, nil, welfare, nil, true)
		if !ok {
			t.Fatal("expected minimum PAP restricted welfare to stay visible")
		}
		if got.CanApplyNow {
			t.Fatal("expected per-user welfare to be disabled when minimum PAP is not met")
		}
	})

	t.Run("per character welfare keeps rows visible but disabled", func(t *testing.T) {
		minimumPap := func() *int { v := 10; return &v }()
		welfare := model.Welfare{
			BaseModel:        model.BaseModel{ID: 13},
			Name:             "Per Character Minimum PAP",
			DistMode:         model.WelfareDistModePerCharacter,
			MinimumPap:       minimumPap,
			RequireSkillPlan: false,
		}

		got, ok := svc.buildEligibleWelfareResp(user, characters, nil, welfare, nil, true)
		if !ok {
			t.Fatal("expected minimum PAP restricted welfare to stay visible")
		}
		if len(got.EligibleCharacters) != 2 {
			t.Fatalf("expected 2 character rows, got %d", len(got.EligibleCharacters))
		}
		for _, row := range got.EligibleCharacters {
			if row.CanApplyNow {
				t.Fatal("expected per-character welfare rows to be disabled when minimum PAP is not met")
			}
		}
	})
}

func TestBuildEligibleWelfareRespIncludesFutureSkillOptions(t *testing.T) {
	svc := &WelfareService{}

	user := &model.User{
		QQ:        "12345",
		DiscordID: "discord-1",
	}
	characters := []model.EveCharacter{
		{CharacterID: 1001, CharacterName: "Alpha"},
		{CharacterID: 1002, CharacterName: "Beta"},
	}

	t.Run("per user welfare stays visible but disabled when only future skill growth could satisfy it", func(t *testing.T) {
		welfare := model.Welfare{
			BaseModel:        model.BaseModel{ID: 10},
			Name:             "Per User Welfare",
			DistMode:         model.WelfareDistModePerUser,
			RequireSkillPlan: true,
			SkillPlanIDs:     []uint{7},
			SkillPlanNames:   []string{"Alpha Plan"},
		}
		skillCache := map[int64]map[uint]bool{
			1001: {7: false},
			1002: {7: false},
		}

		got, ok := svc.buildEligibleWelfareResp(user, characters, nil, welfare, skillCache, false)
		if !ok {
			t.Fatal("expected future-eligible welfare to stay visible")
		}
		if got.CanApplyNow {
			t.Fatal("expected per-user welfare to be disabled when no character satisfies the skill plan yet")
		}
		if len(got.SkillPlanNames) != 1 || got.SkillPlanNames[0] != "Alpha Plan" {
			t.Fatalf("expected skill plan names to be propagated, got %+v", got.SkillPlanNames)
		}
		if len(got.EligibleCharacters) != 0 {
			t.Fatalf("expected no character rows for per-user welfare, got %d", len(got.EligibleCharacters))
		}
	})

	t.Run("per character welfare keeps both current and future rows", func(t *testing.T) {
		welfare := model.Welfare{
			BaseModel:        model.BaseModel{ID: 11},
			Name:             "Per Character Welfare",
			DistMode:         model.WelfareDistModePerCharacter,
			RequireSkillPlan: true,
			SkillPlanIDs:     []uint{7},
		}
		skillCache := map[int64]map[uint]bool{
			1001: {7: true},
			1002: {7: false},
		}

		got, ok := svc.buildEligibleWelfareResp(user, characters, nil, welfare, skillCache, false)
		if !ok {
			t.Fatal("expected per-character welfare to stay visible")
		}
		if len(got.EligibleCharacters) != 2 {
			t.Fatalf("expected 2 character rows, got %d", len(got.EligibleCharacters))
		}
		if !got.EligibleCharacters[0].CanApplyNow {
			t.Fatal("expected the first character to be currently eligible")
		}
		if got.EligibleCharacters[1].CanApplyNow {
			t.Fatal("expected the second character to be future-only")
		}
	})
}

func TestSkillPlanNamesForWelfarePreservesConfiguredOrder(t *testing.T) {
	got := skillPlanNamesForWelfare([]uint{7, 3, 9}, map[uint]string{
		3: "Shield Plan",
		7: "Armor Plan",
		9: "",
	})

	want := []string{"Armor Plan", "Shield Plan"}
	if len(got) != len(want) {
		t.Fatalf("expected %d plan names, got %d (%+v)", len(want), len(got), got)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("got[%d] = %q, want %q", index, got[index], want[index])
		}
	}
}

func TestBuildMyApplicationResponsesIncludesReviewerNickname(t *testing.T) {
	createdAt := time.Date(2026, 3, 30, 8, 0, 0, 0, time.UTC)
	reviewedAt := time.Date(2026, 3, 30, 9, 30, 0, 0, time.UTC)

	apps := []model.WelfareApplication{
		{
			BaseModel:     model.BaseModel{ID: 1, CreatedAt: createdAt},
			WelfareID:     10,
			CharacterName: "Alpha",
			Status:        model.WelfareAppStatusDelivered,
			ReviewedBy:    77,
			ReviewedAt:    &reviewedAt,
		},
		{
			BaseModel:     model.BaseModel{ID: 2, CreatedAt: createdAt},
			WelfareID:     11,
			CharacterName: "Beta",
			Status:        model.WelfareAppStatusRequested,
		},
	}

	got := buildMyApplicationResponses(apps, map[uint]string{
		10: "Starter Pack",
		11: "Advanced Pack",
	}, map[uint]string{
		77: "Officer Fox",
	})

	if len(got) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(got))
	}
	if got[0].ReviewerName != "Officer Fox" {
		t.Fatalf("expected reviewer nickname to be included, got %q", got[0].ReviewerName)
	}
	if got[1].ReviewerName != "" {
		t.Fatalf("expected empty reviewer nickname for unreviewed applications, got %q", got[1].ReviewerName)
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

func TestAdminReviewApplicationDeliverCreditsConfiguredFuxiCoin(t *testing.T) {
	db := newWelfareServiceTestDB(t)
	useWelfareServiceTestDB(t, db)

	payout := 25
	welfare := &model.Welfare{
		Name:          "Starter Pack",
		DistMode:      model.WelfareDistModePerUser,
		PayByFuxiCoin: &payout,
		Status:        model.WelfareStatusActive,
		CreatedBy:     1,
	}
	if err := db.Create(welfare).Error; err != nil {
		t.Fatalf("create welfare: %v", err)
	}

	userID := uint(42)
	app := &model.WelfareApplication{
		WelfareID:     welfare.ID,
		UserID:        &userID,
		CharacterID:   90000001,
		CharacterName: "Pilot One",
		Status:        model.WelfareAppStatusRequested,
	}
	if err := db.Create(app).Error; err != nil {
		t.Fatalf("create application: %v", err)
	}

	svc := NewWelfareService()
	if _, err := svc.AdminReviewApplication(app.ID, 77, &AdminReviewApplicationRequest{Action: "deliver"}); err != nil {
		t.Fatalf("AdminReviewApplication() error = %v", err)
	}

	var updated model.WelfareApplication
	if err := db.First(&updated, app.ID).Error; err != nil {
		t.Fatalf("reload application: %v", err)
	}
	if updated.Status != model.WelfareAppStatusDelivered {
		t.Fatalf("status = %q, want %q", updated.Status, model.WelfareAppStatusDelivered)
	}
	if updated.ReviewedBy != 77 {
		t.Fatalf("reviewed_by = %d, want 77", updated.ReviewedBy)
	}
	if updated.ReviewedAt == nil {
		t.Fatal("expected reviewed_at to be set")
	}

	var wallet model.SystemWallet
	if err := db.Where("user_id = ?", userID).First(&wallet).Error; err != nil {
		t.Fatalf("load wallet: %v", err)
	}
	if wallet.Balance != 25 {
		t.Fatalf("wallet balance = %v, want 25", wallet.Balance)
	}

	var txs []model.WalletTransaction
	if err := db.Order("id ASC").Find(&txs).Error; err != nil {
		t.Fatalf("load wallet transactions: %v", err)
	}
	if len(txs) != 1 {
		t.Fatalf("wallet transaction count = %d, want 1", len(txs))
	}
	if txs[0].RefType != model.WalletRefWelfarePayout {
		t.Fatalf("wallet tx ref_type = %q, want %q", txs[0].RefType, model.WalletRefWelfarePayout)
	}
	if txs[0].RefID != fmt.Sprintf("welfare_application:%d", app.ID) {
		t.Fatalf("wallet tx ref_id = %q", txs[0].RefID)
	}
	if txs[0].OperatorID != 77 {
		t.Fatalf("wallet tx operator_id = %d, want 77", txs[0].OperatorID)
	}
}

func TestAdminReviewApplicationDeliverAttemptsInGameMailButIgnoresMailErrors(t *testing.T) {
	db := newWelfareServiceTestDB(t)
	useWelfareServiceTestDB(t, db)

	welfare := &model.Welfare{
		Name:      "Starter Pack",
		DistMode:  model.WelfareDistModePerUser,
		Status:    model.WelfareStatusActive,
		CreatedBy: 1,
	}
	if err := db.Create(welfare).Error; err != nil {
		t.Fatalf("create welfare: %v", err)
	}

	userID := uint(42)
	app := &model.WelfareApplication{
		WelfareID:     welfare.ID,
		UserID:        &userID,
		CharacterID:   90000001,
		CharacterName: "Pilot One",
		Status:        model.WelfareAppStatusRequested,
	}
	if err := db.Create(app).Error; err != nil {
		t.Fatalf("create application: %v", err)
	}

	svc := NewWelfareService()
	mailAttempted := false
	svc.deliveryMailSender = func(ctx context.Context, reviewerID uint, deliveredWelfare *model.Welfare, deliveredApp *model.WelfareApplication) (MailAttemptSummary, error) {
		mailAttempted = true
		if reviewerID != 77 {
			t.Fatalf("reviewerID = %d, want 77", reviewerID)
		}
		if deliveredWelfare.ID != welfare.ID {
			t.Fatalf("welfare id = %d, want %d", deliveredWelfare.ID, welfare.ID)
		}
		if deliveredApp.ID != app.ID {
			t.Fatalf("application id = %d, want %d", deliveredApp.ID, app.ID)
		}
		return MailAttemptSummary{
			MailSenderCharacterID:      90000077,
			MailSenderCharacterName:    "Officer Main",
			MailRecipientCharacterID:   90000042,
			MailRecipientCharacterName: "Pilot Main",
		}, errors.New("mail failed")
	}

	mailSummary, err := svc.AdminReviewApplication(app.ID, 77, &AdminReviewApplicationRequest{Action: "deliver"})
	if err != nil {
		t.Fatalf("AdminReviewApplication() error = %v", err)
	}
	if !mailAttempted {
		t.Fatal("expected deliver to attempt in-game mail after successful delivery")
	}
	if !strings.Contains(mailSummary.MailError, "mail failed") {
		t.Fatalf("mailError = %q, want to contain %q", mailSummary.MailError, "mail failed")
	}
	if mailSummary.MailSenderCharacterID != 90000077 || mailSummary.MailRecipientCharacterID != 90000042 {
		t.Fatalf("unexpected mail summary: %#v", mailSummary)
	}

	var updated model.WelfareApplication
	if err := db.First(&updated, app.ID).Error; err != nil {
		t.Fatalf("reload application: %v", err)
	}
	if updated.Status != model.WelfareAppStatusDelivered {
		t.Fatalf("status = %q, want %q", updated.Status, model.WelfareAppStatusDelivered)
	}
	if updated.ReviewedBy != 77 {
		t.Fatalf("reviewed_by = %d, want 77", updated.ReviewedBy)
	}
}

func TestBuildWelfareDeliveryMailContentIncludesBilingualOfficerNotice(t *testing.T) {
	subject, body := buildWelfareDeliveryMailContent("Starter Pack", "Amiya")

	if !strings.Contains(subject, "福利发放通知") || !strings.Contains(subject, "Welfare Delivery Notice") {
		t.Fatalf("unexpected subject: %q", subject)
	}
	if !strings.Contains(body, "你的福利「Starter Pack」已由福利官 Amiya 发放") {
		t.Fatalf("expected Chinese body to mention welfare name and officer nickname, got %q", body)
	}
	if !strings.Contains(body, "福利名称：Starter Pack") {
		t.Fatalf("expected Chinese body to include welfare name detail, got %q", body)
	}
	if !strings.Contains(body, "请检查你的伏羲币钱包或合同") {
		t.Fatalf("expected Chinese body to mention FuxiCoin wallet or contract, got %q", body)
	}
	if !strings.Contains(body, "如有疑问，请联系处理此申请的福利官。") {
		t.Fatalf("expected Chinese body to include a professional follow-up note, got %q", body)
	}
	if !strings.Contains(body, "Your welfare \"Starter Pack\" has been delivered by officer Amiya.") {
		t.Fatalf("expected English body to mention welfare name and officer nickname, got %q", body)
	}
	if !strings.Contains(body, "Welfare: Starter Pack") {
		t.Fatalf("expected English body to include welfare detail, got %q", body)
	}
	if !strings.Contains(body, "Please check your FuxiCoin wallet or contract.") {
		t.Fatalf("expected English body to mention FuxiCoin wallet or contract, got %q", body)
	}
	if !strings.Contains(body, "If anything looks incorrect, please contact the officer who handled this delivery.") {
		t.Fatalf("expected English body to include a professional follow-up note, got %q", body)
	}
}

func TestAdminReviewApplicationDeliverWithoutConfiguredPayoutSkipsWalletCredit(t *testing.T) {
	db := newWelfareServiceTestDB(t)
	useWelfareServiceTestDB(t, db)

	welfare := &model.Welfare{
		Name:      "Starter Pack",
		DistMode:  model.WelfareDistModePerUser,
		Status:    model.WelfareStatusActive,
		CreatedBy: 1,
	}
	if err := db.Create(welfare).Error; err != nil {
		t.Fatalf("create welfare: %v", err)
	}

	userID := uint(42)
	app := &model.WelfareApplication{
		WelfareID:     welfare.ID,
		UserID:        &userID,
		CharacterID:   90000001,
		CharacterName: "Pilot One",
		Status:        model.WelfareAppStatusRequested,
	}
	if err := db.Create(app).Error; err != nil {
		t.Fatalf("create application: %v", err)
	}

	svc := NewWelfareService()
	if _, err := svc.AdminReviewApplication(app.ID, 77, &AdminReviewApplicationRequest{Action: "deliver"}); err != nil {
		t.Fatalf("AdminReviewApplication() error = %v", err)
	}

	var txs []model.WalletTransaction
	if err := db.Find(&txs).Error; err != nil {
		t.Fatalf("load wallet transactions: %v", err)
	}
	if len(txs) != 0 {
		t.Fatalf("wallet transaction count = %d, want 0", len(txs))
	}
}

func TestAdminReviewApplicationDeliverUsesApprovalTimePayoutConfig(t *testing.T) {
	db := newWelfareServiceTestDB(t)
	useWelfareServiceTestDB(t, db)

	initialPayout := 10
	welfare := &model.Welfare{
		Name:          "Starter Pack",
		DistMode:      model.WelfareDistModePerUser,
		PayByFuxiCoin: &initialPayout,
		Status:        model.WelfareStatusActive,
		CreatedBy:     1,
	}
	if err := db.Create(welfare).Error; err != nil {
		t.Fatalf("create welfare: %v", err)
	}

	userID := uint(42)
	app := &model.WelfareApplication{
		WelfareID:     welfare.ID,
		UserID:        &userID,
		CharacterID:   90000001,
		CharacterName: "Pilot One",
		Status:        model.WelfareAppStatusRequested,
	}
	if err := db.Create(app).Error; err != nil {
		t.Fatalf("create application: %v", err)
	}

	updatedPayout := 35
	if err := db.Model(&model.Welfare{}).Where("id = ?", welfare.ID).
		Update("pay_by_fuxi_coin", updatedPayout).Error; err != nil {
		t.Fatalf("update welfare payout: %v", err)
	}

	svc := NewWelfareService()
	if _, err := svc.AdminReviewApplication(app.ID, 77, &AdminReviewApplicationRequest{Action: "deliver"}); err != nil {
		t.Fatalf("AdminReviewApplication() error = %v", err)
	}

	var wallet model.SystemWallet
	if err := db.Where("user_id = ?", userID).First(&wallet).Error; err != nil {
		t.Fatalf("load wallet: %v", err)
	}
	if wallet.Balance != 35 {
		t.Fatalf("wallet balance = %v, want 35", wallet.Balance)
	}

	var txs []model.WalletTransaction
	if err := db.Order("id ASC").Find(&txs).Error; err != nil {
		t.Fatalf("load wallet transactions: %v", err)
	}
	if len(txs) != 1 {
		t.Fatalf("wallet transaction count = %d, want 1", len(txs))
	}
	if txs[0].Amount != 35 {
		t.Fatalf("wallet tx amount = %v, want 35", txs[0].Amount)
	}
}

func TestAdminReviewApplicationDeliverWithConfiguredPayoutRequiresUserID(t *testing.T) {
	db := newWelfareServiceTestDB(t)
	useWelfareServiceTestDB(t, db)

	payout := 25
	welfare := &model.Welfare{
		Name:          "Starter Pack",
		DistMode:      model.WelfareDistModePerUser,
		PayByFuxiCoin: &payout,
		Status:        model.WelfareStatusActive,
		CreatedBy:     1,
	}
	if err := db.Create(welfare).Error; err != nil {
		t.Fatalf("create welfare: %v", err)
	}

	app := &model.WelfareApplication{
		WelfareID:     welfare.ID,
		CharacterID:   90000001,
		CharacterName: "Pilot One",
		Status:        model.WelfareAppStatusRequested,
	}
	if err := db.Create(app).Error; err != nil {
		t.Fatalf("create application: %v", err)
	}

	svc := NewWelfareService()
	if _, err := svc.AdminReviewApplication(app.ID, 77, &AdminReviewApplicationRequest{Action: "deliver"}); err == nil {
		t.Fatal("expected deliver to fail when payout requires user_id")
	}

	var updated model.WelfareApplication
	if err := db.First(&updated, app.ID).Error; err != nil {
		t.Fatalf("reload application: %v", err)
	}
	if updated.Status != model.WelfareAppStatusRequested {
		t.Fatalf("status = %q, want %q", updated.Status, model.WelfareAppStatusRequested)
	}

	var txs []model.WalletTransaction
	if err := db.Find(&txs).Error; err != nil {
		t.Fatalf("load wallet transactions: %v", err)
	}
	if len(txs) != 0 {
		t.Fatalf("wallet transaction count = %d, want 0", len(txs))
	}
}

func TestAdminReviewApplicationRejectStillStampsReviewerAuditFields(t *testing.T) {
	db := newWelfareServiceTestDB(t)
	useWelfareServiceTestDB(t, db)

	payout := 25
	welfare := &model.Welfare{
		Name:          "Starter Pack",
		DistMode:      model.WelfareDistModePerUser,
		PayByFuxiCoin: &payout,
		Status:        model.WelfareStatusActive,
		CreatedBy:     1,
	}
	if err := db.Create(welfare).Error; err != nil {
		t.Fatalf("create welfare: %v", err)
	}

	userID := uint(42)
	app := &model.WelfareApplication{
		WelfareID:     welfare.ID,
		UserID:        &userID,
		CharacterID:   90000001,
		CharacterName: "Pilot One",
		Status:        model.WelfareAppStatusRequested,
	}
	if err := db.Create(app).Error; err != nil {
		t.Fatalf("create application: %v", err)
	}

	svc := NewWelfareService()
	if _, err := svc.AdminReviewApplication(app.ID, 77, &AdminReviewApplicationRequest{Action: "reject"}); err != nil {
		t.Fatalf("AdminReviewApplication() error = %v", err)
	}

	var updated model.WelfareApplication
	if err := db.First(&updated, app.ID).Error; err != nil {
		t.Fatalf("reload application: %v", err)
	}
	if updated.Status != model.WelfareAppStatusRejected {
		t.Fatalf("status = %q, want %q", updated.Status, model.WelfareAppStatusRejected)
	}
	if updated.ReviewedBy != 77 {
		t.Fatalf("reviewed_by = %d, want 77", updated.ReviewedBy)
	}
	if updated.ReviewedAt == nil {
		t.Fatal("expected reviewed_at to be set")
	}

	var txs []model.WalletTransaction
	if err := db.Find(&txs).Error; err != nil {
		t.Fatalf("load wallet transactions: %v", err)
	}
	if len(txs) != 0 {
		t.Fatalf("wallet transaction count = %d, want 0", len(txs))
	}
}

func TestAdminReviewApplicationSecondDeliverAttemptDoesNotCreateSecondPayout(t *testing.T) {
	db := newWelfareServiceTestDB(t)
	useWelfareServiceTestDB(t, db)

	payout := 25
	welfare := &model.Welfare{
		Name:          "Starter Pack",
		DistMode:      model.WelfareDistModePerUser,
		PayByFuxiCoin: &payout,
		Status:        model.WelfareStatusActive,
		CreatedBy:     1,
	}
	if err := db.Create(welfare).Error; err != nil {
		t.Fatalf("create welfare: %v", err)
	}

	userID := uint(42)
	app := &model.WelfareApplication{
		WelfareID:     welfare.ID,
		UserID:        &userID,
		CharacterID:   90000001,
		CharacterName: "Pilot One",
		Status:        model.WelfareAppStatusRequested,
	}
	if err := db.Create(app).Error; err != nil {
		t.Fatalf("create application: %v", err)
	}

	svc := NewWelfareService()
	if _, err := svc.AdminReviewApplication(app.ID, 77, &AdminReviewApplicationRequest{Action: "deliver"}); err != nil {
		t.Fatalf("first deliver error = %v", err)
	}
	if _, err := svc.AdminReviewApplication(app.ID, 77, &AdminReviewApplicationRequest{Action: "deliver"}); err == nil {
		t.Fatal("expected second deliver attempt to fail")
	}

	var txs []model.WalletTransaction
	if err := db.Find(&txs).Error; err != nil {
		t.Fatalf("load wallet transactions: %v", err)
	}
	if len(txs) != 1 {
		t.Fatalf("wallet transaction count = %d, want 1", len(txs))
	}
}

func TestImportWelfareRecordsDoesNotCreateWalletTransactions(t *testing.T) {
	db := newWelfareServiceTestDB(t)
	useWelfareServiceTestDB(t, db)

	payout := 25
	welfare := &model.Welfare{
		Name:          "Starter Pack",
		DistMode:      model.WelfareDistModePerUser,
		PayByFuxiCoin: &payout,
		Status:        model.WelfareStatusActive,
		CreatedBy:     1,
	}
	if err := db.Create(welfare).Error; err != nil {
		t.Fatalf("create welfare: %v", err)
	}

	svc := NewWelfareService()
	count, err := svc.ImportWelfareRecords(&ImportWelfareRecordsRequest{
		WelfareID: welfare.ID,
		CSV:       "Alpha,12345\nBeta,67890",
	})
	if err != nil {
		t.Fatalf("ImportWelfareRecords() error = %v", err)
	}
	if count != 2 {
		t.Fatalf("import count = %d, want 2", count)
	}

	var txs []model.WalletTransaction
	if err := db.Find(&txs).Error; err != nil {
		t.Fatalf("load wallet transactions: %v", err)
	}
	if len(txs) != 0 {
		t.Fatalf("wallet transaction count = %d, want 0", len(txs))
	}
}

func TestAdminCreateWelfareRejectsNegativePayByFuxiCoin(t *testing.T) {
	db := newWelfareServiceTestDB(t)
	useWelfareServiceTestDB(t, db)

	negativePayout := -1
	svc := NewWelfareService()
	err := svc.AdminCreateWelfare(&model.Welfare{
		Name:          "Starter Pack",
		DistMode:      model.WelfareDistModePerUser,
		PayByFuxiCoin: &negativePayout,
		Status:        model.WelfareStatusActive,
		CreatedBy:     1,
	})
	if err == nil {
		t.Fatal("expected create to reject negative pay_by_fuxi_coin")
	}
}

func TestAdminUpdateWelfareRejectsNegativePayByFuxiCoin(t *testing.T) {
	db := newWelfareServiceTestDB(t)
	useWelfareServiceTestDB(t, db)

	welfare := &model.Welfare{
		Name:      "Starter Pack",
		DistMode:  model.WelfareDistModePerUser,
		Status:    model.WelfareStatusActive,
		CreatedBy: 1,
	}
	if err := db.Create(welfare).Error; err != nil {
		t.Fatalf("create welfare: %v", err)
	}

	negativePayout := -1
	svc := NewWelfareService()
	_, err := svc.AdminUpdateWelfare(welfare.ID, &AdminUpdateWelfareRequest{
		Name:             welfare.Name,
		Description:      welfare.Description,
		DistMode:         welfare.DistMode,
		PayByFuxiCoin:    &negativePayout,
		RequireSkillPlan: false,
		Status:           welfare.Status,
	})
	if err == nil {
		t.Fatal("expected update to reject negative pay_by_fuxi_coin")
	}
}

func newWelfareServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:welfare_service_test_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.Welfare{},
		&model.WelfareSkillPlan{},
		&model.WelfareApplication{},
		&model.SystemWallet{},
		&model.WalletTransaction{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}

func useWelfareServiceTestDB(t *testing.T, db *gorm.DB) {
	t.Helper()

	oldDB := global.DB
	global.DB = db
	t.Cleanup(func() {
		global.DB = oldDB
	})
}
