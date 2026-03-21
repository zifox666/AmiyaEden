package repository

import (
	"amiya-eden/internal/model"
	"database/sql"
	"strings"
	"sync"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func newDryRunPostgresDB(t *testing.T) *gorm.DB {
	t.Helper()

	sqlDB, err := sql.Open("pgx", "postgres://amiya:test@127.0.0.1:5432/amiya_test")
	if err != nil {
		t.Fatalf("open sql db: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		DryRun:               true,
		DisableAutomaticPing: true,
	})
	if err != nil {
		t.Fatalf("open gorm dry-run db: %v", err)
	}

	return db
}

func TestBuildFleetListQueriesUseQualifiedDeletedAtAndNicknameFallback(t *testing.T) {
	db := newDryRunPostgresDB(t)
	fcUserID := uint(42)
	filter := FleetFilter{
		Importance: model.FleetImportanceCTA,
		FCUserID:   &fcUserID,
	}

	baseQuery := buildFleetListBaseQuery(db, filter)
	countQuery := baseQuery.Session(&gorm.Session{DryRun: true}).Count(new(int64))
	countSQL := countQuery.Statement.SQL.String()

	if !strings.Contains(countSQL, `fleet.deleted_at IS NULL`) {
		t.Fatalf("expected count query to qualify deleted_at, got SQL: %s", countSQL)
	}
	if strings.Contains(countSQL, `WHERE deleted_at IS NULL`) {
		t.Fatalf("expected count query to avoid ambiguous deleted_at filter, got SQL: %s", countSQL)
	}

	findQuery := buildFleetListSelectQuery(baseQuery).
		Session(&gorm.Session{DryRun: true}).
		Order("fleet.start_at DESC").
		Offset(0).
		Limit(20).
		Find(&[]model.FleetListItem{})
	findSQL := findQuery.Statement.SQL.String()

	if !strings.Contains(findSQL, `LEFT JOIN "user" ON "user".id = fleet.fc_user_id`) {
		t.Fatalf("expected fleet list query to join user table, got SQL: %s", findSQL)
	}
	if !strings.Contains(findSQL, `COALESCE(NULLIF("user".nickname, ''), fleet.fc_character_name) AS fc_display_name`) {
		t.Fatalf("expected fleet list query to prefer nickname then fall back to character name, got SQL: %s", findSQL)
	}
	if !strings.Contains(findSQL, `fleet.deleted_at IS NULL`) {
		t.Fatalf("expected joined fleet list query to retain qualified deleted_at filter, got SQL: %s", findSQL)
	}
}

func TestFleetListItemMapsFCDisplayNameForScan(t *testing.T) {
	parsedSchema, err := schema.Parse(&model.FleetListItem{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatalf("parse schema: %v", err)
	}

	field := parsedSchema.LookUpField("FCDisplayName")
	if field == nil {
		t.Fatal("expected FleetListItem schema to include FCDisplayName")
	}
	if field.DBName != "fc_display_name" {
		t.Fatalf("expected FCDisplayName DB name to be fc_display_name, got %q", field.DBName)
	}
	if !field.Readable {
		t.Fatal("expected FCDisplayName to remain readable for query scans")
	}
}
