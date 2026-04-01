package service

import (
	"amiya-eden/internal/model"
	"testing"
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
