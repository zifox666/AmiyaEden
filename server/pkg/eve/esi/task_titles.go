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
//  Character Titles 角色军团头衔
//  GET /characters/{character_id}/titles
//  默认刷新间隔: 6 Hours / 不活跃: 7 Days
// ─────────────────────────────────────────────

func init() {
	Register(&TitlesTask{})
}

// TitlesTask 角色头衔刷新任务
type TitlesTask struct{}

func (t *TitlesTask) Name() string        { return "character_titles" }
func (t *TitlesTask) Description() string { return "角色军团头衔" }
func (t *TitlesTask) Priority() Priority  { return PriorityNormal }

func (t *TitlesTask) Interval() RefreshInterval {
	return RefreshInterval{
		Active:   6 * time.Hour,
		Inactive: 7 * 24 * time.Hour,
	}
}

func (t *TitlesTask) RequiredScopes() []TaskScope {
	return []TaskScope{
		{Scope: "esi-characters.read_titles.v1", Description: "读取角色头衔"},
	}
}

// CharacterTitle 角色头衔
type CharacterTitle struct {
	Name    *string `json:"name,omitempty"`
	TitleID *int    `json:"title_id,omitempty"`
}

func (t *TitlesTask) Execute(ctx *TaskContext) error {
	bgCtx := context.Background()
	path := fmt.Sprintf("/characters/%d/titles/", ctx.CharacterID)

	var titles []CharacterTitle
	if err := ctx.Client.Get(bgCtx, path, ctx.AccessToken, &titles); err != nil {
		return fmt.Errorf("fetch titles: %w", err)
	}

	global.Logger.Debug("[ESI] 角色头衔刷新完成",
		zap.Int64("character_id", ctx.CharacterID),
		zap.Int("count", len(titles)),
	)

	// 入库：先删除旧数据，再插入新数据
	tx := global.DB.Begin()
	if err := tx.Where("character_id = ?", ctx.CharacterID).Delete(&model.EveCharacterTitle{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("delete old titles: %w", err)
	}
	if len(titles) > 0 {
		records := make([]model.EveCharacterTitle, 0, len(titles))
		for _, t := range titles {
			titleID := 0
			if t.TitleID != nil {
				titleID = *t.TitleID
			}
			name := ""
			if t.Name != nil {
				name = *t.Name
			}
			records = append(records, model.EveCharacterTitle{
				CharacterID: ctx.CharacterID,
				TitleID:     titleID,
				Name:        name,
			})
		}
		if err := tx.Create(&records).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("insert titles: %w", err)
		}
	}

	return tx.Commit().Error
}
