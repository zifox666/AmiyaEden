package service

import (
	"amiya-eden/internal/model"
	"testing"
	"time"
)

func TestBuildSrpApplicationResponsesIncludesLastActorNickname(t *testing.T) {
	reviewerID := uint(77)
	payerID := uint(88)
	fleetID := "fleet-actor"
	createdAt := time.Date(2026, time.April, 3, 9, 0, 0, 0, time.UTC)

	apps := []model.SrpApplication{
		{
			ID:            1,
			UserID:        10,
			CharacterName: "Pilot One",
			FleetID:       &fleetID,
			ReviewStatus:  model.SrpReviewApproved,
			PayoutStatus:  model.SrpPayoutPaid,
			ReviewedBy:    &reviewerID,
			PaidBy:        &payerID,
			CreatedAt:     createdAt,
		},
		{
			ID:            2,
			UserID:        11,
			CharacterName: "Pilot Two",
			ReviewStatus:  model.SrpReviewRejected,
			PayoutStatus:  model.SrpPayoutNotPaid,
			ReviewedBy:    &reviewerID,
			CreatedAt:     createdAt,
		},
		{
			ID:            3,
			UserID:        12,
			CharacterName: "Pilot Three",
			ReviewStatus:  model.SrpReviewSubmitted,
			PayoutStatus:  model.SrpPayoutNotPaid,
			CreatedAt:     createdAt,
		},
	}

	got := buildSrpApplicationResponses(
		apps,
		map[uint]model.User{
			10: {Nickname: "Applicant One"},
			11: {Nickname: "Applicant Two"},
			77: {Nickname: "Reviewer Wolf"},
			88: {Nickname: "Payout Bear"},
		},
		map[string]model.Fleet{
			fleetID: {Title: "Night Escort", FCCharacterName: "FC Iris"},
		},
	)

	if len(got) != 3 {
		t.Fatalf("expected 3 responses, got %d", len(got))
	}
	if got[0].Nickname != "Applicant One" {
		t.Fatalf("expected applicant nickname, got %q", got[0].Nickname)
	}
	if got[0].FleetTitle != "Night Escort" || got[0].FleetFCName != "FC Iris" {
		t.Fatalf("expected fleet info to be preserved, got %+v", got[0])
	}
	if got[0].LastActorNickname != "Payout Bear" {
		t.Fatalf("expected paid application to use payer nickname, got %q", got[0].LastActorNickname)
	}
	if got[1].LastActorNickname != "Reviewer Wolf" {
		t.Fatalf("expected rejected application to use reviewer nickname, got %q", got[1].LastActorNickname)
	}
	if got[2].LastActorNickname != "" {
		t.Fatalf("expected submitted application to have no last actor nickname, got %q", got[2].LastActorNickname)
	}
}
