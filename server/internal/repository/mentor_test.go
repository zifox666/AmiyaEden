package repository

import (
	"amiya-eden/internal/model"
	"strings"
	"testing"

	"gorm.io/gorm"
)

func TestBuildMentorRelationshipListSelectQueryOrdersByMenteeRecentLoginDesc(t *testing.T) {
	db := newDryRunPostgresDB(t)

	baseQuery := buildMentorRelationshipListQuery(db, 42, []string{model.MentorRelationStatusActive})

	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return buildMentorRelationshipListSelectQuery(baseQuery, 2, 20).
			Find(&[]model.MentorMenteeRelationship{})
	})

	if !strings.Contains(sql, `LEFT JOIN "user" AS mentee_user ON mentee_user.id = mentor_mentee_relationship.mentee_user_id`) {
		t.Fatalf("expected mentee list query to join mentee user for last login ordering, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `LEFT JOIN eve_character AS mentee_primary_character ON mentee_primary_character.character_id = mentee_user.primary_character_id`) {
		t.Fatalf("expected mentee list query to join mentee primary character for stable ordering, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `mentor_mentee_relationship.status IN`) {
		t.Fatalf("expected mentee list query to qualify mentor relationship status filter, got SQL: %s", sql)
	}
	if strings.Contains(sql, ` AND status IN `) {
		t.Fatalf("expected mentee list query to avoid ambiguous unqualified status filter, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `ORDER BY mentee_user.last_login_at DESC NULLS LAST,mentee_primary_character.character_name ASC,mentor_mentee_relationship.id ASC`) &&
		!strings.Contains(sql, `ORDER BY mentee_user.last_login_at DESC NULLS LAST, mentee_primary_character.character_name ASC, mentor_mentee_relationship.id ASC`) {
		t.Fatalf("expected mentee list query to order by recent login first with stable character and id fallbacks, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `LIMIT 20`) || !strings.Contains(sql, `OFFSET 20`) {
		t.Fatalf("expected mentee list query to apply paging, got SQL: %s", sql)
	}
}

func TestBuildPendingMentorRelationshipListSelectQueryOrdersByMenteeRecentLoginDesc(t *testing.T) {
	db := newDryRunPostgresDB(t)

	baseQuery := buildMentorRelationshipListQuery(db, 42, []string{model.MentorRelationStatusPending})

	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return buildPendingMentorRelationshipListSelectQuery(baseQuery).
			Find(&[]model.MentorMenteeRelationship{})
	})

	if !strings.Contains(sql, `LEFT JOIN "user" AS mentee_user ON mentee_user.id = mentor_mentee_relationship.mentee_user_id`) {
		t.Fatalf("expected pending mentee query to join mentee user for last login ordering, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `LEFT JOIN eve_character AS mentee_primary_character ON mentee_primary_character.character_id = mentee_user.primary_character_id`) {
		t.Fatalf("expected pending mentee query to join mentee primary character for stable ordering, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `mentor_mentee_relationship.status IN`) {
		t.Fatalf("expected pending mentee query to qualify mentor relationship status filter, got SQL: %s", sql)
	}
	if strings.Contains(sql, ` AND status IN `) {
		t.Fatalf("expected pending mentee query to avoid ambiguous unqualified status filter, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `ORDER BY mentee_user.last_login_at DESC NULLS LAST,mentee_primary_character.character_name ASC,mentor_mentee_relationship.id ASC`) &&
		!strings.Contains(sql, `ORDER BY mentee_user.last_login_at DESC NULLS LAST, mentee_primary_character.character_name ASC, mentor_mentee_relationship.id ASC`) {
		t.Fatalf("expected pending mentee query to order by recent login first with stable character and id fallbacks, got SQL: %s", sql)
	}
}

func TestCountPendingByMentorUserIDScopesToMentorAndPendingStatus(t *testing.T) {
	db := newDryRunPostgresDB(t)

	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		var count int64
		return tx.Model(&model.MentorMenteeRelationship{}).
			Where("mentor_mentee_relationship.mentor_user_id = ?", 42).
			Where("mentor_mentee_relationship.status = ?", model.MentorRelationStatusPending).
			Count(&count)
	})

	if !strings.Contains(sql, `mentor_mentee_relationship.mentor_user_id = 42`) {
		t.Fatalf("expected pending mentor badge query to scope by mentor user id, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `mentor_mentee_relationship.status = 'pending'`) {
		t.Fatalf("expected pending mentor badge query to scope to pending status, got SQL: %s", sql)
	}
	if strings.Contains(sql, `status = 'active'`) {
		t.Fatalf("expected pending mentor badge query to exclude active rows, got SQL: %s", sql)
	}
}

func TestListDistributedRewardAmountsByRelationshipIDsQuerySumsRewardAmount(t *testing.T) {
	db := newDryRunPostgresDB(t)

	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return buildMentorRewardAmountSummaryQuery(tx, []uint{11, 22}).
			Scan(&[]struct{}{})
	})

	if !strings.Contains(sql, `FROM "mentor_reward_distribution"`) {
		t.Fatalf("expected mentor reward distribution table, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `SUM(reward_amount) AS total_reward_amount`) {
		t.Fatalf("expected reward amount aggregation, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `GROUP BY "relationship_id"`) && !strings.Contains(sql, `GROUP BY relationship_id`) {
		t.Fatalf("expected grouping by relationship_id, got SQL: %s", sql)
	}
}
