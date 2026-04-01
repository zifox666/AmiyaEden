package service

import (
	"testing"
	"time"
)

func TestEvaluateMenteeEligibility(t *testing.T) {
	rules := MenteeEligibilityRules{
		MaxCharacterSP:    4_000_000,
		MaxAccountAgeDays: 7,
	}
	now := time.Date(2026, time.April, 1, 12, 0, 0, 0, time.UTC)

	t.Run("eligible when account is recent and all characters stay below threshold", func(t *testing.T) {
		result := EvaluateMenteeEligibility(
			now.Add(-6*24*time.Hour),
			[]MenteeCharacterSnapshot{{CharacterID: 9001, TotalSP: 3_999_999}},
			now,
			rules,
		)

		if !result.IsEligible {
			t.Fatalf("expected mentee to be eligible, got %+v", result)
		}
		if result.DisqualifiedReason != "" {
			t.Fatalf("expected empty disqualified reason, got %q", result.DisqualifiedReason)
		}
	})

	t.Run("disqualifies when account is too old", func(t *testing.T) {
		result := EvaluateMenteeEligibility(
			now.Add(-8*24*time.Hour),
			[]MenteeCharacterSnapshot{{CharacterID: 9001, TotalSP: 1_000_000}},
			now,
			rules,
		)

		if result.IsEligible {
			t.Fatalf("expected mentee to be ineligible, got %+v", result)
		}
		if result.DisqualifiedReason != MenteeDisqualifiedReasonAccountTooOld {
			t.Fatalf("expected account-too-old reason, got %q", result.DisqualifiedReason)
		}
	})

	t.Run("disqualifies when no characters are bound", func(t *testing.T) {
		result := EvaluateMenteeEligibility(now.Add(-24*time.Hour), nil, now, rules)

		if result.IsEligible {
			t.Fatalf("expected mentee to be ineligible, got %+v", result)
		}
		if result.DisqualifiedReason != MenteeDisqualifiedReasonNoCharacters {
			t.Fatalf("expected no-characters reason, got %q", result.DisqualifiedReason)
		}
	})

	t.Run("disqualifies when any character reaches threshold", func(t *testing.T) {
		result := EvaluateMenteeEligibility(
			now.Add(-24*time.Hour),
			[]MenteeCharacterSnapshot{{CharacterID: 9001, TotalSP: 4_000_000}},
			now,
			rules,
		)

		if result.IsEligible {
			t.Fatalf("expected mentee to be ineligible, got %+v", result)
		}
		if result.DisqualifiedReason != MenteeDisqualifiedReasonSkillPointsTooHigh {
			t.Fatalf("expected skill-points-too-high reason, got %q", result.DisqualifiedReason)
		}
	})
}
