package repository

import (
	"amiya-eden/internal/model"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"
)

func TestBuildPendingBatchPayoutApplicationsQueryUsesUserScopedLocking(t *testing.T) {
	db := newDryRunPostgresDB(t)

	tx := buildPendingBatchPayoutApplicationsQuery(db, 42).
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
