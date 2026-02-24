package cache

import (
	"amiya-eden/global"
	"context"
	"encoding/json"
	"time"
)

// Set 设置缓存（自动 JSON 序列化）
func Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return global.Redis.Set(ctx, key, data, expiration).Err()
}

// Get 获取缓存并反序列化到 dest
// 若 key 不存在返回 redis.Nil 错误
func Get(ctx context.Context, key string, dest any) error {
	data, err := global.Redis.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// GetString 获取字符串类型缓存
func GetString(ctx context.Context, key string) (string, error) {
	return global.Redis.Get(ctx, key).Result()
}

// SetString 设置字符串类型缓存
func SetString(ctx context.Context, key string, value string, expiration time.Duration) error {
	return global.Redis.Set(ctx, key, value, expiration).Err()
}

// Del 删除一个或多个缓存 key
func Del(ctx context.Context, keys ...string) error {
	return global.Redis.Del(ctx, keys...).Err()
}

// Exists 判断 key 是否存在
func Exists(ctx context.Context, key string) (bool, error) {
	n, err := global.Redis.Exists(ctx, key).Result()
	return n > 0, err
}

// Expire 重置 key 的过期时间
func Expire(ctx context.Context, key string, expiration time.Duration) error {
	return global.Redis.Expire(ctx, key, expiration).Err()
}

// TTL 获取 key 的剩余过期时间
func TTL(ctx context.Context, key string) (time.Duration, error) {
	return global.Redis.TTL(ctx, key).Result()
}

// Incr 对 key 的整数值加 1
func Incr(ctx context.Context, key string) (int64, error) {
	return global.Redis.Incr(ctx, key).Result()
}
