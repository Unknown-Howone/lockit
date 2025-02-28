package lockit

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisLocker struct {
	client *redis.Client
}

func NewRedisLocker(client *redis.Client) *RedisLocker {
	return &RedisLocker{
		client: client,
	}
}

// TryLock 尝试获取分布式锁
// key: 锁的名称
// value: 锁的值（通常是客户端的标识或 UUID）
// expiration: 锁的过期时间
func (r *RedisLocker) TryLock(ctx context.Context, key, value string, expiration time.Duration) (bool, error) {
	ok, err := r.client.SetNX(ctx, key, value, expiration).Result()
	if err != nil {
		return false, fmt.Errorf("failed to set lock: %w", err)
	}
	return ok, nil
}

// Unlock 释放分布式锁
// key: 锁的名称
// value: 锁的值，必须与 TryLock 时提供的 value 一致
func (r *RedisLocker) Unlock(ctx context.Context, key, value string) error {
	// 使用 Lua 脚本确保只有持有锁的客户端能释放锁
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`

	result, err := r.client.Eval(ctx, script, []string{key}, value).Result()
	if err != nil {
		return fmt.Errorf("failed to execute unlock: %w", err)
	}

	if result.(int64) == 0 {
		return errors.New("unlock failed: the lock is not owned by the provided value")
	}

	return nil
}

// IsLocked 检查是否有客户端持有锁
// key: 锁的名称
func (r *RedisLocker) IsLocked(ctx context.Context, key string) (bool, error) {
	value, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		// 锁不存在，表示没有持有锁
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to get lock status: %w", err)
	}
	// 锁存在，表示被占用
	return value != "", nil
}
