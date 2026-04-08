package jobs

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type fakeAutoSrpRunner struct {
	fleetIDs     []string
	err          error
	beforeReturn func()
}

func (f *fakeAutoSrpRunner) ProcessAutoSRP(fleetID string) error {
	f.fleetIDs = append(f.fleetIDs, fleetID)
	if f.beforeReturn != nil {
		f.beforeReturn()
	}
	return f.err
}

func newAutoSrpSchedulerTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	oldLogger := global.Logger
	global.Logger = zap.NewNop()
	t.Cleanup(func() { global.Logger = oldLogger })

	dsn := fmt.Sprintf("file:auto_srp_scheduler_test_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.Fleet{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}

// TestFleetAutoSrpSchedulerScheduleAfterPAPIssuedPreservesExistingDBSchedule verifies
// that ScheduleAfterPAPIssued does not accidentally clear a pre-existing
// auto_srp_scheduled_for value that was persisted by the fleet service.
// ScheduleAfterPAPIssued itself only sets an in-memory timer; the DB write
// happens inside the fleet service transaction.
func TestFleetAutoSrpSchedulerScheduleAfterPAPIssuedPreservesExistingDBSchedule(t *testing.T) {
	db := newAutoSrpSchedulerTestDB(t)
	oldDB := global.DB
	global.DB = db
	t.Cleanup(func() { global.DB = oldDB })

	if err := db.Create(&model.Fleet{ID: "fleet-1", AutoSrpMode: model.FleetAutoSrpAutoApprove}).Error; err != nil {
		t.Fatalf("create fleet: %v", err)
	}

	runner := &fakeAutoSrpRunner{}
	scheduler := newFleetAutoSrpScheduler(repository.NewFleetRepository(), runner)
	issuedAt := time.Now().UTC()

	scheduledFor := model.NormalizeFleetAutoSrpScheduledFor(issuedAt.Add(model.FleetAutoSrpDelay))
	if err := db.Model(&model.Fleet{}).Where("id = ?", "fleet-1").Update("auto_srp_scheduled_for", scheduledFor).Error; err != nil {
		t.Fatalf("persist schedule: %v", err)
	}
	if err := scheduler.ScheduleAfterPAPIssued("fleet-1", issuedAt); err != nil {
		t.Fatalf("schedule after PAP: %v", err)
	}

	var stored model.Fleet
	if err := db.Where("id = ?", "fleet-1").First(&stored).Error; err != nil {
		t.Fatalf("reload fleet: %v", err)
	}
	if stored.AutoSrpScheduledFor == nil {
		t.Fatalf("expected pre-existing DB schedule to survive")
	}
	want := model.NormalizeFleetAutoSrpScheduledFor(issuedAt.Add(model.FleetAutoSrpDelay))
	if !stored.AutoSrpScheduledFor.Equal(want) {
		t.Fatalf("auto_srp_scheduled_for = %v, want %v", stored.AutoSrpScheduledFor, want)
	}
	if len(runner.fleetIDs) != 0 {
		t.Fatalf("expected no auto SRP run during scheduling, got %v", runner.fleetIDs)
	}

	scheduler.mu.Lock()
	timer := scheduler.timers["fleet-1"]
	scheduler.mu.Unlock()
	if timer == nil {
		t.Fatalf("expected in-memory timer to be created")
	}
}

func TestFleetAutoSrpSchedulerScheduleAfterPAPIssuedSkipsDisabledFleet(t *testing.T) {
	db := newAutoSrpSchedulerTestDB(t)
	oldDB := global.DB
	global.DB = db
	t.Cleanup(func() { global.DB = oldDB })

	if err := db.Create(&model.Fleet{ID: "fleet-1", AutoSrpMode: model.FleetAutoSrpDisabled}).Error; err != nil {
		t.Fatalf("create fleet: %v", err)
	}

	runner := &fakeAutoSrpRunner{}
	scheduler := newFleetAutoSrpScheduler(repository.NewFleetRepository(), runner)
	issuedAt := time.Date(2026, time.April, 8, 12, 0, 0, 0, time.UTC)

	if err := scheduler.ScheduleAfterPAPIssued("fleet-1", issuedAt); err != nil {
		t.Fatalf("schedule after PAP: %v", err)
	}

	var stored model.Fleet
	if err := db.Where("id = ?", "fleet-1").First(&stored).Error; err != nil {
		t.Fatalf("reload fleet: %v", err)
	}
	if stored.AutoSrpScheduledFor != nil {
		t.Fatalf("expected disabled fleet to keep no schedule, got %v", stored.AutoSrpScheduledFor)
	}
	if _, ok := scheduler.timers["fleet-1"]; ok {
		t.Fatalf("expected disabled fleet to avoid creating an in-memory timer")
	}
}

func TestFleetAutoSrpSchedulerRunScheduledFleetSkipsStaleSchedule(t *testing.T) {
	db := newAutoSrpSchedulerTestDB(t)
	oldDB := global.DB
	global.DB = db
	t.Cleanup(func() { global.DB = oldDB })

	requestedAt := time.Date(2026, time.April, 8, 14, 0, 0, 0, time.UTC)
	actualAt := requestedAt.Add(30 * time.Minute)
	if err := db.Create(&model.Fleet{
		ID:                  "fleet-1",
		AutoSrpMode:         model.FleetAutoSrpAutoApprove,
		AutoSrpScheduledFor: &actualAt,
	}).Error; err != nil {
		t.Fatalf("create fleet: %v", err)
	}

	runner := &fakeAutoSrpRunner{}
	scheduler := newFleetAutoSrpScheduler(repository.NewFleetRepository(), runner)
	scheduler.runScheduledFleet("fleet-1", requestedAt)

	if len(runner.fleetIDs) != 0 {
		t.Fatalf("expected stale schedule to be ignored, got %v", runner.fleetIDs)
	}

	var stored model.Fleet
	if err := db.Where("id = ?", "fleet-1").First(&stored).Error; err != nil {
		t.Fatalf("reload fleet: %v", err)
	}
	if stored.AutoSrpScheduledFor == nil || !stored.AutoSrpScheduledFor.Equal(actualAt) {
		t.Fatalf("expected newer schedule to remain, got %v", stored.AutoSrpScheduledFor)
	}
}

func TestFleetAutoSrpSchedulerRunScheduledFleetRunsAndClearsSchedule(t *testing.T) {
	db := newAutoSrpSchedulerTestDB(t)
	oldDB := global.DB
	global.DB = db
	t.Cleanup(func() { global.DB = oldDB })

	scheduledAt := time.Date(2026, time.April, 8, 14, 0, 0, 0, time.UTC)
	if err := db.Create(&model.Fleet{
		ID:                  "fleet-1",
		AutoSrpMode:         model.FleetAutoSrpAutoApprove,
		AutoSrpScheduledFor: &scheduledAt,
	}).Error; err != nil {
		t.Fatalf("create fleet: %v", err)
	}

	runner := &fakeAutoSrpRunner{}
	scheduler := newFleetAutoSrpScheduler(repository.NewFleetRepository(), runner)
	scheduler.runScheduledFleet("fleet-1", scheduledAt)

	if len(runner.fleetIDs) != 1 || runner.fleetIDs[0] != "fleet-1" {
		t.Fatalf("expected one auto SRP run for fleet-1, got %v", runner.fleetIDs)
	}

	var stored model.Fleet
	if err := db.Where("id = ?", "fleet-1").First(&stored).Error; err != nil {
		t.Fatalf("reload fleet: %v", err)
	}
	if stored.AutoSrpScheduledFor != nil {
		t.Fatalf("expected schedule to be cleared after run, got %v", stored.AutoSrpScheduledFor)
	}
}

func TestFleetAutoSrpSchedulerRunScheduledFleetClearsScheduleOnError(t *testing.T) {
	db := newAutoSrpSchedulerTestDB(t)
	oldDB := global.DB
	global.DB = db
	t.Cleanup(func() { global.DB = oldDB })

	scheduledAt := time.Date(2026, time.April, 8, 14, 0, 0, 0, time.UTC)
	if err := db.Create(&model.Fleet{
		ID:                  "fleet-1",
		AutoSrpMode:         model.FleetAutoSrpAutoApprove,
		AutoSrpScheduledFor: &scheduledAt,
	}).Error; err != nil {
		t.Fatalf("create fleet: %v", err)
	}

	runner := &fakeAutoSrpRunner{err: fmt.Errorf("boom")}
	scheduler := newFleetAutoSrpScheduler(repository.NewFleetRepository(), runner)
	scheduler.runScheduledFleet("fleet-1", scheduledAt)

	var stored model.Fleet
	if err := db.Where("id = ?", "fleet-1").First(&stored).Error; err != nil {
		t.Fatalf("reload fleet: %v", err)
	}
	if stored.AutoSrpScheduledFor != nil {
		t.Fatalf("expected one-off schedule to be consumed even on failure, got %v", stored.AutoSrpScheduledFor)
	}
	if _, ok := scheduler.timers["fleet-1"]; ok {
		t.Fatalf("expected no automatic retry timer after failure")
	}
	if len(runner.fleetIDs) != 1 {
		t.Fatalf("expected exactly one failed attempt, got %v", runner.fleetIDs)
	}
}

func TestFleetAutoSrpSchedulerRunScheduledFleetDoesNotClearNewerSchedule(t *testing.T) {
	db := newAutoSrpSchedulerTestDB(t)
	oldDB := global.DB
	global.DB = db
	t.Cleanup(func() { global.DB = oldDB })

	oldSchedule := time.Date(2026, time.April, 8, 14, 0, 0, 0, time.UTC)
	newSchedule := oldSchedule.Add(2 * time.Hour)
	if err := db.Create(&model.Fleet{
		ID:                  "fleet-1",
		AutoSrpMode:         model.FleetAutoSrpAutoApprove,
		AutoSrpScheduledFor: &oldSchedule,
	}).Error; err != nil {
		t.Fatalf("create fleet: %v", err)
	}

	runner := &fakeAutoSrpRunner{beforeReturn: func() {
		if err := repository.NewFleetRepository().SetAutoSrpScheduledFor("fleet-1", &newSchedule); err != nil {
			t.Fatalf("reschedule fleet: %v", err)
		}
	}}
	scheduler := newFleetAutoSrpScheduler(repository.NewFleetRepository(), runner)
	scheduler.runScheduledFleet("fleet-1", oldSchedule)

	var stored model.Fleet
	if err := db.Where("id = ?", "fleet-1").First(&stored).Error; err != nil {
		t.Fatalf("reload fleet: %v", err)
	}
	if stored.AutoSrpScheduledFor == nil || !stored.AutoSrpScheduledFor.Equal(newSchedule) {
		t.Fatalf("expected newer schedule to survive old run, got %v", stored.AutoSrpScheduledFor)
	}
}
