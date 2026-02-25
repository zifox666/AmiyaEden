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
//  Character Affiliation 角色归属
//  POST /characters/affiliation
//  批量任务：每次最多 1000 个角色 ID
//  默认刷新间隔: 2 Hours
// ─────────────────────────────────────────────

func init() {
	Register(&AffiliationTask{})
}

// AffiliationTask 角色归属刷新任务
type AffiliationTask struct{}

func (t *AffiliationTask) Name() string        { return "character_affiliation" }
func (t *AffiliationTask) Description() string { return "角色归属信息（军团/联盟/阵营）" }
func (t *AffiliationTask) Priority() Priority  { return PriorityNormal }

func (t *AffiliationTask) Interval() RefreshInterval {
	return RefreshInterval{
		Active:   2 * time.Hour,
		Inactive: 2 * time.Hour, // 统一任务，不区分活跃度
	}
}

func (t *AffiliationTask) RequiredScopes() []TaskScope {
	return nil // 公开接口，无需 scope
}

// AffiliationResult affiliation 查询结果
type AffiliationResult struct {
	AllianceID    *int64 `json:"alliance_id,omitempty"`
	CharacterID   int64  `json:"character_id"`
	CorporationID int64  `json:"corporation_id"`
	FactionID     *int64 `json:"faction_id,omitempty"`
}

func (t *AffiliationTask) Execute(ctx *TaskContext) error {
	// 单个角色模式：仅查询当前角色
	return t.fetchAffiliation(ctx.Client, []int64{ctx.CharacterID})
}

// ExecuteBatch 批量查询角色归属（最多 1000 个）
func (t *AffiliationTask) ExecuteBatch(client *Client, characterIDs []int64) error {
	// 分批处理，每批最多 1000
	const batchSize = 1000
	for i := 0; i < len(characterIDs); i += batchSize {
		end := i + batchSize
		if end > len(characterIDs) {
			end = len(characterIDs)
		}
		if err := t.fetchAffiliation(client, characterIDs[i:end]); err != nil {
			return err
		}
	}
	return nil
}

func (t *AffiliationTask) fetchAffiliation(client *Client, ids []int64) error {
	var results []AffiliationResult
	ctx := context.Background()

	if err := client.PostJSON(ctx, "/characters/affiliation/", "", ids, &results); err != nil {
		return fmt.Errorf("fetch affiliation: %w", err)
	}

	// 入库：更新 eve_character 表的归属字段
	for _, r := range results {
		updates := map[string]interface{}{
			"corporation_id": r.CorporationID,
			"alliance_id":    r.AllianceID,
			"faction_id":     r.FactionID,
		}
		if err := global.DB.Model(&model.EveCharacter{}).
			Where("character_id = ?", r.CharacterID).
			Updates(updates).Error; err != nil {
			global.Logger.Warn("[ESI] 更新角色归属失败",
				zap.Int64("character_id", r.CharacterID),
				zap.Error(err),
			)
		}
	}

	global.Logger.Debug("[ESI] 角色归属刷新并入库完成",
		zap.Int("count", len(results)),
	)

	return nil
}
