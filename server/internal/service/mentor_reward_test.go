package service

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"fmt"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestValidateMentorRewardStageInputs(t *testing.T) {
	t.Run("accepts strictly increasing valid stages", func(t *testing.T) {
		err := validateMentorRewardStageInputs([]MentorRewardStageInput{
			{StageOrder: 1, Name: "SP 10M", ConditionType: model.MentorConditionSkillPoints, Threshold: 10_000_000, RewardAmount: 100},
			{StageOrder: 2, Name: "PAP 10", ConditionType: model.MentorConditionPapCount, Threshold: 10, RewardAmount: 200},
		})
		if err != nil {
			t.Fatalf("expected valid stages, got %v", err)
		}
	})

	t.Run("rejects non-increasing stage order", func(t *testing.T) {
		err := validateMentorRewardStageInputs([]MentorRewardStageInput{
			{StageOrder: 2, Name: "Second", ConditionType: model.MentorConditionSkillPoints, Threshold: 1, RewardAmount: 1},
			{StageOrder: 2, Name: "Duplicate", ConditionType: model.MentorConditionPapCount, Threshold: 2, RewardAmount: 2},
		})
		if err == nil {
			t.Fatal("expected validation error for duplicate stage order")
		}
	})

	t.Run("rejects invalid condition type", func(t *testing.T) {
		err := validateMentorRewardStageInputs([]MentorRewardStageInput{{
			StageOrder:    1,
			Name:          "Invalid",
			ConditionType: "unknown",
			Threshold:     1,
			RewardAmount:  1,
		}})
		if err == nil {
			t.Fatal("expected validation error for invalid condition type")
		}
	})

	t.Run("rejects non-integer threshold", func(t *testing.T) {
		err := validateMentorRewardStageInputs([]MentorRewardStageInput{{
			StageOrder:    1,
			Name:          "Fractional threshold",
			ConditionType: model.MentorConditionPapCount,
			Threshold:     1.5,
			RewardAmount:  10,
		}})
		if err == nil {
			t.Fatal("expected validation error for non-integer threshold")
		}
	})

	t.Run("rejects non-integer reward amount", func(t *testing.T) {
		err := validateMentorRewardStageInputs([]MentorRewardStageInput{{
			StageOrder:    1,
			Name:          "Fractional reward",
			ConditionType: model.MentorConditionSkillPoints,
			Threshold:     10_000_000,
			RewardAmount:  99.5,
		}})
		if err == nil {
			t.Fatal("expected validation error for non-integer reward amount")
		}
	})
}

