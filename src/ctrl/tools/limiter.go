package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

// RateLimiter 定义限额计数器
type RateLimiter struct {
	redisClient *redis.Client
	limit       int // 每个间隔限制的请求次数
	div         int //多少秒的限额，1秒就是1000，1分钟就是60000
}

// NewRateLimiter 创建一个新的限额计数器
func NewRateLimiter(client *redis.Client, div, limit int) *RateLimiter {
	return &RateLimiter{
		redisClient: client,
		limit:       limit,
		div:         div,
	}
}

// IsRequestAllowed 判断用户的请求是否被允许
func (r *RateLimiter) IsRequestAllowed(keyID string) (bool, error) {
	// 获取当前分钟级时间戳
	currentStart := time.Now().UnixMilli() / int64(r.div)
	key := fmt.Sprintf(keyID+":%d", currentStart)

	// 使用 Redis 的 INCR 操作
	count, err := r.redisClient.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}

	// 如果是第一次设置计数器，则设置过期时间
	if count == 1 {
		expire := time.Duration(r.div/1000) * time.Second
		_, err = r.redisClient.Expire(ctx, key, expire).Result()
		if err != nil {
			return false, err
		}
	}

	// 检查是否超过限制
	if count > int64(r.limit) {
		return false, nil
	}
	return true, nil
}
