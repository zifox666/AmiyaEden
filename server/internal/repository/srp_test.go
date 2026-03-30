package repository

import (
	"amiya-eden/internal/model"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"
)

func TestBuildApprovedUnpaidBatchPayoutApplicationsQueryUsesUserScopedLocking(t *testing.T) {
	db := newDryRunPostgresDB(t)

	tx := buildApprovedUnpaidBatchPayoutApplicationsQuery(db, 42).
		Session(&gorm.Session{DryRun: true}).
		Find(&[]model.SrpApplication{})
	sql := tx.Statement.SQL.String()

	if !strings.Contains(sql, `FROM "srp_application"`) {
		t.Fatalf("expected srp_application select, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `user_id = $1`) {
		t.Fatalf("expected user-scoped filter, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `payout_status = $2`) || !strings.Contains(sql, `review_status = $3`) {
		t.Fatalf("expected payout/review status filters, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `FOR UPDATE`) {
		t.Fatalf("expected row locking for batch payout selection, got SQL: %s", sql)
	}
}

func TestBuildPendingBadgeSrpCountQueryUsesPendingApprovalStatuses(t *testing.T) {
	db := newDryRunPostgresDB(t)

	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return buildPendingBadgeSrpCountQuery(tx).Count(new(int64))
	})

	if !strings.Contains(sql, `FROM "srp_application"`) {
		t.Fatalf("expected srp_application count query, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `review_status IN (`) {
		t.Fatalf("expected pending review scope on badge count query, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `payout_status =`) {
		t.Fatalf("expected unpaid scope on badge count query, got SQL: %s", sql)
	}
}

func TestBuildBatchPayoutApplicationsUpdateTargetsSelectedApplicationIDs(t *testing.T) {
	db := newDryRunPostgresDB(t)
	paidAt := time.Unix(1_700_000_000, 0).UTC()

	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return buildBatchPayoutApplicationsUpdateQuery(tx, []uint{7, 9}, 99, paidAt)
	})

	if !strings.Contains(sql, `UPDATE "srp_application"`) {
		t.Fatalf("expected srp_application update, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `id IN (`) {
		t.Fatalf("expected ID-scoped update, got SQL: %s", sql)
	}
	if strings.Contains(sql, `user_id =`) {
		t.Fatalf("expected update to avoid broad user-scoped predicate, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `payout_status =`) || !strings.Contains(sql, `review_status =`) {
		t.Fatalf("expected payout/review status guard on update, got SQL: %s", sql)
	}
}

func TestSummarizeBatchPayoutApplicationsUsesExactSelectedRows(t *testing.T) {
	summary, ids := summarizeBatchPayoutApplications(42, []model.SrpApplication{
		{ID: 7, FinalAmount: 12.5},
		{ID: 9, FinalAmount: 30},
	})

	if summary.UserID != 42 {
		t.Fatalf("expected user ID 42, got %d", summary.UserID)
	}
	if summary.ApplicationCount != 2 {
		t.Fatalf("expected 2 applications, got %d", summary.ApplicationCount)
	}
	if summary.TotalAmount != 42.5 {
		t.Fatalf("expected total amount 42.5, got %v", summary.TotalAmount)
	}
	if len(ids) != 2 || ids[0] != 7 || ids[1] != 9 {
		t.Fatalf("expected selected IDs [7 9], got %v", ids)
	}
}

func TestBuildSrpApplicationListQueryAppliesPendingTabScope(t *testing.T) {
	db := newDryRunPostgresDB(t)

	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return buildSrpApplicationListQuery(tx, SrpApplicationFilter{Tab: SrpTabPending}).
			Find(&[]model.SrpApplication{})
	})

	if !strings.Contains(sql, `FROM "srp_application"`) {
		t.Fatalf("expected srp_application select, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `review_status IN (`) {
		t.Fatalf("expected pending tab review scope, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `payout_status =`) {
		t.Fatalf("expected pending tab payout scope, got SQL: %s", sql)
	}
}

func TestBuildSrpApplicationListQueryAppliesHistoryTabScope(t *testing.T) {
	db := newDryRunPostgresDB(t)

	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return buildSrpApplicationListQuery(tx, SrpApplicationFilter{Tab: SrpTabHistory}).
			Find(&[]model.SrpApplication{})
	})

	if !strings.Contains(sql, `FROM "srp_application"`) {
		t.Fatalf("expected srp_application select, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `payout_status =`) || !strings.Contains(sql, `OR review_status =`) {
		t.Fatalf("expected history tab to include paid or rejected scope, got SQL: %s", sql)
	}
}

func TestBuildSrpApplicationListQueryAppliesCharacterAndNicknameKeywordFilter(t *testing.T) {
	db := newDryRunPostgresDB(t)

	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return buildSrpApplicationListQuery(tx, SrpApplicationFilter{Keyword: "bee"}).Find(&[]model.SrpApplication{})
	})

	if !strings.Contains(sql, `FROM "user" AS applicant_user`) {
		t.Fatalf("expected keyword filter to query current applicant nickname, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `LOWER(applicant_user.nickname) LIKE`) {
		t.Fatalf("expected applicant nickname keyword predicate, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `LOWER(character_name) LIKE`) {
		t.Fatalf("expected application character keyword predicate, got SQL: %s", sql)
	}
}
