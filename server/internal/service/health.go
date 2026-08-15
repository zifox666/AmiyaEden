package service

import (
	"amiya-eden/global"
	"context"
	"time"
)

type HealthStatus struct {
	DataDB bool `json:"data_db"`
	SDEDB  bool `json:"sde_db"`
	Redis  bool `json:"redis"`
}

func (s HealthStatus) Healthy() bool { return s.DataDB && s.SDEDB && s.Redis }

type HealthService struct{}

func NewHealthService() *HealthService { return &HealthService{} }

func (s *HealthService) Check() HealthStatus {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	status := HealthStatus{}
	if global.DB != nil {
		if db, err := global.DB.DB(); err == nil {
			status.DataDB = db.PingContext(ctx) == nil
		}
	}
	if global.SDEDB != nil {
		if db, err := global.SDEDB.DB(); err == nil {
			status.SDEDB = db.PingContext(ctx) == nil
		}
	}
	if global.Redis != nil {
		status.Redis = global.Redis.Ping(ctx).Err() == nil
	}
	return status
}
