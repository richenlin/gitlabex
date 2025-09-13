package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisService Redis服务
type RedisService struct {
	client *redis.Client
}

// NewRedisService 创建Redis服务
func NewRedisService(host, port, password string) (*RedisService, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       0, // 使用默认数据库
	})

	// 测试连接
	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisService{
		client: rdb,
	}, nil
}

// SetOAuthState 存储OAuth state
func (r *RedisService) SetOAuthState(state string, expiration time.Duration) error {
	ctx := context.Background()
	key := fmt.Sprintf("oauth_state:%s", state)

	err := r.client.Set(ctx, key, "valid", expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set OAuth state: %w", err)
	}

	return nil
}

// ValidateOAuthState 验证OAuth state
func (r *RedisService) ValidateOAuthState(state string) (bool, error) {
	ctx := context.Background()
	key := fmt.Sprintf("oauth_state:%s", state)

	result, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			// key不存在
			return false, nil
		}
		return false, fmt.Errorf("failed to get OAuth state: %w", err)
	}

	// 验证成功后删除state，确保一次性使用
	r.client.Del(ctx, key)

	return result == "valid", nil
}

// SetCache 设置缓存
func (r *RedisService) SetCache(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal cache value: %w", err)
	}

	err = r.client.Set(ctx, key, jsonData, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

// GetCache 获取缓存
func (r *RedisService) GetCache(ctx context.Context, key string, dest interface{}) error {
	jsonData, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("cache key not found: %s", key)
		}
		return fmt.Errorf("failed to get cache: %w", err)
	}

	err = json.Unmarshal([]byte(jsonData), dest)
	if err != nil {
		return fmt.Errorf("failed to unmarshal cache value: %w", err)
	}

	return nil
}

// DeleteCache 删除缓存
func (r *RedisService) DeleteCache(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete cache: %w", err)
	}
	return nil
}

// DeleteCachePattern 删除匹配模式的缓存
func (r *RedisService) DeleteCachePattern(ctx context.Context, pattern string) error {
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys by pattern: %w", err)
	}

	if len(keys) > 0 {
		err = r.client.Del(ctx, keys...).Err()
		if err != nil {
			return fmt.Errorf("failed to delete cache keys: %w", err)
		}
	}

	return nil
}

// CacheExists 检查缓存是否存在
func (r *RedisService) CacheExists(ctx context.Context, key string) (bool, error) {
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check cache existence: %w", err)
	}
	return exists > 0, nil
}

// GetCacheTTL 获取缓存剩余时间
func (r *RedisService) GetCacheTTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get cache TTL: %w", err)
	}
	return ttl, nil
}

// SetCacheWithTags 设置带标签的缓存（用于批量删除）
func (r *RedisService) SetCacheWithTags(ctx context.Context, key string, value interface{}, tags []string, expiration time.Duration) error {
	// 设置主缓存
	err := r.SetCache(ctx, key, value, expiration)
	if err != nil {
		return err
	}

	// 为每个标签创建索引
	for _, tag := range tags {
		tagKey := fmt.Sprintf("cache_tag:%s", tag)
		err = r.client.SAdd(ctx, tagKey, key).Err()
		if err != nil {
			return fmt.Errorf("failed to add cache to tag: %w", err)
		}
		// 设置标签的过期时间
		r.client.Expire(ctx, tagKey, expiration)
	}

	return nil
}

// DeleteCacheByTag 根据标签删除缓存
func (r *RedisService) DeleteCacheByTag(ctx context.Context, tag string) error {
	tagKey := fmt.Sprintf("cache_tag:%s", tag)
	keys, err := r.client.SMembers(ctx, tagKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys by tag: %w", err)
	}

	if len(keys) > 0 {
		err = r.client.Del(ctx, keys...).Err()
		if err != nil {
			return fmt.Errorf("failed to delete cache keys by tag: %w", err)
		}
	}

	// 删除标签索引
	r.client.Del(ctx, tagKey)
	return nil
}

// Close 关闭Redis连接
func (r *RedisService) Close() error {
	return r.client.Close()
}
