package jobs

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"amiya-eden/internal/service"
	"sync"
	"time"

	"go.uber.org/zap"
)

type autoSrpRunner interface {
	ProcessAutoSRP(fleetID string) error
}

type fleetAutoSrpScheduler struct {
	repo   *repository.FleetRepository
	runner autoSrpRunner

	mu           sync.Mutex
	timers       map[string]*time.Timer
	scheduledFor map[string]time.Time
}

func newFleetAutoSrpScheduler(repo *repository.FleetRepository, runner autoSrpRunner) *fleetAutoSrpScheduler {
	return &fleetAutoSrpScheduler{
		repo:         repo,
		runner:       runner,
		timers:       make(map[string]*time.Timer),
		scheduledFor: make(map[string]time.Time),
	}
}

func registerAutoSrpScheduler() {
	scheduler := newFleetAutoSrpScheduler(repository.NewFleetRepository(), service.NewAutoSrpService())
	if err := scheduler.Restore(); err != nil {
		global.Logger.Error("恢复自动 SRP 调度失败", zap.Error(err))
	}

	service.FleetAutoSRPFunc = func(fleetID string, issuedAt time.Time) {
		if err := scheduler.ScheduleAfterPAPIssued(fleetID, issuedAt); err != nil {
			global.Logger.Warn("[AutoSRP] PAP 后调度失败",
				zap.String("fleet_id", fleetID),
				zap.Time("issued_at", issuedAt),
				zap.Error(err),
			)
		}
	}
}

func (s *fleetAutoSrpScheduler) Restore() error {
	fleets, err := s.repo.ListWithAutoSrpScheduled()
	if err != nil {
		return err
	}

	for i := range fleets {
		fleet := fleets[i]
		if fleet.AutoSrpScheduledFor == nil {
			continue
		}
		s.scheduleTimer(fleet.ID, *fleet.AutoSrpScheduledFor)
	}

	return nil
}

func (s *fleetAutoSrpScheduler) ScheduleAfterPAPIssued(fleetID string, issuedAt time.Time) error {
	fleet, err := s.repo.GetByID(fleetID)
	if err != nil {
		return err
	}
	if fleet.AutoSrpMode == model.FleetAutoSrpDisabled {
		return nil
	}

	scheduledFor := model.NormalizeFleetAutoSrpScheduledFor(issuedAt.Add(model.FleetAutoSrpDelay))
	s.scheduleTimer(fleetID, scheduledFor)
	return nil
}

func (s *fleetAutoSrpScheduler) scheduleTimer(fleetID string, scheduledFor time.Time) {
	delay := time.Until(scheduledFor)
	if delay < 0 {
		delay = 0
	}

	s.mu.Lock()
	if timer := s.timers[fleetID]; timer != nil {
		timer.Stop()
	}
	s.scheduledFor[fleetID] = scheduledFor
	s.timers[fleetID] = time.AfterFunc(delay, func() {
		s.runScheduledFleet(fleetID, scheduledFor)
	})
	s.mu.Unlock()
}

func (s *fleetAutoSrpScheduler) runScheduledFleet(fleetID string, scheduledFor time.Time) {
	fleet, err := s.repo.GetByID(fleetID)
	if err != nil {
		global.Logger.Warn("[AutoSRP] 加载已调度舰队失败",
			zap.String("fleet_id", fleetID),
			zap.Time("scheduled_for", scheduledFor),
			zap.Error(err),
		)
		s.clearTimer(fleetID, scheduledFor)
		return
	}

	if fleet.AutoSrpMode == model.FleetAutoSrpDisabled {
		_ = s.repo.SetAutoSrpScheduledFor(fleetID, nil)
		s.clearTimer(fleetID, scheduledFor)
		return
	}
	if fleet.AutoSrpScheduledFor == nil || !fleet.AutoSrpScheduledFor.Equal(scheduledFor) {
		s.clearTimer(fleetID, scheduledFor)
		return
	}
	claimed, err := s.repo.ClaimAutoSrpScheduledForIfMatch(fleetID, scheduledFor)
	if err != nil {
		global.Logger.Warn("[AutoSRP] 领取调度任务失败",
			zap.String("fleet_id", fleetID),
			zap.Time("scheduled_for", scheduledFor),
			zap.Error(err),
		)
		s.clearTimer(fleetID, scheduledFor)
		return
	}
	if !claimed {
		s.clearTimer(fleetID, scheduledFor)
		return
	}

	if err := s.runner.ProcessAutoSRP(fleetID); err != nil {
		global.Logger.Warn("[AutoSRP] 执行已调度任务失败",
			zap.String("fleet_id", fleetID),
			zap.Time("scheduled_for", scheduledFor),
			zap.Error(err),
		)
	}
	s.clearTimer(fleetID, scheduledFor)
}

func (s *fleetAutoSrpScheduler) clearTimer(fleetID string, scheduledFor time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if current, ok := s.scheduledFor[fleetID]; !ok || !current.Equal(scheduledFor) {
		return
	}
	delete(s.scheduledFor, fleetID)
	delete(s.timers, fleetID)
}
