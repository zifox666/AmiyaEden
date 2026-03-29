package esi

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"amiya-eden/internal/utils"
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// ─────────────────────────────────────────────
//  Character Corporation Roles 军团职权权限
//  GET /characters/{character_id}/roles/
//  默认刷新间隔: 2 Hours / 不活跃: 1 Day
//  需要 scope: esi-characters.read_corporation_roles.v1
// ─────────────────────────────────────────────

func init() {
	Register(&CorpRolesTask{})
}

// CorpRolesTask 军团职权权限刷新任务
type CorpRolesTask struct{}

func (t *CorpRolesTask) Name() string        { return "character_corp_roles" }
func (t *CorpRolesTask) Description() string { return "人物军团职权" }
func (t *CorpRolesTask) Priority() Priority  { return PriorityHigh }

func (t *CorpRolesTask) Interval() RefreshInterval {
	return RefreshInterval{
		Active:   2 * time.Hour,
		Inactive: 24 * time.Hour,
	}
}

func (t *CorpRolesTask) RequiredScopes() []TaskScope {
	return []TaskScope{
		{Scope: "esi-characters.read_corporation_roles.v1", Description: "读取人物军团职权"},
	}
}

// corpRolesResponse ESI 返回的军团职权数据
type corpRolesResponse struct {
	Roles        []string `json:"roles"`
	RolesAtBase  []string `json:"roles_at_base"`
	RolesAtHQ    []string `json:"roles_at_hq"`
	RolesAtOther []string `json:"roles_at_other"`
}

func (t *CorpRolesTask) Execute(ctx *TaskContext) error {
	bgCtx := context.Background()
	path := fmt.Sprintf("/characters/%d/roles/", ctx.CharacterID)

	var rolesResp corpRolesResponse
	if err := ctx.Client.Get(bgCtx, path, ctx.AccessToken, &rolesResp); err != nil {
		return fmt.Errorf("fetch corporation roles: %w", err)
	}

	// 合并四个职权列表并去重
	roleSet := make(map[string]struct{})
	for _, r := range rolesResp.Roles {
		roleSet[r] = struct{}{}
	}
	for _, r := range rolesResp.RolesAtBase {
		roleSet[r] = struct{}{}
	}
	for _, r := range rolesResp.RolesAtHQ {
		roleSet[r] = struct{}{}
	}
	for _, r := range rolesResp.RolesAtOther {
		roleSet[r] = struct{}{}
	}

	roles := make([]string, 0, len(roleSet))
	for r := range roleSet {
		roles = append(roles, r)
	}

	var corpID int64
	if err := global.DB.Model(&model.EveCharacter{}).
		Where("character_id = ?", ctx.CharacterID).
		Pluck("corporation_id", &corpID).Error; err != nil {
		return fmt.Errorf("query corporation id: %w", err)
	}

	global.Logger.Debug("[ESI] 人物军团职权刷新完成",
		zap.Int64("character_id", ctx.CharacterID),
		zap.Int64("corporation_id", corpID),
		zap.Int("count", len(roles)),
		zap.Strings("roles", roles),
	)

	// 入库：同步人物的军团职权
	autoRoleRepo := repository.NewAutoRoleRepository()
	if !isCorporationAllowed(corpID, utils.GetAllowCorporations()) {
		if err := autoRoleRepo.SyncCharacterCorpRoles(ctx.CharacterID, nil); err != nil {
			return fmt.Errorf("clear corp roles for disallowed corporation: %w", err)
		}
		global.Logger.Debug("[ESI] 人物所在军团不在 allow_corporations，已忽略军团职权信号",
			zap.Int64("character_id", ctx.CharacterID),
			zap.Int64("corporation_id", corpID))
		return nil
	}

	if err := autoRoleRepo.SyncCharacterCorpRoles(ctx.CharacterID, roles); err != nil {
		return fmt.Errorf("sync corp roles: %w", err)
	}

	return nil
}
