package esi

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// ─────────────────────────────────────────────
//  Character Clones 角色克隆相关
//  GET /characters/{character_id}/clones
//  GET /characters/{character_id}/implants
//  GET /characters/{character_id}/fatigue
//  默认刷新间隔: 6 Hours / 不活跃: 7 Days
// ─────────────────────────────────────────────

func init() {
	Register(&ClonesTask{})
}

// ClonesTask 角色克隆信息刷新任务
type ClonesTask struct{}

func (t *ClonesTask) Name() string        { return "character_clones" }
func (t *ClonesTask) Description() string { return "角色克隆体/植入体/跳跃疲劳" }
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
		{Scope: " esi-characters.read_fatigue.v1", Description: "读取跳跃疲劳信息"},
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

	// 序列化跳跃克隆为 JSON
	jumpClonesJSON, _ := json.Marshal(clones.JumpClones)

	// 构造入库记录
	record := model.EveCharacterClone{
		CharacterID:           ctx.CharacterID,
		LastCloneJumpDate:     clones.LastCloneJumpDate,
		LastStationChangeDate: clones.LastStationChangeDate,
		JumpClonesJSON:        string(jumpClonesJSON),
	}
	if clones.HomeLocation != nil {
		record.HomeLocationID = clones.HomeLocation.LocationID
		record.HomeLocationType = clones.HomeLocation.LocationType
	}

	// 2. 获取当前植入体
	implantPath := fmt.Sprintf("/characters/%d/implants/", ctx.CharacterID)
	var implants []int
	if err := ctx.Client.Get(bgCtx, implantPath, ctx.AccessToken, &implants); err != nil {
		global.Logger.Warn("[ESI] 获取植入体失败",
			zap.Int64("character_id", ctx.CharacterID),
			zap.Error(err),
		)
	} else {
		implantsJSON, _ := json.Marshal(implants)
		record.ImplantsJSON = string(implantsJSON)
	}

	// 3. 获取跳跃疲劳
	fatiguePath := fmt.Sprintf("/characters/%d/fatigue/", ctx.CharacterID)
	var fatigue JumpFatigue
	if err := ctx.Client.Get(bgCtx, fatiguePath, ctx.AccessToken, &fatigue); err != nil {
		global.Logger.Warn("[ESI] 获取跳跃疲劳失败",
			zap.Int64("character_id", ctx.CharacterID),
			zap.Error(err),
		)
	} else {
		record.JumpFatigueExpire = fatigue.JumpFatigueExpireDate
		record.LastJumpDate = fatigue.LastJumpDate
	}

	// 4. Upsert 入库
	var existing model.EveCharacterClone
	result := global.DB.Where("character_id = ?", ctx.CharacterID).First(&existing)
	if result.Error != nil {
		// 新记录
		if err := global.DB.Create(&record).Error; err != nil {
			return fmt.Errorf("insert clone info: %w", err)
		}
	} else {
		// 更新
		record.ID = existing.ID
		if err := global.DB.Save(&record).Error; err != nil {
			return fmt.Errorf("update clone info: %w", err)
		}
	}

	global.Logger.Debug("[ESI] 克隆体信息入库完成",
		zap.Int64("character_id", ctx.CharacterID),
	)

	return nil
}
