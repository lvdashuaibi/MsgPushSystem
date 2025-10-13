package config

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/lvdashuaibi/MsgPushSystem/src/pkg/log"
)

var Conf *TomlConfig

var (
	TestFilePath string
	configFile   string // 配置文件路径
)

// 初始化命令行参数
func init() {
	flag.StringVar(&configFile, "config", "", "配置文件路径")
}

// TomlConfig 配置
type TomlConfig struct {
	Common commonConfig
	MySQL  mysqlConfig
	Redis  redisConfig
	Kafka  kafkaConfig
	Task   TaskConfig
}

type commonConfig struct {
	Port            int    `toml:"port"`
	OpenTLS         bool   `toml:"open_tls"`
	MySQLAsMq       bool   `toml:"mysql_as_mq"`
	AliAppID        string `toml:"ali_app_id"`
	AliAppSecret    string `toml:"ali_app_secret"`
	EmailAccount    string `toml:"email_account"`
	EmailAuthCode   string `toml:"email_auth_code"`
	LarkAppID       string `toml:"lark_app_id"`     // 飞书应用ID
	LarkAppSecret   string `toml:"lark_app_secret"` // 飞书应用密钥
	ConsumePriority int    `toml:"consume_priority"`
	OpenCache       bool   `toml:"open_cache"`
	MaxRetryCount   int    `toml:"max_retry_count"` // 最大重试次数，默认20次
}

type mysqlConfig struct {
	Url    string `toml:"url"`
	User   string `toml:"user"`
	Pwd    string `toml:"pwd"`
	Dbname string `toml:"db_name"`
}

type redisConfig struct {
	Url                    string `toml:"url"`
	Pwd                    string `toml:"pwd"`
	MaxIdle                int    `toml:"max_idle"`
	MaxActive              int    `toml:"max_active"`
	IdleTimeout            int    `toml:"idle_timeout"`
	CacheTimeout           int    `toml:"cache_timeout"`
	CacheTimeoutVerifyCode int    `toml:"cache_timeout_verify_code"`
	CacheTimeoutDay        int    `toml:"cache_timeout_day"`
}

type kafkaConfig struct {
	Brokers []string               `toml:"brokers"`
	Topics  map[string]TopicConfig `toml:"topics"`
}

type TopicConfig struct {
	Name      string `toml:"name"`
	Priority  int    `toml:"priority"`
	Ack       int    `toml:"ack"`
	Async     bool   `toml:"async"`
	Offset    int64  `toml:"offset"`
	GroupID   string `toml:"group_id"`
	Partition int    `toml:"partition"`
}

type TaskConfig struct {
	TableMaxRows        int   `toml:"table_max_rows"`
	AliveThreshold      int   `toml:"alive_threshold"`
	SplitInterval       int   `toml:"split_interval"`
	LongProcessInterval int   `toml:"long_process_interval"`
	MoveInterval        int   `toml:"move_interval"`
	MaxProcessTime      int64 `toml:"max_process_time"`
}

// LoadConfig 导入配置
func (c *TomlConfig) LoadConfig(env string) {
	var filePath string

	// 如果通过命令行参数指定了配置文件路径，则使用该路径
	if configFile != "" {
		filePath = configFile
	} else {
		// 否则，使用环境变量或默认路径
		if env == "" {
			env = "test"
		}

		filePath = "../config/config-" + env + ".toml"
		if TestFilePath != "" {
			filePath = TestFilePath
		}
	}

	// 如果是相对路径，获取绝对路径
	if !filepath.IsAbs(filePath) {
		absPath, err := filepath.Abs(filePath)
		if err == nil {
			filePath = absPath
		}
	}

	log.Infof("使用配置文件: %s", filePath)

	if _, err := os.Stat(filePath); err != nil {
		log.Errorf("配置文件不存在: %s,err %s", filePath, err.Error())
		panic(err)
	}

	if _, err := toml.DecodeFile(filePath, &c); err != nil {
		log.Errorf("解析配置文件失败: %s", err)
		panic(err)
	}

	// 设置最大重试次数(默认20次)
	if c.Common.MaxRetryCount == 0 {
		c.Common.MaxRetryCount = 20
	}
}

const (
	USAGE = "Usage: msgcenter [-e <test|prod>] or [--config <config_file_path>]"
)

// GetConfEnv 获取配置的环境变量
func GetConfEnv() string {
	// 解析命令行参数
	flag.Parse()

	// 如果指定了配置文件，则环境变量不再需要
	if configFile != "" {
		return ""
	}

	env := os.Getenv("ENV")
	if env == "" {
		// 如果没有足够的参数，则使用默认环境
		if len(os.Args) < 2 {
			env = "test"
		} else if len(os.Args) >= 4 {
			env = "test"
		} else {
			env = os.Args[1]
		}
	}

	return env
}

func Init() {
	// 初始化配置
	env := GetConfEnv()

	// 初始化日志
	log.Init(
		log.WithLogPath("./log/"),
		log.WithLogLevel("info"),
		log.WithFileName("msgcenter.log"),
		log.WithMaxBackups(100),
		log.WithMaxSize(1024*1024*10),
		log.WithConsole(true),
	)

	InitConf(env)
}

// InitConf 初始化配置
func InitConf(env string) {
	Conf = new(TomlConfig)
	Conf.LoadConfig(env)
	printLog()
}

func printLog() {
	log.Infof("======== [Common] ========")
	log.Infof("%+v", Conf.Common)
	log.Infof("======== [MySQL] ========")
	log.Infof("%+v", Conf.MySQL)
	log.Infof("======== [Redis] ========")
	log.Infof("%+v", Conf.Redis)
	log.Infof("======== [Kafka] ========")
	log.Infof("%+v", Conf.Kafka)
}
