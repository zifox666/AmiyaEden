package service

import (
	"amiya-eden/internal/model"
	"testing"
	"time"
)

func TestSelectCaptainWalletJournalMatch(t *testing.T) {
	base := time.Date(2026, 3, 27, 20, 0, 0, 0, time.UTC)

	playerJournal := model.EVECharacterWalletJournal{
		ID:        501,
		Date:      base,
		ContextID: 30000142,
		RefType:   "bounty_prizes",
		Reason:    "123: 1, 456: 2",
	}

	t.Run("returns nil when no candidates exist", func(t *testing.T) {
		if got := selectCaptainWalletJournalMatch(playerJournal, nil); got != nil {
			t.Fatalf("expected nil candidate, got %+v", got)
		}
	})

	t.Run("chooses smallest absolute time difference first", func(t *testing.T) {
		candidates := []model.EVECharacterWalletJournal{
			{ID: 2, Date: base.Add(-10 * time.Minute), ContextID: 30000142, RefType: "bounty_prizes", Reason: "123:1,456:2"},
			{ID: 1, Date: base.Add(-3 * time.Minute), ContextID: 30000142, RefType: "bounty_prizes", Reason: "123:1,456:2"},
		}

		got := selectCaptainWalletJournalMatch(playerJournal, candidates)
		if got == nil || got.ID != 1 {
			t.Fatalf("expected closest candidate id=1, got %+v", got)
		}
	})

	t.Run("breaks equal time difference by earlier date then smaller id", func(t *testing.T) {
		candidates := []model.EVECharacterWalletJournal{
			{ID: 10, Date: base.Add(-5 * time.Minute), ContextID: 30000142, RefType: "bounty_prizes", Reason: "123:1,456:2"},
			{ID: 9, Date: base.Add(-5 * time.Minute), ContextID: 30000142, RefType: "bounty_prizes", Reason: "123:1,456:2"},
			{ID: 8, Date: base.Add(5 * time.Minute), ContextID: 30000142, RefType: "bounty_prizes", Reason: "123:1,456:2"},
		}

		got := selectCaptainWalletJournalMatch(playerJournal, candidates)
		if got == nil || got.ID != 9 {
			t.Fatalf("expected earlier-date tie to fall through to smaller id=9, got %+v", got)
		}
	})

	t.Run("matches candidates with different reason as long as ref_type matches", func(t *testing.T) {
		candidates := []model.EVECharacterWalletJournal{
			{ID: 1, Date: base.Add(-1 * time.Minute), ContextID: 30000142, RefType: "bounty_prizes", Reason: "999:1"},
			{ID: 2, Date: base.Add(-3 * time.Minute), ContextID: 30000142, RefType: "bounty_prizes", Reason: "456:1,123:1,456:1"},
			{ID: 3, Date: base.Add(-2 * time.Minute), ContextID: 30000142, RefType: "ess_escrow_transfer"},
		}

		got := selectCaptainWalletJournalMatch(playerJournal, candidates)
		if got == nil || got.ID != 1 {
			t.Fatalf("expected closest time match id=1, got %+v", got)
		}
	})

	t.Run("returns nil when no candidate has matching ref_type", func(t *testing.T) {
		candidates := []model.EVECharacterWalletJournal{
			{ID: 1, Date: base.Add(-1 * time.Minute), ContextID: 30000142, RefType: "ess_escrow_transfer"},
			{ID: 2, Date: base.Add(-2 * time.Minute), ContextID: 30000142, RefType: "ess_escrow_transfer"},
		}

		if got := selectCaptainWalletJournalMatch(playerJournal, candidates); got != nil {
			t.Fatalf("expected nil when ref_types do not match, got %+v", got)
		}
	})
}

func TestShouldConsiderAttributionJournal(t *testing.T) {
	now := time.Date(2026, 3, 27, 12, 0, 0, 0, time.UTC)

	t.Run("accepts bounty_prizes within lookback", func(t *testing.T) {
		journal := model.EVECharacterWalletJournal{
			ID:      1,
			RefType: "bounty_prizes",
			Date:    now.Add(-24 * time.Hour),
		}

		if !shouldConsiderAttributionJournal(journal, now, 30) {
			t.Fatal("expected bounty journal within lookback to be considered")
		}
	})

	t.Run("rejects ess_escrow_transfer because the sync is bounty only", func(t *testing.T) {
		journal := model.EVECharacterWalletJournal{
			ID:      1,
			RefType: "ess_escrow_transfer",
			Date:    now.Add(-24 * time.Hour),
		}

		if shouldConsiderAttributionJournal(journal, now, 30) {
			t.Fatal("did not expect ess journal to be considered directly")
		}
	})

	t.Run("rejects unsupported ref type", func(t *testing.T) {
		journal := model.EVECharacterWalletJournal{
			ID:      1,
			RefType: "player_donation",
			Date:    now,
		}

		if shouldConsiderAttributionJournal(journal, now, 30) {
			t.Fatal("did not expect unsupported ref type to be considered")
		}
	})

	t.Run("rejects records older than lookback window", func(t *testing.T) {
		journal := model.EVECharacterWalletJournal{
			ID:      1,
			RefType: "bounty_prizes",
			Date:    now.AddDate(0, 0, -31),
		}

		if shouldConsiderAttributionJournal(journal, now, 30) {
			t.Fatal("did not expect old journal to be considered")
		}
	})

	t.Run("candidate matching follows the supported ref type list", func(t *testing.T) {
		supportedPlayerAttributionRefTypes["ess_escrow_transfer"] = struct{}{}
		defer delete(supportedPlayerAttributionRefTypes, "ess_escrow_transfer")

		playerJournal := model.EVECharacterWalletJournal{
			ID:      1,
			RefType: "ess_escrow_transfer",
			Date:    now.Add(-time.Minute),
			Reason:  "123:1",
		}
		candidateJournal := model.EVECharacterWalletJournal{
			ID:      2,
			RefType: "ess_escrow_transfer",
			Date:    now,
			Reason:  "123:1",
		}

		if !shouldConsiderAttributionJournal(playerJournal, now, 30) {
			t.Fatal("expected newly supported ref type to be considered")
		}
		if !captainJournalCandidateMatches(playerJournal, candidateJournal) {
			t.Fatal("expected candidate matching to respect the same supported ref type list")
		}
	})
}
