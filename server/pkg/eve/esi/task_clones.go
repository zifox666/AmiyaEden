package esi

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// ─────────────────────────────────────────────
//  Character Clones 人物克隆相关
//  GET /characters/{character_id}/clones
//  GET /characters/{character_id}/implants
//  GET /characters/{character_id}/fatigue
//  默认刷新间隔: 6 Hours / 不活跃: 7 Days
// ─────────────────────────────────────────────

func init() {
	Register(&ClonesTask{})
}

// ClonesTask 人物克隆信息刷新任务
type ClonesTask struct{}

func (t *ClonesTask) Name() string        { return "character_clones" }
func (t *ClonesTask) Description() string { return "人物克隆体/植入体/跳跃疲劳" }
func (t *ClonesTask) Priority() Priority  { return PriorityNormal }

func (t *ClonesTask) Interval() RefreshInterval {
	return RefreshInterval{
		Active:   6 * time.Hour,
		Inactive: 7 * 24 * time.Hour,
	}
}

func (t *ClonesTask) RequiredScopes() []TaskScope {
	return []TaskScope{
		{Scope: "esi-clones.read_clones.v1", Description: "读取克隆体信息"},
		{Scope: "esi-clones.read_implants.v1", Description: "读取植入体信息"},
		{Scope: "esi-characters.read_fatigue.v1", Description: "读取跳跃疲劳信息"},
	}
}

// CloneInfo 克隆体信息
type CloneInfo struct {
	HomeLocation *struct {
		LocationID   int64  `json:"location_id"`
		LocationType string `json:"location_type"`
	} `json:"home_location,omitempty"`
	JumpClones []struct {
		Implants     []int  `json:"implants"`
		JumpCloneID  int64  `json:"jump_clone_id"`
		LocationID   int64  `json:"location_id"`
		LocationType string `json:"location_type"`
		Name         string `json:"name,omitempty"`
	} `json:"jump_clones"`
	LastCloneJumpDate     *time.Time `json:"last_clone_jump_date,omitempty"`
	LastStationChangeDate *time.Time `json:"last_station_change_date,omitempty"`
}

// JumpFatigue 跳跃疲劳
type JumpFatigue struct {
	JumpFatigueExpireDate *time.Time `json:"jump_fatigue_expire_date,omitempty"`
	LastJumpDate          *time.Time `json:"last_jump_date,omitempty"`
	LastUpdateDate        *time.Time `json:"last_update_date,omitempty"`
}

func (t *ClonesTask) Execute(ctx *TaskContext) error {
	bgCtx := context.Background()

	// 1. 获取克隆体信息
	clonePath := fmt.Sprintf("/characters/%d/clones/", ctx.CharacterID)
	var clones CloneInfo
	if err := ctx.Client.Get(bgCtx, clonePath, ctx.AccessToken, &clones); err != nil {
		return fmt.Errorf("fetch clones: %w", err)
	}

	global.Logger.Debug("[ESI] 克隆体信息刷新完成",
		zap.Int64("character_id", ctx.CharacterID),
		zap.Int("jump_clones", len(clones.JumpClones)),
	)

	// 2. 获取当前活跃植入体
	implantPath := fmt.Sprintf("/characters/%d/implants/", ctx.CharacterID)
	var activeImplants []int
	if err := ctx.Client.Get(bgCtx, implantPath, ctx.AccessToken, &activeImplants); err != nil {
		global.Logger.Warn("[ESI] 获取植入体失败",
			zap.Int64("character_id", ctx.CharacterID),
			zap.Error(err),
		)
		activeImplants = nil
	}

	// 3. 获取跳跃疲劳
	fatiguePath := fmt.Sprintf("/characters/%d/fatigue/", ctx.CharacterID)
	var fatigue JumpFatigue
	if err := ctx.Client.Get(bgCtx, fatiguePath, ctx.AccessToken, &fatigue); err != nil {
		global.Logger.Warn("[ESI] 获取跳跃疲劳失败",
			zap.Int64("character_id", ctx.CharacterID),
			zap.Error(err),
		)
	}

	// 4. Upsert EveCharacterCloneBaseInfo
	baseInfo := model.EveCharacterCloneBaseInfo{
		CharacterID:           ctx.CharacterID,
		LastCloneJumpDate:     clones.LastCloneJumpDate,
		LastStationChangeDate: clones.LastStationChangeDate,
		JumpFatigueExpire:     fatigue.JumpFatigueExpireDate,
		LastJumpDate:          fatigue.LastJumpDate,
	}
	if clones.HomeLocation != nil {
		baseInfo.HomeLocationID = clones.HomeLocation.LocationID
		baseInfo.HomeLocationType = clones.HomeLocation.LocationType
	}

	var existingBase model.EveCharacterCloneBaseInfo
	if err := global.DB.Where("character_id = ?", ctx.CharacterID).First(&existingBase).Error; err != nil {
		if err := global.DB.Create(&baseInfo).Error; err != nil {
			return fmt.Errorf("insert clone base info: %w", err)
		}
	} else {
		baseInfo.ID = existingBase.ID
		if err := global.DB.Save(&baseInfo).Error; err != nil {
			return fmt.Errorf("update clone base info: %w", err)
		}
	}

	// 5. 重建植入体记录：先清空旧数据，再批量插入
	if err := global.DB.Where("character_id = ?", ctx.CharacterID).
		Delete(&model.EveCharacterImplants{}).Error; err != nil {
		return fmt.Errorf("delete old implants: %w", err)
	}

	var implantRecords []model.EveCharacterImplants
	// 跳跃克隆的植入体
	for _, jc := range clones.JumpClones {
		if len(jc.Implants) == 0 {
			// 没有植入体时插入占位行（ImplantID=0），保留位置信息
			implantRecords = append(implantRecords, model.EveCharacterImplants{
				JumpCloneID:  jc.JumpCloneID,
				CharacterID:  ctx.CharacterID,
				ImplantID:    0,
				LocationID:   jc.LocationID,
				LocationType: jc.LocationType,
			})
		} else {
			for _, implantID := range jc.Implants {
				implantRecords = append(implantRecords, model.EveCharacterImplants{
					JumpCloneID:  jc.JumpCloneID,
					CharacterID:  ctx.CharacterID,
					ImplantID:    implantID,
					LocationID:   jc.LocationID,
					LocationType: jc.LocationType,
				})
			}
		}
	}
	// 当前活跃植入体（JumpCloneID = 0 表示当前克隆体）
	for _, implantID := range activeImplants {
		implantRecords = append(implantRecords, model.EveCharacterImplants{
			JumpCloneID: 0,
			CharacterID: ctx.CharacterID,
			ImplantID:   implantID,
		})
	}
	if len(implantRecords) > 0 {
		if err := global.DB.Create(&implantRecords).Error; err != nil {
			return fmt.Errorf("insert implants: %w", err)
		}
	}

	global.Logger.Debug("[ESI] 克隆体信息入库完成",
		zap.Int64("character_id", ctx.CharacterID),
	)

	return nil
}
