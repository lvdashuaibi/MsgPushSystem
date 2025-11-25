package lock

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
)

// RedisLock Redis分布式锁
type RedisLock struct {
	mutex      *redsync.Mutex
	key        string
	expireTime time.Duration
	watchdog   bool
	cancel     context.CancelFunc
}

// Config 锁配置
type Config struct {
	expireSeconds int
	watchdogMode  bool
}

// Option 配置选项
type Option func(*Config)

// WithExpireSeconds 设置过期时间（秒）
func WithExpireSeconds(seconds int) Option {
	return func(c *Config) {
		c.expireSeconds = seconds
	}
}

// WithWatchDogMode 启用看门狗模式
func WithWatchDogMode() Option {
	return func(c *Config) {
		c.watchdogMode = true
	}
}

var (
	rs *redsync.Redsync
)

// InitRedsync 初始化Redsync（需要在使用锁之前调用）
func InitRedsync(rdb *redis.Client) {
	pool := goredis.NewPool(rdb)
	rs = redsync.New(pool)
}

// NewRedisLock 创建Redis分布式锁
func NewRedisLock(key string, opts ...Option) *RedisLock {
	config := &Config{
		expireSeconds: 30,
		watchdogMode:  false,
	}

	for _, opt := range opts {
		opt(config)
	}

	if rs == nil {
		// 如果没有初始化，使用默认Redis客户端
		rdb := redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		})
		InitRedsync(rdb)
	}

	expireTime := time.Duration(config.expireSeconds) * time.Second
	mutex := rs.NewMutex(key, redsync.WithExpiry(expireTime))

	return &RedisLock{
		mutex:      mutex,
		key:        key,
		expireTime: expireTime,
		watchdog:   config.watchdogMode,
	}
}

// TryLock 尝试获取锁
func (l *RedisLock) TryLock(ctx context.Context) bool {
	err := l.mutex.LockContext(ctx)
	if err != nil {
		return false
	}

	// 如果启用看门狗模式，启动自动续期
	if l.watchdog {
		l.startWatchdog(ctx)
	}

	return true
}

// Lock 获取锁（阻塞）
func (l *RedisLock) Lock(ctx context.Context) error {
	err := l.mutex.LockContext(ctx)
	if err != nil {
		return err
	}

	// 如果启用看门狗模式，启动自动续期
	if l.watchdog {
		l.startWatchdog(ctx)
	}

	return nil
}

// Unlock 释放锁
func (l *RedisLock) Unlock() error {
	// 停止看门狗
	if l.cancel != nil {
		l.cancel()
	}

	ok, err := l.mutex.Unlock()
	if !ok && err == nil {
		return fmt.Errorf("failed to unlock")
	}
	return err
}

// UnlockWithContext 带上下文释放锁（兼容性方法）
func (l *RedisLock) UnlockWithContext(ctx context.Context) error {
	return l.Unlock()
}

// startWatchdog 启动看门狗自动续期
func (l *RedisLock) startWatchdog(ctx context.Context) {
	watchdogCtx, cancel := context.WithCancel(ctx)
	l.cancel = cancel

	go func() {
		ticker := time.NewTicker(l.expireTime / 3) // 每1/3过期时间续期一次
		defer ticker.Stop()

		for {
			select {
			case <-watchdogCtx.Done():
				return
			case <-ticker.C:
				// 续期锁
				l.mutex.Extend()
			}
		}
	}()
}

// IsLocked 检查锁是否被持有
func (l *RedisLock) IsLocked() bool {
	// 这里可以通过检查Redis中的键来判断
	// 但redsync库没有直接提供这个方法，可以通过尝试获取锁来判断
	return false // 简化实现
}
