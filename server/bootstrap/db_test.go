package bootstrap

import (
	"strings"
	"testing"
)

func TestCustomIndexStatementsIncludeActiveAffiliationUniqueness(t *testing.T) {
	stmts := newbroCustomIndexStatements()
	if len(stmts) == 0 {
		t.Fatal("expected custom index statements")
	}

	found := false
	for _, stmt := range stmts {
		if strings.Contains(stmt, "newbro_captain_affiliation") &&
			strings.Contains(stmt, "UNIQUE INDEX") &&
			strings.Contains(stmt, "player_user_id") &&
			strings.Contains(stmt, "ended_at IS NULL") &&
			strings.Contains(stmt, "deleted_at IS NULL") {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("expected active affiliation uniqueness statement, got %v", stmts)
	}
}

func TestObsoleteColumnDropsIncludeLegacyPortraitColumns(t *testing.T) {
	drops := obsoleteColumnDrops()
	joined := make([]string, 0, len(drops))
	for _, drop := range drops {
		joined = append(joined, drop.table+"."+drop.col)
	}

	for _, expected := range []string{"user.avatar", "eve_character.portrait_url", "hall_of_fame_card.avatar"} {
		if !strings.Contains(strings.Join(joined, "\n"), expected) {
			t.Fatalf("expected obsolete column drop %q, got %v", expected, joined)
		}
	}
}
