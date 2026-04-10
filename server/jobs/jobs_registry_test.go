package jobs

import (
	"amiya-eden/global"
	"amiya-eden/internal/taskregistry"
	"amiya-eden/pkg/eve/esi"
	"testing"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func TestRegisterAllRegistersExpectedTaskDefinitions(t *testing.T) {
	oldLogger := global.Logger
	oldDB := global.DB
	oldQueueFactory := newESIQueueForJobs
	oldStartupRun := startInitialESIQueueRun
	oldQueue := esiQueue

	t.Cleanup(func() {
		global.Logger = oldLogger
		global.DB = oldDB
		newESIQueueForJobs = oldQueueFactory
		startInitialESIQueueRun = oldStartupRun
		esiQueue = oldQueue
	})

	global.Logger = zap.NewNop()
	global.DB = newAutoSrpSchedulerTestDB(t)
	newESIQueueForJobs = func() *esi.Queue {
		return &esi.Queue{}
	}
	startInitialESIQueueRun = func(queue *esi.Queue) {}

	reg := taskregistry.New()
	RegisterAll(reg)

	want := map[string]struct {
		category taskregistry.TaskCategory
		taskType taskregistry.TaskType
		cron     string
		hasRun   bool
	}{
		"esi_refresh":                    {category: taskregistry.TaskCategoryESI, taskType: taskregistry.TaskTypeRecurring, cron: "0 */5 * * * *", hasRun: true},
		"alliance_pap_hourly":            {category: taskregistry.TaskCategoryOperation, taskType: taskregistry.TaskTypeRecurring, cron: "0 0 * * * *", hasRun: true},
		"alliance_pap_archive":           {category: taskregistry.TaskCategoryOperation, taskType: taskregistry.TaskTypeRecurring, cron: "0 0 1 1 * *", hasRun: true},
		"auto_role_sync":                 {category: taskregistry.TaskCategorySystem, taskType: taskregistry.TaskTypeRecurring, cron: "0 2/10 * * * *", hasRun: true},
		"task_execution_history_cleanup": {category: taskregistry.TaskCategorySystem, taskType: taskregistry.TaskTypeRecurring, cron: "0 0 4 1 * *", hasRun: true},
		"corp_access_check":              {category: taskregistry.TaskCategorySystem, taskType: taskregistry.TaskTypeRecurring, cron: "0 0/5 * * * *", hasRun: true},
		"captain_attribution_sync":       {category: taskregistry.TaskCategoryOperation, taskType: taskregistry.TaskTypeRecurring, cron: "@every 13h", hasRun: true},
		"captain_reward_processing":      {category: taskregistry.TaskCategoryOperation, taskType: taskregistry.TaskTypeRecurring, cron: "@every 100h", hasRun: true},
		"mentor_reward":                  {category: taskregistry.TaskCategoryOperation, taskType: taskregistry.TaskTypeRecurring, cron: "0 0 3 * * *", hasRun: true},
		"auto_srp":                       {category: taskregistry.TaskCategoryOperation, taskType: taskregistry.TaskTypeTriggered, cron: "", hasRun: false},
	}

	all := reg.All()
	if len(all) != len(want) {
		t.Fatalf("registered task count = %d, want %d", len(all), len(want))
	}

	for name, wantDef := range want {
		got, ok := reg.Get(name)
		if !ok {
			t.Fatalf("expected task %q to be registered", name)
		}
		if got.Category != wantDef.category {
			t.Fatalf("task %q category = %q, want %q", name, got.Category, wantDef.category)
		}
		if got.Type != wantDef.taskType {
			t.Fatalf("task %q type = %q, want %q", name, got.Type, wantDef.taskType)
		}
		if got.DefaultCron != wantDef.cron {
			t.Fatalf("task %q cron = %q, want %q", name, got.DefaultCron, wantDef.cron)
		}
		if (got.RunFunc != nil) != wantDef.hasRun {
			t.Fatalf("task %q has run func = %t, want %t", name, got.RunFunc != nil, wantDef.hasRun)
		}
	}
}

func TestRegisterAutoSrpTaskRegistersTriggeredTask(t *testing.T) {
	oldLogger := global.Logger
	oldDB := global.DB

	t.Cleanup(func() {
		global.Logger = oldLogger
		global.DB = oldDB
	})

	global.Logger = zap.NewNop()
	global.DB = newAutoSrpSchedulerTestDB(t)

	reg := taskregistry.New()
	registerAutoSrpTask(reg)

	def, ok := reg.Get("auto_srp")
	if !ok {
		t.Fatal("expected auto_srp task definition to be registered")
	}
	if def.Category != taskregistry.TaskCategoryOperation {
		t.Fatalf("category = %q, want %q", def.Category, taskregistry.TaskCategoryOperation)
	}
	if def.Type != taskregistry.TaskTypeTriggered {
		t.Fatalf("type = %q, want %q", def.Type, taskregistry.TaskTypeTriggered)
	}
	if def.DefaultCron != "" {
		t.Fatalf("default cron = %q, want empty", def.DefaultCron)
	}
	if def.RunFunc != nil {
		t.Fatal("expected auto_srp to remain event-driven without a manual run function")
	}
	if global.DB == (*gorm.DB)(nil) {
		t.Fatal("expected test database to remain configured")
	}
}