func TestIsMentorConditionMet(t *testing.T) {
	metrics := &mentorMetrics{TotalSP: 12_000_000, TotalPap: 18, DaysActive: 45}

	tests := []struct {
		name  string
		stage model.MentorRewardStage
		want  bool
	}{
		{
			name:  "skill points",
			stage: model.MentorRewardStage{ConditionType: model.MentorConditionSkillPoints, Threshold: 10_000_000},
			want:  true,
		},
		{
			name:  "pap count",
			stage: model.MentorRewardStage{ConditionType: model.MentorConditionPapCount, Threshold: 20},
			want:  false,
		},
		{
			name:  "days active",
			stage: model.MentorRewardStage{ConditionType: model.MentorConditionDaysActive, Threshold: 30},
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isMentorConditionMet(tt.stage, metrics); got != tt.want {
				t.Fatalf("isMentorConditionMet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMentorRewardServiceProcessRewardsKeepsStageOrderProgressWhenStagesAreReplaced(t *testing.T) {
	now := time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC)
	db := newMentorRewardTestDB(t)
	useMentorRewardTestDB(t, db)
	fixture := seedMentorRewardTestFixture(t, db, now)

	originalStages := []model.MentorRewardStage{
		{
			StageOrder:    1,
			Name:          "Initial SP milestone",
			ConditionType: model.MentorConditionSkillPoints,
			Threshold:     10_000_000,
			RewardAmount:  100,
		},
		{
			StageOrder:    2,
			Name:          "Future PAP milestone",
			ConditionType: model.MentorConditionPapCount,
			Threshold:     10,
			RewardAmount:  200,
		},
	}
	if err := db.Create(&originalStages).Error; err != nil {
		t.Fatalf("create original stages: %v", err)
	}
	if err := db.Create(&model.MentorRewardDistribution{
		RelationshipID: fixture.relationship.ID,
		StageID:        999,
		StageOrder:     1,
		MentorUserID:   fixture.mentor.ID,
		MenteeUserID:   fixture.mentee.ID,
		RewardAmount:   100,
		DistributedAt:  now.Add(-time.Hour),
		WalletRefID:    "seed-stage-order-1",
	}).Error; err != nil {
		t.Fatalf("create existing distribution: %v", err)
	}

	svc := NewMentorRewardService()
	if err := svc.stageRepo.ReplaceAll([]model.MentorRewardStage{
		{
			StageOrder:    1,
			Name:          "Replaced SP milestone row",
			ConditionType: model.MentorConditionSkillPoints,
			Threshold:     10_000_000,
			RewardAmount:  300,
		},
		{
			StageOrder:    2,
			Name:          "Unmet PAP milestone",
			ConditionType: model.MentorConditionPapCount,
			Threshold:     10,
			RewardAmount:  400,
		},
	}); err != nil {
		t.Fatalf("replace stages: %v", err)
	}

	result, err := svc.ProcessRewards(now)
	if err != nil {
		t.Fatalf("ProcessRewards() error = %v", err)
	}
	if result.ProcessedRelationships != 1 {
		t.Fatalf("ProcessRewards() processed_relationships = %d, want 1", result.ProcessedRelationships)
	}
	if result.RewardsDistributed != 0 {
		t.Fatalf("ProcessRewards() rewards_distributed = %d, want 0", result.RewardsDistributed)
	}
	if result.TotalCoinAwarded != 0 {
		t.Fatalf("ProcessRewards() total_coin_awarded = %v, want 0", result.TotalCoinAwarded)
	}
	if result.GraduatedCount != 0 {
		t.Fatalf("ProcessRewards() graduated_count = %d, want 0", result.GraduatedCount)
	}
	if got := countMentorRewardDistributions(t, db, fixture.relationship.ID); got != 1 {
		t.Fatalf("distribution count = %d, want 1", got)
	}
	if got := countWalletTransactions(t, db, fixture.mentor.ID, model.WalletRefMentorReward); got != 0 {
		t.Fatalf("mentor reward wallet transaction count = %d, want 0", got)
	}
	if got := loadMentorRelationship(t, db, fixture.relationship.ID).Status; got != model.MentorRelationStatusActive {
		t.Fatalf("relationship status = %q, want %q", got, model.MentorRelationStatusActive)
	}
}

func TestMentorRewardServiceProcessActiveRelationshipSnapshotSkipsRevokedRelationship(t *testing.T) {
	now := time.Date(2026, time.April, 3, 9, 0, 0, 0, time.UTC)
	db := newMentorRewardTestDB(t)
	useMentorRewardTestDB(t, db)
	fixture := seedMentorRewardTestFixture(t, db, now)

	stage := model.MentorRewardStage{
		StageOrder:    1,
		Name:          "Immediate reward",
		ConditionType: model.MentorConditionSkillPoints,
		Threshold:     1,
		RewardAmount:  100,
	}
	if err := db.Create(&stage).Error; err != nil {
		t.Fatalf("create stage: %v", err)
	}

	svc := NewMentorRewardService()
	activeRelationships, err := svc.relRepo.ListActiveRelationships()
	if err != nil {
		t.Fatalf("ListActiveRelationships() error = %v", err)
	}
	if len(activeRelationships) != 1 {
		t.Fatalf("ListActiveRelationships() len = %d, want 1", len(activeRelationships))
	}
	if err := svc.relRepo.UpdateStatus(fixture.relationship.ID, model.MentorRelationStatusRevoked, map[string]any{
		"revoked_at": now,
		"revoked_by": fixture.mentor.ID,
	}); err != nil {
		t.Fatalf("revoke relationship: %v", err)
	}

	outcome, err := svc.processActiveRelationshipSnapshot(activeRelationships[0], []model.MentorRewardStage{stage}, &mentorMetrics{
		TotalSP:    10,
		TotalPap:   0,
		DaysActive: 0,
	}, now)
	if err != nil {
		t.Fatalf("processActiveRelationshipSnapshot() error = %v", err)
	}
	if outcome.Processed {
		t.Fatal("expected revoked relationship snapshot to be skipped")
	}
	if outcome.RewardsDistributed != 0 {
		t.Fatalf("rewards_distributed = %d, want 0", outcome.RewardsDistributed)
	}
	if outcome.TotalCoinAwarded != 0 {
		t.Fatalf("total_coin_awarded = %v, want 0", outcome.TotalCoinAwarded)
	}
	if outcome.Graduated {
		t.Fatal("expected revoked relationship to stay ungraduated")
	}
	if got := countMentorRewardDistributions(t, db, fixture.relationship.ID); got != 0 {
		t.Fatalf("distribution count = %d, want 0", got)
	}
	if got := countWalletTransactions(t, db, fixture.mentor.ID, model.WalletRefMentorReward); got != 0 {
		t.Fatalf("mentor reward wallet transaction count = %d, want 0", got)
	}
	relationship := loadMentorRelationship(t, db, fixture.relationship.ID)
	if relationship.Status != model.MentorRelationStatusRevoked {
		t.Fatalf("relationship status = %q, want %q", relationship.Status, model.MentorRelationStatusRevoked)
	}
	if relationship.GraduatedAt != nil {
		t.Fatalf("graduated_at = %v, want nil", relationship.GraduatedAt)
	}
}

type mentorRewardTestFixture struct {
	mentor       model.User
	mentee       model.User
	relationship model.MentorMenteeRelationship
}

func newMentorRewardTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:mentor_reward_test_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.User{},
		&model.EveCharacter{},
		&model.EveCharacterSkill{},
		&model.FleetPapLog{},
		&model.SystemWallet{},
		&model.WalletTransaction{},
		&model.MentorMenteeRelationship{},
		&model.MentorRewardStage{},
		&model.MentorRewardDistribution{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}

func useMentorRewardTestDB(t *testing.T, db *gorm.DB) {
	t.Helper()

	oldDB := global.DB
	global.DB = db
	t.Cleanup(func() {
		global.DB = oldDB
	})
}

func seedMentorRewardTestFixture(t *testing.T, db *gorm.DB, now time.Time) mentorRewardTestFixture {
	t.Helper()

	mentor := model.User{
		Nickname: "Mentor",
		BaseModel: model.BaseModel{
			CreatedAt: now.Add(-60 * 24 * time.Hour),
			UpdatedAt: now.Add(-60 * 24 * time.Hour),
		},
	}
	if err := db.Create(&mentor).Error; err != nil {
		t.Fatalf("create mentor: %v", err)
	}

	lastLoginAt := now.Add(-2 * time.Hour)
	mentee := model.User{
		Nickname:           "Mentee",
		PrimaryCharacterID: 900001,
		LastLoginAt:        &lastLoginAt,
		BaseModel: model.BaseModel{
			CreatedAt: now.Add(-30 * 24 * time.Hour),
			UpdatedAt: now.Add(-30 * 24 * time.Hour),
		},
	}
	if err := db.Create(&mentee).Error; err != nil {
		t.Fatalf("create mentee: %v", err)
	}

	character := model.EveCharacter{
		CharacterID:   mentee.PrimaryCharacterID,
		CharacterName: "Curious Mentee",
		UserID:        mentee.ID,
	}
	if err := db.Create(&character).Error; err != nil {
		t.Fatalf("create mentee character: %v", err)
	}
	if err := db.Create(&model.EveCharacterSkill{
		CharacterID: mentee.PrimaryCharacterID,
		TotalSP:     12_000_000,
	}).Error; err != nil {
		t.Fatalf("create skill snapshot: %v", err)
	}

	respondedAt := now.Add(-29 * 24 * time.Hour)
	relationship := model.MentorMenteeRelationship{
		MenteeUserID:                    mentee.ID,
		MenteePrimaryCharacterIDAtStart: mentee.PrimaryCharacterID,
		MentorUserID:                    mentor.ID,
		Status:                          model.MentorRelationStatusActive,
		AppliedAt:                       now.Add(-30 * 24 * time.Hour),
		RespondedAt:                     &respondedAt,
	}
	if err := db.Create(&relationship).Error; err != nil {
		t.Fatalf("create relationship: %v", err)
	}

	return mentorRewardTestFixture{
		mentor:       mentor,
		mentee:       mentee,
		relationship: relationship,
	}
}

func countMentorRewardDistributions(t *testing.T, db *gorm.DB, relationshipID uint) int64 {
	t.Helper()

	var count int64
	if err := db.Model(&model.MentorRewardDistribution{}).Where("relationship_id = ?", relationshipID).Count(&count).Error; err != nil {
		t.Fatalf("count mentor reward distributions: %v", err)
	}
	return count
}

func countWalletTransactions(t *testing.T, db *gorm.DB, userID uint, refType string) int64 {
	t.Helper()

	var count int64
	if err := db.Model(&model.WalletTransaction{}).Where("user_id = ? AND ref_type = ?", userID, refType).Count(&count).Error; err != nil {
		t.Fatalf("count wallet transactions: %v", err)
	}
	return count
}

func loadMentorRelationship(t *testing.T, db *gorm.DB, relationshipID uint) model.MentorMenteeRelationship {
	t.Helper()

	var relationship model.MentorMenteeRelationship
	if err := db.First(&relationship, relationshipID).Error; err != nil {
		t.Fatalf("load relationship: %v", err)
	}
	return relationship
}
