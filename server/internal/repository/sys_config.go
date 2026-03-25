package repository

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"amiya-eden/pkg/cache"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"
)

const (
	sysConfigCachePrefix = "sys_config:"
	sysConfigCacheTTL    = 10 * time.Minute
)

// SysConfigRepository 系统配置（key/value）数据访问层，带 Redis 缓存
type SysConfigRepository struct{}

func NewSysConfigRepository() *SysConfigRepository {
	return &SysConfigRepository{}
}

func cacheKey(key string) string { return sysConfigCachePrefix + key }

// Get 获取配置值字符串；若不存在返回 defaultVal
func (r *SysConfigRepository) Get(key, defaultVal string) (string, error) {
	ctx := context.Background()

	// 1. 先查缓存
	if val, err := cache.GetString(ctx, cacheKey(key)); err == nil {
		return val, nil
	}

	// 2. 查数据库
	var cfg model.SystemConfig
	err := global.DB.Where("key = ?", key).First(&cfg).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 写入默认值到 DB
			if saveErr := r.Set(key, defaultVal, ""); saveErr != nil {
				return defaultVal, nil
			}
			return defaultVal, nil
		}
		return defaultVal, err
	}

	// 3. 回写缓存
	_ = cache.SetString(ctx, cacheKey(key), cfg.Value, sysConfigCacheTTL)
	return cfg.Value, nil
}

// Set 设置配置值并使缓存失效
func (r *SysConfigRepository) Set(key, value, desc string) error {
	var cfg model.SystemConfig
	err := global.DB.Where("key = ?", key).First(&cfg).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			cfg = model.SystemConfig{Key: key, Value: value, Desc: desc}
			if err2 := global.DB.Create(&cfg).Error; err2 != nil {
				return err2
			}
		} else {
			return err
		}
	} else {
		updates := map[string]interface{}{"value": value}
		if desc != "" {
			updates["desc"] = desc
		}
		if err2 := global.DB.Model(&cfg).Where("key = ?", key).Updates(updates).Error; err2 != nil {
			return err2
		}
	}

	// 刷新缓存
	_ = cache.SetString(context.Background(), cacheKey(key), value, sysConfigCacheTTL)
	return nil
}

// GetFloat 获取 float64 配置；解析失败或不存在时返回 defaultVal
func (r *SysConfigRepository) GetFloat(key string, defaultVal float64) float64 {
	raw, err := r.Get(key, fmt.Sprintf("%g", defaultVal))
	if err != nil {
		return defaultVal
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return defaultVal
	}
	return v
}

// GetBool 获取 bool 配置；解析失败或不存在时返回 defaultVal
func (r *SysConfigRepository) GetBool(key string, defaultVal bool) bool {
	def := "false"
	if defaultVal {
		def = "true"
	}
	raw, err := r.Get(key, def)
	if err != nil {
		return defaultVal
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return defaultVal
	}
	return v
}

// Invalidate 手动使某个 key 的缓存失效
func (r *SysConfigRepository) Invalidate(keys ...string) {
	cacheKeys := make([]string, len(keys))
	for i, k := range keys {
		cacheKeys[i] = cacheKey(k)
	}
	_ = cache.Del(context.Background(), cacheKeys...)
}

// GetInt64Slice 获取 int64 数组配置；解析失败或不存在时返回 defaultVal
func (r *SysConfigRepository) GetInt64Slice(key string, defaultVal []int64) ([]int64, error) {
	raw, err := r.Get(key, "")
	if err != nil {
		return defaultVal, nil
	}
	if raw == "" {
		return defaultVal, nil
	}

	var result []int64
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return defaultVal, nil
	}
	return result, nil
}

// SetInt64Slice 设置 int64 数组配置并使缓存失效
func (r *SysConfigRepository) SetInt64Slice(key string, value []int64, desc string) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.Set(key, string(data), desc)
}
