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
//  Character Assets 角色资产
//  GET /characters/{character_id}/assets
//  GET /characters/{character_id}/assets/locations
//  GET /characters/{character_id}/assets/names
//  默认刷新间隔: 1 Day / 不活跃: 7 Days
// ─────────────────────────────────────────────

func init() {
	Register(&AssetsTask{})
}

// AssetsTask 角色资产刷新任务
type AssetsTask struct{}

func (t *AssetsTask) Name() string        { return "character_assets" }
func (t *AssetsTask) Description() string { return "角色资产（物品/位置/名称）" }
func (t *AssetsTask) Priority() Priority  { return PriorityNormal }

func (t *AssetsTask) Interval() RefreshInterval {
	return RefreshInterval{
		Active:   24 * time.Hour,
		Inactive: 7 * 24 * time.Hour,
	}
}

func (t *AssetsTask) RequiredScopes() []TaskScope {
	return []TaskScope{
		{Scope: "esi-assets.read_assets.v1", Description: "读取角色资产"},
	}
}

// AssetItem 资产条目
type AssetItem struct {
	IsBlueprintCopy *bool  `json:"is_blueprint_copy,omitempty"`
	IsSingleton     bool   `json:"is_singleton"`
	ItemID          int64  `json:"item_id"`
	LocationFlag    string `json:"location_flag"`
	LocationID      int64  `json:"location_id"`
	LocationType    string `json:"location_type"`
	Quantity        int    `json:"quantity"`
	TypeID          int    `json:"type_id"`
}

// AssetLocation 资产位置
type AssetLocation struct {
	ItemID   int64 `json:"item_id"`
	Position struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
		Z float64 `json:"z"`
	} `json:"position"`
}

// AssetName 资产名称
type AssetName struct {
	ItemID int64  `json:"item_id"`
	Name   string `json:"name"`
}

func (t *AssetsTask) Execute(ctx *TaskContext) error {
	bgCtx := context.Background()

	// 1. 获取资产列表（自动分页）
	path := fmt.Sprintf("/characters/%d/assets/", ctx.CharacterID)
	var assets []AssetItem
	if _, err := ctx.Client.GetPaginated(bgCtx, path, ctx.AccessToken, &assets); err != nil {
		return fmt.Errorf("fetch assets: %w", err)
	}

	global.Logger.Debug("[ESI] 角色资产刷新完成",
		zap.Int64("character_id", ctx.CharacterID),
		zap.Int("asset_count", len(assets)),
	)

	// 入库：先删除该角色旧数据，再批量插入
	tx := global.DB.Begin()
	if err := tx.Where("character_id = ?", ctx.CharacterID).Delete(&model.EveCharacterAsset{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("delete old assets: %w", err)
	}

	if len(assets) > 0 {
		records := make([]model.EveCharacterAsset, 0, len(assets))
		for _, a := range assets {
			records = append(records, model.EveCharacterAsset{
				CharacterID:     ctx.CharacterID,
				ItemID:          a.ItemID,
				TypeID:          a.TypeID,
				Quantity:        a.Quantity,
				LocationID:      a.LocationID,
				LocationType:    a.LocationType,
				LocationFlag:    a.LocationFlag,
				IsSingleton:     a.IsSingleton,
				IsBlueprintCopy: a.IsBlueprintCopy,
			})
		}
		// 分批插入（每批 500）
		const batch = 500
		for i := 0; i < len(records); i += batch {
			end := i + batch
			if end > len(records) {
				end = len(records)
			}
			if err := tx.Create(records[i:end]).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("insert assets batch: %w", err)
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("commit assets: %w", err)
	}

	global.Logger.Debug("[ESI] 角色资产入库完成",
		zap.Int64("character_id", ctx.CharacterID),
		zap.Int("count", len(assets)),
	)

	return nil
}
