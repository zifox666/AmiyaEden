package esi

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"amiya-eden/pkg/utils"
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm/clause"
)

func init() {
	Register(&StructureTask{})
}

type StructureTask struct{}

func (t *StructureTask) Name() string        { return "eve_structures" }
func (t *StructureTask) Description() string { return "EVE 建筑信息" }
func (t *StructureTask) Priority() Priority  { return PriorityLow }

func (t *StructureTask) Interval() RefreshInterval {
	return RefreshInterval{
		Active:   3 * 24 * time.Hour,
		Inactive: 7 * 24 * time.Hour,
	}
}

func (t *StructureTask) RequiredScopes() []TaskScope {
	return []TaskScope{
		{Scope: "esi-universe.read_structures.v1", Description: "读取建筑信息"},
		{Scope: "esi-corporations.read_structures.v1", Description: "读取军团建筑信息"},
	}
}

// corpStructureESIResp ESI 返回的军团建筑原始数据
type corpStructureESIResp struct {
	CorporationID      int64  `json:"corporation_id"`
	FuelExpires        string `json:"fuel_expires"`
	Name               string `json:"name"`
	NextReinforceApply string `json:"next_reinforce_apply"`
	NextReinforceHour  int    `json:"next_reinforce_hour"`
	ProfileID          int64  `json:"profile_id"`
	ReinforceHour      int    `json:"reinforce_hour"`
	Services           []struct {
		Name  string `json:"name"`
		State string `json:"state"`
	} `json:"services"`
	State           string `json:"state"`
	StateTimerEnd   string `json:"state_timer_end"`
	StateTimerStart string `json:"state_timer_start"`
	StructureID     int64  `json:"structure_id"`
	SystemID        int64  `json:"system_id"`
	TypeID          int64  `json:"type_id"`
	UnanchorsAt     string `json:"unanchors_at"`
}

// eveStructureDetail ESI 返回的建筑详情
type eveStructureDetail struct {
	Name     string `json:"name"`
	OwnerID  int64  `json:"owner_id"`
	Position struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
		Z float64 `json:"z"`
	} `json:"position"`
	SolarSystemID int64 `json:"solar_system_id"`
	TypeID        int64 `json:"type_id"`
}

func (t *StructureTask) Execute(ctx *TaskContext) error {
	bgCtx := context.Background()
	now := time.Now().Unix()

	// 0. 判断是否权限
	var corpRoles []string
	err := global.DB.Model(&model.EveCharacterCorpRole{}).
		Where("character_id = ?", ctx.CharacterID).
		Pluck("corp_role", &corpRoles).Error
	if err != nil {
		return fmt.Errorf("query corp roles: %w", err)
	}
	if !utils.ContainsAny(corpRoles, []string{"Director", "Station_Manager"}) {
		global.Logger.Debug("[ESI] 角色没有足够的军团权限，跳过建筑信息刷新",
			zap.Int64("character_id", ctx.CharacterID),
			zap.Strings("corp_roles", corpRoles))
		return nil
	}

	var corpID int64
	err = global.DB.Model(&model.EveCharacter{}).
		Where("character_id = ?", ctx.CharacterID).
		Pluck("corporation_id", &corpID).Error
	if err != nil {
		return fmt.Errorf("query corporation id: %w", err)
	}

	// 1. 获取角色所在军团的建筑列表
	var esiStructures []corpStructureESIResp
	corpStructuresPath := fmt.Sprintf("/corporations/%d/structures/", corpID)
	_, err = ctx.Client.GetPaginated(bgCtx, corpStructuresPath, ctx.AccessToken, &esiStructures)
	if err != nil {
		global.Logger.Warn("[ESI] 获取军团建筑信息失败",
			zap.Int64("character_id", ctx.CharacterID),
			zap.Int64("corporation_id", corpID),
			zap.Error(err),
		)
		return fmt.Errorf("fetch corp structures: %w", err)
	}

	if len(esiStructures) == 0 {
		return nil
	}

	// 2. 批量 Upsert CorpStructureInfo
	corpRecords := make([]model.CorpStructureInfo, 0, len(esiStructures))
	for _, s := range esiStructures {
		services := make(model.CorpStructureServices, 0, len(s.Services))
		for _, svc := range s.Services {
			services = append(services, model.CorpStructureService{
				Name:  svc.Name,
				State: svc.State,
			})
		}
		corpRecords = append(corpRecords, model.CorpStructureInfo{
			CorporationID:      corpID,
			StructureID:        s.StructureID,
			FuelExpires:        s.FuelExpires,
			Name:               s.Name,
			NextReinforceApply: s.NextReinforceApply,
			NextReinforceHour:  s.NextReinforceHour,
			ProfileID:          s.ProfileID,
			ReinforceHour:      s.ReinforceHour,
			State:              s.State,
			StateTimerEnd:      s.StateTimerEnd,
			StateTimerStart:    s.StateTimerStart,
			SystemID:           s.SystemID,
			TypeID:             s.TypeID,
			UnanchorsAt:        s.UnanchorsAt,
			Services:           services,
			UpdateAt:           now,
		})
	}
	if err := global.DB.Clauses(clause.OnConflict{UpdateAll: true}).
		Create(&corpRecords).Error; err != nil {
		return fmt.Errorf("upsert corp structures: %w", err)
	}

	// 3. 逐个获取建筑详情并 Upsert EveStructure
	for _, s := range esiStructures {
		var detail eveStructureDetail
		structurePath := fmt.Sprintf("/universe/structures/%d/", s.StructureID)
		if err := ctx.Client.Get(bgCtx, structurePath, ctx.AccessToken, &detail); err != nil {
			global.Logger.Warn("[ESI] 获取建筑详情失败",
				zap.Int64("structure_id", s.StructureID),
				zap.Error(err),
			)
			continue
		}

		record := model.EveStructure{
			StructureID:   s.StructureID,
			StructureName: detail.Name,
			OwnerID:       detail.OwnerID,
			TypeID:        detail.TypeID,
			SolarSystemID: detail.SolarSystemID,
			X:             detail.Position.X,
			Y:             detail.Position.Y,
			Z:             detail.Position.Z,
			UpdateAt:      now,
		}
		if err := global.DB.Clauses(clause.OnConflict{UpdateAll: true}).
			Create(&record).Error; err != nil {
			global.Logger.Warn("[ESI] Upsert 建筑详情失败",
				zap.Int64("structure_id", s.StructureID),
				zap.Error(err),
			)
		}
	}

	global.Logger.Debug("[ESI] 建筑信息刷新完成",
		zap.Int64("character_id", ctx.CharacterID),
		zap.Int("count", len(esiStructures)),
	)

	return nil
}
