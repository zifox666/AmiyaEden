package service

import (
	"encoding/json"
	"testing"
	"time"
)

func TestSortCaptainCandidatesByLastOnlineDesc(t *testing.T) {
	now := time.Date(2026, 3, 27, 10, 0, 0, 0, time.UTC)
	older := now.Add(-2 * time.Hour)

	candidates := []NewbroCaptainCandidate{
		{CaptainUserID: 3, LastOnlineAt: nil},
		{CaptainUserID: 2, LastOnlineAt: &older},
		{CaptainUserID: 1, LastOnlineAt: &now},
		{CaptainUserID: 4, LastOnlineAt: &now},
	}

	sortCaptainCandidatesByLastOnline(candidates)

	got := []uint{
		candidates[0].CaptainUserID,
		candidates[1].CaptainUserID,
		candidates[2].CaptainUserID,
		candidates[3].CaptainUserID,
	}
	want := []uint{1, 4, 2, 3}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected order at index %d: got %v want %v", i, got, want)
		}
	}
}

func TestNewbroResponseContractsOmitDerivedPortraitURLs(t *testing.T) {
	payload, err := json.Marshal(NewbroCaptainCandidate{
		CaptainUserID:        7,
		CaptainCharacterID:   9001,
		CaptainCharacterName: "Amiya Prime",
		CaptainNickname:      "Captain",
	})
	if err != nil {
		t.Fatalf("marshal captain candidate: %v", err)
	}
	var candidate map[string]any
	if err := json.Unmarshal(payload, &candidate); err != nil {
		t.Fatalf("unmarshal captain candidate: %v", err)
	}
	if _, exists := candidate["captain_portrait_url"]; exists {
		t.Fatalf("expected captain candidate to omit captain_portrait_url, got %#v", candidate["captain_portrait_url"])
	}

	payload, err = json.Marshal(CaptainEligiblePlayerListItem{
		PlayerUserID:        8,
		PlayerCharacterID:   9002,
		PlayerCharacterName: "Amiya Alt",
		PlayerNickname:      "Player",
	})
	if err != nil {
		t.Fatalf("marshal captain eligible player item: %v", err)
	}
	var player map[string]any
	if err := json.Unmarshal(payload, &player); err != nil {
		t.Fatalf("unmarshal captain eligible player item: %v", err)
	}
	if _, exists := player["player_portrait_url"]; exists {
		t.Fatalf("expected captain eligible player item to omit player_portrait_url, got %#v", player["player_portrait_url"])
	}

	payload, err = json.Marshal(NewbroAffiliationSummary{
		AffiliationID:        1,
		CaptainUserID:        7,
		CaptainCharacterID:   9001,
		CaptainCharacterName: "Amiya Prime",
	})
	if err != nil {
		t.Fatalf("marshal newbro affiliation summary: %v", err)
	}
	var summary map[string]any
	if err := json.Unmarshal(payload, &summary); err != nil {
		t.Fatalf("unmarshal newbro affiliation summary: %v", err)
	}
	if _, exists := summary["captain_portrait_url"]; exists {
		t.Fatalf("expected newbro affiliation summary to omit captain_portrait_url, got %#v", summary["captain_portrait_url"])
	}
}
