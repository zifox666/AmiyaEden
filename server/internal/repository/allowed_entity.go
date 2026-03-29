package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"amiya-eden/pkg/cache"
	"context"
	"encoding/json"
	"time"
)

const (
	allowEntityCachePrefix = "allow_entity:"
	allowEntityCacheTTL    = 5 * time.Minute
)

// AllowedEntityRepository 准入名单数据访问层，带 Redis 缓存
type AllowedEntityRepository struct{}

func NewAllowedEntityRepository() *AllowedEntityRepository {
	return &AllowedEntityRepository{}
}

// List 获取指定名单类型的所有实体
func (r *AllowedEntityRepository) List(listType string) ([]model.AllowedEntity, error) {
	var entities []model.AllowedEntity
	err := global.DB.
		Where("list_type = ?", listType).
		Order("created_at ASC").
		Find(&entities).Error
	return entities, err
}

// Add 添加实体到名单（已存在则忽略）
func (r *AllowedEntityRepository) Add(e *model.AllowedEntity) error {
	result := global.DB.
		Where("list_type = ? AND entity_id = ?", e.ListType, e.EntityID).
		FirstOrCreate(e)
	if result.Error != nil {
		return result.Error
	}
	r.invalidateCache(e.ListType)
	return nil
}

// Remove 从名单中删除实体
func (r *AllowedEntityRepository) Remove(id uint) error {
	var e model.AllowedEntity
	if err := global.DB.First(&e, id).Error; err != nil {
		return err
	}
	if err := global.DB.Delete(&model.AllowedEntity{}, id).Error; err != nil {
		return err
	}
	r.invalidateCache(e.ListType)
	return nil
}

// GetAllIDs 返回指定名单的军团 ID 列表和联盟 ID 列表（带缓存）
func (r *AllowedEntityRepository) GetAllIDs(listType string) (corpIDs []int64, allianceIDs []int64, err error) {
	type cachedIDs struct {
		CorpIDs     []int64 `json:"corp_ids"`
		AllianceIDs []int64 `json:"alliance_ids"`
	}

	ctx := context.Background()
	key := allowEntityCachePrefix + listType + ":ids"

	// 1. 尝试读缓存
	if raw, err2 := cache.GetString(ctx, key); err2 == nil {
		var cached cachedIDs
		if json.Unmarshal([]byte(raw), &cached) == nil {
			return cached.CorpIDs, cached.AllianceIDs, nil
		}
	}

	// 2. 查数据库
	entities, err := r.List(listType)
	if err != nil {
		return nil, nil, err
	}

	var corps, alliances []int64
	for _, e := range entities {
		switch e.EntityType {
		case model.AllowEntityTypeCorporation:
			corps = append(corps, e.EntityID)
		case model.AllowEntityTypeAlliance:
			alliances = append(alliances, e.EntityID)
		}
	}

	// 3. 回写缓存
	if b, err2 := json.Marshal(cachedIDs{CorpIDs: corps, AllianceIDs: alliances}); err2 == nil {
		_ = cache.SetString(ctx, key, string(b), allowEntityCacheTTL)
	}

	return corps, alliances, nil
}

// GetCorporationIDs 仅返回军团 ID 列表（带缓存）
func (r *AllowedEntityRepository) GetCorporationIDs(listType string) ([]int64, error) {
	corpIDs, _, err := r.GetAllIDs(listType)
	return corpIDs, err
}

// IsNonEmpty 判断指定名单是否有配置（有则返回 true，代表需要准入控制）
func (r *AllowedEntityRepository) IsNonEmpty(listType string) (bool, error) {
	var count int64
	err := global.DB.Model(&model.AllowedEntity{}).Where("list_type = ?", listType).Count(&count).Error
	return count > 0, err
}

// invalidateCache 使对应名单的缓存失效
func (r *AllowedEntityRepository) invalidateCache(listType string) {
	ctx := context.Background()
	_ = cache.Del(ctx, allowEntityCachePrefix+listType+":ids")
}
