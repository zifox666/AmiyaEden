package esi

import (
	"amiya-eden/internal/model"
	"strings"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const postgresMaxParameters = 65535

type capturedInsert struct {
	sql    string
	params int
}

// newDryRunDB 返回只生成 SQL、不连接数据库的 PostgreSQL 会话，并记录每条 INSERT
func newDryRunDB(t *testing.T) (*gorm.DB, *[]capturedInsert) {
	t.Helper()
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: &gorm.PreparedStmtDB{}}), &gorm.Config{
		DryRun:                 true,
		SkipDefaultTransaction: true,
		DisableAutomaticPing:   true,
	})
	if err != nil {
		t.Fatal(err)
	}

	inserts := &[]capturedInsert{}
	err = db.Callback().Create().After("gorm:create").Register("test:capture", func(tx *gorm.DB) {
		*inserts = append(*inserts, capturedInsert{sql: tx.Statement.SQL.String(), params: len(tx.Statement.Vars)})
	})
	if err != nil {
		t.Fatal(err)
	}
	return db, inserts
}

func assertInserts(t *testing.T, inserts []capturedInsert, wantStatements int) {
	t.Helper()
	if len(inserts) != wantStatements {
		t.Fatalf("statements = %d, want %d", len(inserts), wantStatements)
	}
	for i, insert := range inserts {
		if insert.params > postgresMaxParameters {
			t.Errorf("statement %d has %d params, limit is %d", i, insert.params, postgresMaxParameters)
		}
		if !strings.Contains(insert.sql, "ON CONFLICT DO NOTHING") {
			t.Errorf("statement %d must ignore duplicates: %.120s", i, insert.sql)
		}
	}
}

// 回归：角色 2121847691 流水过多时单条 INSERT 超出 65535 个参数
func TestInsertLargeJournalStaysUnderParameterLimit(t *testing.T) {
	db, inserts := newDryRunDB(t)

	rows := make([]model.EVECharacterWalletJournal, 5000)
	for i := range rows {
		rows[i].ID = int64(i + 1)
	}

	if err := insertIgnoringDuplicates(db, rows); err != nil {
		t.Fatal(err)
	}
	assertInserts(t, *inserts, 10)
}

// 回归：角色 2124135878 同一批数据内存在重复 transaction_id 时触发主键冲突
func TestInsertTransactionsIgnoresDuplicates(t *testing.T) {
	db, inserts := newDryRunDB(t)

	rows := []model.EVECharacterWalletTransaction{{TransactionID: 1}, {TransactionID: 1}, {TransactionID: 2}}

	if err := insertIgnoringDuplicates(db, rows); err != nil {
		t.Fatal(err)
	}
	assertInserts(t, *inserts, 1)
}

func TestInsertSkipsEmptyRows(t *testing.T) {
	db, inserts := newDryRunDB(t)

	if err := insertIgnoringDuplicates(db, []model.EVECharacterWalletJournal{}); err != nil {
		t.Fatal(err)
	}
	assertInserts(t, *inserts, 0)
}
