package data

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	conf "github.com/BitofferHub/msgcenter/src/config"
	"github.com/BitofferHub/pkg/middlewares/cache"
	"github.com/BitofferHub/pkg/middlewares/gormcli"
	"github.com/BitofferHub/pkg/middlewares/log"
	"github.com/BitofferHub/pkg/middlewares/mq"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

// Data .
type Data struct {
	db        *gorm.DB
	rdb       *cache.Client
	producers map[PriorityEnum]mq.Producer
	consumers map[PriorityEnum]mq.Consumer
}

var data *Data

func GetData() *Data {
	return data
}
func (p *Data) GetDB() *gorm.DB {
	return p.db
}

func (p *Data) GetCache() *cache.Client {
	return p.rdb
}

// GetMsgTemplate retrieves a message template by ID, using cache when available
func (p *Data) GetMsgTemplate(ctx context.Context, templateID string) (*MsgTemplate, error) {
	var template *MsgTemplate

	// Try to get from cache if enabled
	if conf.Conf.Common.OpenCache {
		templateCacheKey := p.genTemplateCacheKey(templateID)
		cacheData, _, _ := p.GetCache().Get(ctx, templateCacheKey)
		if len(cacheData) > 0 {
			template = new(MsgTemplate)
			if err := json.Unmarshal([]byte(cacheData), template); err == nil {
				log.Infof("template cache hit %+v", template)
				return template, nil
			}
		}
	}

	// Cache miss or disabled, retrieve from database
	log.Infof("template cache miss")
	var err error
	template, err = MsgTemplateNsp.Find(p.GetDB(), templateID)
	if err != nil {
		log.Errorf("find msg template err %s", err.Error())
		return nil, err
	}

	// Cache the result if enabled
	if conf.Conf.Common.OpenCache {
		if cacheData, err := json.Marshal(template); err == nil {
			templateCacheKey := p.genTemplateCacheKey(templateID)
			p.GetCache().Set(ctx, templateCacheKey, string(cacheData), 30*time.Second)
		}
	}

	return template, nil
}

func (p *Data) genTemplateCacheKey(templateID string) string {
	return fmt.Sprintf("%s%s", REDIS_KEY_TEMPLATE, templateID)
}

func (p *Data) GetProducer(key PriorityEnum) mq.Producer {
	return p.producers[key]
}

func (p *Data) GetConsumer(key PriorityEnum) mq.Consumer {
	return p.consumers[key]
}

func (p *Data) GetLowMQProducer() mq.Producer {
	return p.producers[PRIORITY_LOW]
}

func (p *Data) GetLowMQConsumer() mq.Consumer {
	return p.consumers[PRIORITY_LOW]
}

func (p *Data) GetMiddleMQProducer() mq.Producer {
	return p.producers[PRIORITY_MIDDLE]
}

func (p *Data) GetMiddleMQConsumer() mq.Consumer {
	return p.consumers[PRIORITY_MIDDLE]
}

func (p *Data) GetHighMQProducer() mq.Producer {
	return p.producers[PRIORITY_HIGH]
}

func (p *Data) GetHighMQConsumer() mq.Consumer {
	return p.consumers[PRIORITY_HIGH]
}

func (p *Data) GetRetryMQProducer() mq.Producer {
	return p.producers[PRIORITY_RETRY]
}

func (p *Data) GetRetryMQConsumer() mq.Consumer {
	return p.consumers[PRIORITY_RETRY]
}

// NewData
//
//	@Author <a href="https://bitoffer.cn">狂飙训练营</a>
//	@Description:
//	@param dt
//	@return *Data
//	@return error
func NewData(cf *conf.TomlConfig) (*Data, error) {
	fmt.Printf("conf is %+v\n", cf)
	gormcli.Init(
		gormcli.WithAddr(cf.MySQL.Url),
		gormcli.WithUser(cf.MySQL.User),
		gormcli.WithPassword(cf.MySQL.Pwd),
		gormcli.WithDataBase(cf.MySQL.Dbname),
		gormcli.WithMaxIdleConn(2000),
		gormcli.WithMaxOpenConn(20000),
		gormcli.WithMaxIdleTime(30),
		gormcli.WithSlowThresholdMillisecond(10),
	)
	cache.Init(
		cache.WithAddr(cf.Redis.Url),
		cache.WithPassWord(cf.Redis.Pwd),
		cache.WithDB(0),
	)
	producers := generateProducer(cf)
	consumers := generateConsumer(cf)

	dta := &Data{
		db:        gormcli.GetDB(),
		rdb:       cache.GetRedisCli(),
		producers: producers,
		consumers: consumers,
	}
	data = dta
	fmt.Println("producer 2", data.GetLowMQProducer())
	fmt.Printf("data is %+v\n", data)
	fmt.Printf("data db is %+v\n", data.GetDB())
	return dta, nil
}

func generateProducer(cf *conf.TomlConfig) map[PriorityEnum]mq.Producer {
	log.Infof("生成生产者 %+v", cf.Kafka)
	producers := make(map[PriorityEnum]mq.Producer)

	for _, topicConfig := range cf.Kafka.Topics {
		producer := mq.NewKafkaProducer(
			mq.WithBrokers(cf.Kafka.Brokers),
			mq.WithTopic(topicConfig.Name),
			mq.WithAck(int8(topicConfig.Ack)),
			mq.WithGroupID(topicConfig.GroupID),
			mq.WithPartition(topicConfig.Partition),
			mq.WithAsync())

		if producer == nil {
			panic(fmt.Sprintf("nil producer for %s", topicConfig.Name))
		}
		producers[PriorityEnum(topicConfig.Priority)] = producer
	}

	return producers
}

func generateConsumer(cf *conf.TomlConfig) map[PriorityEnum]mq.Consumer {
	log.Infof("生成消费者 %+v", cf.Kafka)
	consumers := make(map[PriorityEnum]mq.Consumer)

	for _, topicConfig := range cf.Kafka.Topics {
		consumerOpts := []mq.Option{
			mq.WithBrokers(cf.Kafka.Brokers),
			mq.WithTopic(topicConfig.Name),
			mq.WithGroupID(topicConfig.GroupID),
			mq.WithPartition(topicConfig.Partition),
		}

		// 如果配置了消费者组ID，则添加
		if topicConfig.GroupID != "" {
			consumerOpts = append(consumerOpts, mq.WithGroupID(topicConfig.GroupID))
		}

		consumer := mq.NewKafkaConsumer(consumerOpts...)
		if consumer == nil {
			panic(fmt.Sprintf("nil consumer for %s", topicConfig.Name))
		}
		consumers[PriorityEnum(topicConfig.Priority)] = consumer
	}

	return consumers
}
