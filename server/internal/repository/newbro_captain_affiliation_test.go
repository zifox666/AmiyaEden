package repository

import (
	"amiya-eden/internal/model"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"
)

func TestBuildCaptainEligiblePlayerListQueryIncludesEligibilityAndSearchFilters(t *testing.T) {
	db := newDryRunPostgresDB(t)

	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return buildCaptainEligiblePlayerListQuery(tx, 42, "bee").
			Find(&[]model.User{})
	})

	if !strings.Contains(sql, `FROM "user"`) {
		t.Fatalf("expected user select, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `JOIN newbro_player_state`) {
		t.Fatalf("expected eligibility join, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `is_currently_newbro =`) {
		t.Fatalf("expected current-newbro filter, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `LEFT JOIN eve_character AS primary_character`) {
		t.Fatalf("expected primary character join for search/display, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `nickname LIKE`) || !strings.Contains(sql, `primary_character.character_name LIKE`) {
		t.Fatalf("expected nickname and character name search, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `LEFT JOIN newbro_captain_affiliation AS current_affiliation`) {
		t.Fatalf("expected active affiliation join, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `current_affiliation.captain_user_id <>`) {
		t.Fatalf("expected query to exclude players already under the same captain, got SQL: %s", sql)
	}
}

func TestBuildCaptainEligiblePlayerListSelectQueryProjectsSortAliasForRecentLoginOrdering(t *testing.T) {
	db := newDryRunPostgresDB(t)

	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return buildCaptainEligiblePlayerListSelectQuery(tx, 42, "bee", 2, 10).
			Find(&[]model.User{})
	})

	if !strings.Contains(sql, `SELECT DISTINCT "user".*`) {
		t.Fatalf("expected eligible player query to select full distinct user rows, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `AS player_sort_name`) {
		t.Fatalf("expected eligible player query to project sort alias for Postgres ordering, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `ORDER BY "user".last_login_at DESC NULLS LAST, player_sort_name ASC, "user".id ASC`) {
		t.Fatalf("expected eligible player query to order by recent login first and stable name/id fallbacks, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `LIMIT 10`) || !strings.Contains(sql, `OFFSET 10`) {
		t.Fatalf("expected eligible player query to apply paging, got SQL: %s", sql)
	}
}

func TestBuildAdminAffiliationHistoryQueryAppliesCaptainCharacterAndTimeFilters(t *testing.T) {
	db := newDryRunPostgresDB(t)
	start := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 31, 23, 59, 59, 0, time.UTC)

	filter := AdminAffiliationHistoryFilter{
		CaptainSearch:       "Bee",
		PlayerSearch:        "Alpha",
		ChangeStartedAtFrom: &start,
		ChangeStartedAtTo:   &end,
	}

	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return buildAdminAffiliationHistoryQuery(tx, filter).Find(&[]model.NewbroCaptainAffiliation{})
	})

	if !strings.Contains(sql, `FROM "newbro_captain_affiliation"`) {
		t.Fatalf("expected newbro_captain_affiliation table, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `captain_user.nickname ILIKE`) {
		t.Fatalf("expected captain search filter, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `player_character.character_name ILIKE`) {
		t.Fatalf("expected player search filter, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `started_at >= `) {
		t.Fatalf("expected started_at lower-bound filter, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `started_at <= `) {
		t.Fatalf("expected started_at upper-bound filter, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `ended_at >= `) {
		t.Fatalf("expected ended_at lower-bound filter, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `ended_at <= `) {
		t.Fatalf("expected ended_at upper-bound filter, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `ORDER BY started_at DESC, id DESC`) {
		t.Fatalf("expected affiliation history to order by most recent start time, got SQL: %s", sql)
	}
}

func TestBuildAdminAffiliationHistoryQueryAppliesCaseInsensitiveCaptainAndPlayerSearch(t *testing.T) {
	db := newDryRunPostgresDB(t)

	filter := AdminAffiliationHistoryFilter{
		CaptainSearch: "Bee",
		PlayerSearch:  "Alpha",
	}

	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return buildAdminAffiliationHistoryQuery(tx, filter).Find(&[]model.NewbroCaptainAffiliation{})
	})

	if !strings.Contains(sql, `LEFT JOIN "user" AS captain_user`) {
		t.Fatalf("expected captain user join for search, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `LEFT JOIN eve_character AS captain_character`) {
		t.Fatalf("expected captain character join for search, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `captain_character.character_id = captain_user.primary_character_id`) {
		t.Fatalf("expected captain search to join current captain primary character, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `LEFT JOIN "user" AS player_user`) {
		t.Fatalf("expected player user join for search, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `LEFT JOIN eve_character AS player_character`) {
		t.Fatalf("expected player character join for search, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `captain_user.nickname ILIKE`) || !strings.Contains(sql, `captain_character.character_name ILIKE`) {
		t.Fatalf("expected case-insensitive captain nickname/name filter, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `player_user.nickname ILIKE`) || !strings.Contains(sql, `player_character.character_name ILIKE`) {
		t.Fatalf("expected case-insensitive player nickname/name filter, got SQL: %s", sql)
	}
}
