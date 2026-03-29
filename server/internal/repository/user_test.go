package repository

import (
	"amiya-eden/internal/model"
	"strings"
	"testing"

	"gorm.io/gorm"
)

func TestGetByIDForUpdateTxUsesRowLock(t *testing.T) {
	db := newDryRunPostgresDB(t)

	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return buildGetByIDForUpdateQuery(tx, 42)
	})

	if !strings.Contains(sql, `FOR UPDATE`) {
		t.Fatalf("expected row lock query to use FOR UPDATE, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `FROM "user"`) {
		t.Fatalf("expected user table query, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `"user"."id" = 42`) && !strings.Contains(sql, `"id" = 42`) {
		t.Fatalf("expected user id predicate, got SQL: %s", sql)
	}
}

func TestBuildUserListQueryAppliesNicknameQQAndCharacterKeywordFilter(t *testing.T) {
	db := newDryRunPostgresDB(t)

	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return buildUserListQuery(tx, UserFilter{Keyword: "amiya"}).Find(&[]model.User{})
	})

	if !strings.Contains(sql, `LOWER(nickname) LIKE`) {
		t.Fatalf("expected nickname keyword filter, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `LOWER(qq) LIKE`) {
		t.Fatalf("expected qq keyword filter, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `LOWER(character_name) LIKE`) {
		t.Fatalf("expected character keyword filter, got SQL: %s", sql)
	}
}

func TestBuildUserListQueryAppliesActiveRoleFilterWithoutLegacyFallback(t *testing.T) {
	db := newDryRunPostgresDB(t)

	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return buildUserListQuery(tx, UserFilter{Role: model.RoleCaptain}).Find(&[]model.User{})
	})

	if !strings.Contains(sql, `FROM user_role`) {
		t.Fatalf("expected role filter to query user_role association, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `user_role.role_code =`) {
		t.Fatalf("expected role filter to match role code in user_role, got SQL: %s", sql)
	}
	if strings.Contains(sql, `NOT EXISTS (SELECT 1 FROM user_role`) {
		t.Fatalf("expected role filter to avoid any legacy role fallback, got SQL: %s", sql)
	}
	if strings.Contains(sql, `AND role =`) || strings.Contains(sql, ` role = `) {
		t.Fatalf("expected role filter to avoid legacy user.role predicate, got SQL: %s", sql)
	}
}

func TestBuildUserListSelectQueryOrdersByRecentLoginDescending(t *testing.T) {
	db := newDryRunPostgresDB(t)

	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return buildUserListSelectQuery(tx, UserFilter{}, 2, 20).Find(&[]model.User{})
	})

	if !strings.Contains(sql, `ORDER BY last_login_at DESC NULLS LAST,id DESC`) && !strings.Contains(sql, `ORDER BY last_login_at DESC NULLS LAST, id DESC`) {
		t.Fatalf("expected user list query to order by latest login first, got SQL: %s", sql)
	}
	if !strings.Contains(sql, `LIMIT 20`) || !strings.Contains(sql, `OFFSET 20`) {
		t.Fatalf("expected user list query to apply paging, got SQL: %s", sql)
	}
}
