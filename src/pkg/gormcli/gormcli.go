package gormcli

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	addr                     string
	user                     string
	password                 string
	database                 string
	maxIdleConn              int
	maxOpenConn              int
	maxIdleTime              int
	slowThresholdMillisecond int
}

var (
	db     *gorm.DB
	config *Config
)

// Option 配置选项
type Option func(*Config)

// WithAddr 设置数据库地址
func WithAddr(addr string) Option {
	return func(c *Config) {
		c.addr = addr
	}
}

// WithUser 设置用户名
func WithUser(user string) Option {
	return func(c *Config) {
		c.user = user
	}
}

// WithPassword 设置密码
func WithPassword(password string) Option {
	return func(c *Config) {
		c.password = password
	}
}

// WithDataBase 设置数据库名
func WithDataBase(database string) Option {
	return func(c *Config) {
		c.database = database
	}
}

// WithMaxIdleConn 设置最大空闲连接数
func WithMaxIdleConn(maxIdleConn int) Option {
	return func(c *Config) {
		c.maxIdleConn = maxIdleConn
	}
}

// WithMaxOpenConn 设置最大打开连接数
func WithMaxOpenConn(maxOpenConn int) Option {
	return func(c *Config) {
		c.maxOpenConn = maxOpenConn
	}
}

// WithMaxIdleTime 设置最大空闲时间
func WithMaxIdleTime(maxIdleTime int) Option {
	return func(c *Config) {
		c.maxIdleTime = maxIdleTime
	}
}

// WithSlowThresholdMillisecond 设置慢查询阈值
func WithSlowThresholdMillisecond(slowThresholdMillisecond int) Option {
	return func(c *Config) {
		c.slowThresholdMillisecond = slowThresholdMillisecond
	}
}

// Init 初始化数据库连接
func Init(opts ...Option) error {
	config = &Config{
		maxIdleConn:              10,
		maxOpenConn:              100,
		maxIdleTime:              30,
		slowThresholdMillisecond: 200,
	}

	for _, opt := range opts {
		opt(config)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.user, config.password, config.addr, config.database)

	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	sqlDB.SetMaxIdleConns(config.maxIdleConn)
	sqlDB.SetMaxOpenConns(config.maxOpenConn)
	sqlDB.SetConnMaxLifetime(time.Duration(config.maxIdleTime) * time.Second)

	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return db
}
