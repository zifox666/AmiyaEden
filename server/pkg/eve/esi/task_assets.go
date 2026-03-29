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
//  Character Assets 人物资产
//  GET /characters/{character_id}/assets
//  POST /characters/{character_id}/assets/names
//  默认刷新间隔: 1 Day / 不活跃: 7 Days
// ─────────────────────────────────────────────

// 可查询名称的物品类别
const (
	CelestialCategory  = 2
	ShipCategory       = 6
	DeployableCategory = 22
	StarbaseCategory   = 23
	OrbitalsCategory   = 46
	StructureCategory  = 65
)

var nameCategoryIDs = map[int]struct{}{
	CelestialCategory:  {},
	ShipCategory:       {},
	DeployableCategory: {},
	StarbaseCategory:   {},
	OrbitalsCategory:   {},
	StructureCategory:  {},
}

func init() {
	Register(&AssetsTask{})
}

// AssetsTask 人物资产刷新任务
type AssetsTask struct{}

func (t *AssetsTask) Name() string        { return "character_assets" }
func (t *AssetsTask) Description() string { return "人物资产（物品/位置/名称）" }
func (t *AssetsTask) Priority() Priority  { return PriorityNormal }

func (t *AssetsTask) Interval() RefreshInterval {
	return RefreshInterval{
		Active:   24 * time.Hour,
		Inactive: 7 * 24 * time.Hour,
	}
}

func (t *AssetsTask) RequiredScopes() []TaskScope {
	return []TaskScope{
		{Scope: "esi-assets.read_assets.v1", Description: "读取人物资产"},
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

// AssetName 资产名称（ESI 返回）
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

	global.Logger.Debug("[ESI] 人物资产刷新完成",
		zap.Int64("character_id", ctx.CharacterID),
		zap.Int("asset_count", len(assets)),
	)

	// 2. 收集需要查询名称的 item_id
	//    条件: is_singleton==true 且 categoryID 属于可命名类别
	//    需要先获取所有 typeID 对应的 categoryID
	typeIDs := make(map[int]struct{})
	for _, a := range assets {
		typeIDs[a.TypeID] = struct{}{}
	}
	typeCategoryMap := getTypeCategoryMap(typeIDs)

	var nameableItemIDs []int64
	for _, a := range assets {
		if !a.IsSingleton {
			continue
		}
		catID, ok := typeCategoryMap[a.TypeID]
		if !ok {
			continue
		}
		if _, match := nameCategoryIDs[catID]; match {
			nameableItemIDs = append(nameableItemIDs, a.ItemID)
		}
	}

	// 3. 批量查询资产名称 (POST /characters/{id}/assets/names/)
	nameMap := make(map[int64]string)
	if len(nameableItemIDs) > 0 {
		namePath := fmt.Sprintf("/characters/%d/assets/names/", ctx.CharacterID)
		const nameBatch = 1000
		for i := 0; i < len(nameableItemIDs); i += nameBatch {
			end := i + nameBatch
			if end > len(nameableItemIDs) {
				end = len(nameableItemIDs)
			}
			batch := nameableItemIDs[i:end]
			var names []AssetName
			if err := ctx.Client.PostJSON(bgCtx, namePath, ctx.AccessToken, batch, &names); err != nil {
				global.Logger.Warn("[ESI] 查询资产名称失败",
					zap.Int64("character_id", ctx.CharacterID),
					zap.Error(err),
				)
			} else {
				for _, n := range names {
					if n.Name != "" && n.Name != "None" {
						nameMap[n.ItemID] = n.Name
					}
				}
			}
		}
	}

	global.Logger.Debug("[ESI] 资产名称查询完成",
		zap.Int64("character_id", ctx.CharacterID),
		zap.Int("named_items", len(nameMap)),
	)

	// 4. 入库：先删除该人物旧数据，再批量插入
	tx := global.DB.Begin()
	if err := tx.Where("character_id = ?", ctx.CharacterID).Delete(&model.EveCharacterAsset{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("delete old assets: %w", err)
	}

	if len(assets) > 0 {
		records := make([]model.EveCharacterAsset, 0, len(assets))
		for _, a := range assets {
			rec := model.EveCharacterAsset{
				CharacterID:     ctx.CharacterID,
				ItemID:          a.ItemID,
				TypeID:          a.TypeID,
				Quantity:        a.Quantity,
				LocationID:      a.LocationID,
				LocationType:    a.LocationType,
				LocationFlag:    a.LocationFlag,
				IsSingleton:     a.IsSingleton,
				IsBlueprintCopy: a.IsBlueprintCopy,
			}
			if name, ok := nameMap[a.ItemID]; ok {
				rec.AssetName = name
			}
			records = append(records, rec)
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

	global.Logger.Debug("[ESI] 人物资产入库完成",
		zap.Int64("character_id", ctx.CharacterID),
		zap.Int("count", len(assets)),
	)

	return nil
}

// getTypeCategoryMap 查询 typeID -> categoryID 映射
func getTypeCategoryMap(typeIDs map[int]struct{}) map[int]int {
	result := make(map[int]int)
	if len(typeIDs) == 0 {
		return result
	}

	ids := make([]int, 0, len(typeIDs))
	for id := range typeIDs {
		ids = append(ids, id)
	}

	type row struct {
		TypeID     int `gorm:"column:typeID"`
		CategoryID int `gorm:"column:categoryID"`
	}
	var rows []row

	// 分批查询
	const batchSize = 500
	for i := 0; i < len(ids); i += batchSize {
		end := i + batchSize
		if end > len(ids) {
			end = len(ids)
		}
		var batch []row
		global.DB.Table(`"invTypes" t`).
			Select(`t."typeID", g."categoryID"`).
			Joins(`JOIN "invGroups" g ON g."groupID" = t."groupID"`).
			Where(`t."typeID" IN ?`, ids[i:end]).
			Scan(&batch)
		rows = append(rows, batch...)
	}

	for _, r := range rows {
		result[r.TypeID] = r.CategoryID
	}
	return result
}
