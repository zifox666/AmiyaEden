package esi

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

func init() {
	Register(&FittingsTask{})
}

type FittingsTask struct{}

func (t *FittingsTask) Name() string        { return "character_fittings" }
func (t *FittingsTask) Description() string { return "角色装配" }
func (t *FittingsTask) Priority() Priority  { return PriorityNormal }

func (t *FittingsTask) Interval() RefreshInterval {
	return RefreshInterval{
		Active:   6 * time.Hour,
		Inactive: 7 * 24 * time.Hour,
	}
}

func (t *FittingsTask) RequiredScopes() []TaskScope {
	return []TaskScope{
		{Scope: "esi-fittings.read_fittings.v1", Description: "读取角色装配"},
		{Scope: "esi-fittings.write_fittings.v1", Description: "修改角色装配"},
	}
}

type fittingInfo struct {
	Description string `json:"description"`
	Name        string `json:"name"`
	ShipTypeID  int64  `json:"ship_type_id"`
	FittingID   int64  `json:"fitting_id"`
	Items       []struct {
		TypeID   int64  `json:"type_id"`
		Quantity int    `json:"quantity"`
		Flag     string `json:"flag"`
	} `json:"items"`
}

type fittingsResponse []fittingInfo

func (t *FittingsTask) Execute(ctx *TaskContext) error {
	bgCtx := context.Background()
	path := fmt.Sprintf("/characters/%d/fittings/", ctx.CharacterID)

	var fittings fittingsResponse
	if err := ctx.Client.Get(bgCtx, path, ctx.AccessToken, &fittings); err != nil {
		return fmt.Errorf("fetch fittings: %w", err)
	}

	global.Logger.Debug("[ESI] 角色装配刷新完成",
		zap.Int64("character_id", ctx.CharacterID),
		zap.Int("count", len(fittings)),
	)

	// 入库：先删除该角色旧数据，再批量插入
	tx := global.DB.Begin()

	if err := tx.Where("character_id = ?", ctx.CharacterID).Delete(&model.EveCharacterFittingItem{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("delete old fitting items: %w", err)
	}
	if err := tx.Where("character_id = ?", ctx.CharacterID).Delete(&model.EveCharacterFitting{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("delete old fittings: %w", err)
	}

	for _, f := range fittings {
		fitting := model.EveCharacterFitting{
			FittingID:   f.FittingID,
			CharacterID: ctx.CharacterID,
			Name:        f.Name,
			ShipTypeID:  f.ShipTypeID,
			Description: f.Description,
		}
		if err := tx.Create(&fitting).Error; err != nil {
			global.Logger.Warn("[ESI] 创建装配记录失败",
				zap.Int64("character_id", ctx.CharacterID),
				zap.Int64("fitting_id", f.FittingID),
				zap.Error(err),
			)
			continue
		}

		if len(f.Items) > 0 {
			items := make([]model.EveCharacterFittingItem, 0, len(f.Items))
			for _, item := range f.Items {
				items = append(items, model.EveCharacterFittingItem{
					FittingID:   f.FittingID,
					CharacterID: ctx.CharacterID,
					TypeID:      item.TypeID,
					Quantity:    item.Quantity,
					Flag:        item.Flag,
				})
			}
			if err := tx.Create(&items).Error; err != nil {
				global.Logger.Warn("[ESI] 创建装配物品记录失败",
					zap.Int64("character_id", ctx.CharacterID),
					zap.Int64("fitting_id", f.FittingID),
					zap.Error(err),
				)
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("commit fittings: %w", err)
	}

	global.Logger.Debug("[ESI] 角色装配入库完成",
		zap.Int64("character_id", ctx.CharacterID),
		zap.Int("count", len(fittings)),
	)

	return nil
}
