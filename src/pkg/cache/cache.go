package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	rdb *redis.Client
}

type Config struct {
	addr     string
	password string
	db       int
}

var (
	client *Client
	config *Config
)

// Option 配置选项
type Option func(*Config)

// WithAddr 设置Redis地址
func WithAddr(addr string) Option {
	return func(c *Config) {
		c.addr = addr
	}
}

// WithPassWord 设置密码
func WithPassWord(password string) Option {
	return func(c *Config) {
		c.password = password
	}
}

// WithDB 设置数据库
func WithDB(db int) Option {
	return func(c *Config) {
		c.db = db
	}
}

// Init 初始化Redis客户端
func Init(opts ...Option) error {
	config = &Config{
		addr:     "localhost:6379",
		password: "",
		db:       0,
	}

	for _, opt := range opts {
		opt(config)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     config.addr,
		Password: config.password,
		DB:       config.db,
	})

	// 测试连接
	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return err
	}

	client = &Client{rdb: rdb}
	return nil
}

// GetRedisCli 获取Redis客户端
func GetRedisCli() *Client {
	return client
}

// Set 设置键值对
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.rdb.Set(ctx, key, value, expiration).Err()
}

// Get 获取值
func (c *Client) Get(ctx context.Context, key string) (string, time.Duration, error) {
	val, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		return "", 0, err
	}

	ttl, err := c.rdb.TTL(ctx, key).Result()
	if err != nil {
		return val, 0, err
	}

	return val, ttl, nil
}

// Del 删除键
func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.rdb.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
func (c *Client) Exists(ctx context.Context, keys ...string) (int64, error) {
	return c.rdb.Exists(ctx, keys...).Result()
}

// Expire 设置过期时间
func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return c.rdb.Expire(ctx, key, expiration).Err()
}

// GetClient 获取原始Redis客户端
func (c *Client) GetClient() *redis.Client {
	return c.rdb
}

// ZAdd 添加有序集合成员
func (c *Client) ZAdd(ctx context.Context, key string, members ...redis.Z) error {
	return c.rdb.ZAdd(ctx, key, members...).Err()
}

// ZRem 删除有序集合成员
func (c *Client) ZRem(ctx context.Context, key string, members ...interface{}) error {
	return c.rdb.ZRem(ctx, key, members...).Err()
}

// ZRangeByScore 按分数范围获取有序集合成员
func (c *Client) ZRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) ([]string, error) {
	return c.rdb.ZRangeByScore(ctx, key, opt).Result()
}

// GetRedisBaseConn 获取Redis基础连接（兼容性方法）
func (c *Client) GetRedisBaseConn() *redis.Client {
	return c.rdb
}

// EvalResults 执行Lua脚本并返回结果（兼容性方法）
func (c *Client) EvalResults(ctx context.Context, script string, keys []string, args ...interface{}) ([]interface{}, error) {
	result, err := c.rdb.Eval(ctx, script, keys, args...).Result()
	if err != nil {
		return nil, err
	}

	// 将结果转换为[]interface{}
	if resultSlice, ok := result.([]interface{}); ok {
		return resultSlice, nil
	}

	// 如果不是切片，包装成切片返回
	return []interface{}{result}, nil
}
