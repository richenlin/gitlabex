package services

import (
	"context"
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

// Close 关闭Redis连接
func (r *RedisService) Close() error {
	return r.client.Close()
}
