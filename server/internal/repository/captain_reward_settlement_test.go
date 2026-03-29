package repository

import (
	"amiya-eden/internal/model"
	"strings"
	"testing"

	"gorm.io/gorm"
)

func TestBuildCaptainRewardSettlementListQueryOrdersLatestFirstAndPaginates(t *testing.T) {
	db := newDryRunPostgresDB(t)

	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return buildCaptainRewardSettlementListQuery(tx, CaptainRewardSettlementFilter{}, 2, 200).
			Find(&[]model.CaptainRewardSettlement{})
	})

	if !strings.Contains(sql, `FROM "captain_reward_settlement"`) {
		t.Fatalf("expected captain_reward_settlement table, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `ORDER BY processed_at DESC, id DESC`) {
		t.Fatalf("expected processed_at desc ordering, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `LIMIT 200`) || !strings.Contains(sql, `OFFSET 200`) {
		t.Fatalf("expected page 2 with page size 200, got SQL: %s", sql)
	}
}

func TestBuildCaptainRewardSettlementListQueryCanScopeToCaptain(t *testing.T) {
	db := newDryRunPostgresDB(t)
	captainUserID := uint(3001)

	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return buildCaptainRewardSettlementListQuery(tx, CaptainRewardSettlementFilter{CaptainUserID: &captainUserID}, 1, 50).
			Find(&[]model.CaptainRewardSettlement{})
	})

	if !strings.Contains(sql, `captain_user_id =`) {
		t.Fatalf("expected captain_user_id filter, got SQL: %s", sql)
	}
}

func TestBuildCaptainRewardSettlementListQueryAppliesCaptainIdentityKeywordFilter(t *testing.T) {
	db := newDryRunPostgresDB(t)

	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return buildCaptainRewardSettlementListQuery(tx, CaptainRewardSettlementFilter{Keyword: "bee"}, 1, 20).
			Find(&[]model.CaptainRewardSettlement{})
	})

	if !strings.Contains(sql, `FROM "user" AS captain_user`) {
		t.Fatalf("expected keyword filter to check current captain nickname, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `FROM eve_character AS captain_character`) {
		t.Fatalf("expected keyword filter to check captain characters, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `LOWER(captain_user.nickname) LIKE`) {
		t.Fatalf("expected captain nickname keyword predicate, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `LOWER(captain_character.character_name) LIKE`) {
		t.Fatalf("expected captain character keyword predicate, got SQL: %s", sql)
	}
}
