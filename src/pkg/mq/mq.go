package mq

import (
	"context"
	"log"
	"strings"

	"github.com/IBM/sarama"
)

// Producer 生产者接口
type Producer interface {
	SendMessage(topic string, message []byte) error
	Close() error
}

// Consumer 消费者接口
type Consumer interface {
	ConsumeMessages(ctx context.Context, handler func([]byte) error) error
	Close() error
}

// KafkaProducer Kafka生产者
type KafkaProducer struct {
	producer sarama.SyncProducer
	topic    string
}

// KafkaConsumer Kafka消费者
type KafkaConsumer struct {
	consumer sarama.ConsumerGroup
	topic    string
	groupID  string
}

// Config 配置
type Config struct {
	brokers   []string
	topic     string
	groupID   string
	partition int32
	ack       int8
	async     bool
}

// Option 配置选项
type Option func(*Config)

// WithBrokers 设置Kafka brokers
func WithBrokers(brokers []string) Option {
	return func(c *Config) {
		c.brokers = brokers
	}
}

// WithTopic 设置主题
func WithTopic(topic string) Option {
	return func(c *Config) {
		c.topic = topic
	}
}

// WithGroupID 设置消费者组ID
func WithGroupID(groupID string) Option {
	return func(c *Config) {
		c.groupID = groupID
	}
}

// WithPartition 设置分区
func WithPartition(partition int32) Option {
	return func(c *Config) {
		c.partition = partition
	}
}

// WithAck 设置确认模式
func WithAck(ack int8) Option {
	return func(c *Config) {
		c.ack = ack
	}
}

// WithAsync 设置异步模式
func WithAsync() Option {
	return func(c *Config) {
		c.async = true
	}
}

// NewKafkaProducer 创建Kafka生产者
func NewKafkaProducer(opts ...Option) Producer {
	config := &Config{
		brokers: []string{"localhost:9092"},
		ack:     1,
		async:   false,
	}

	for _, opt := range opts {
		opt(config)
	}

	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.RequiredAcks = sarama.RequiredAcks(config.ack)
	saramaConfig.Producer.Retry.Max = 3
	saramaConfig.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(config.brokers, saramaConfig)
	if err != nil {
		log.Printf("Failed to create Kafka producer: %v", err)
		return nil
	}

	return &KafkaProducer{
		producer: producer,
		topic:    config.topic,
	}
}

// NewKafkaConsumer 创建Kafka消费者
func NewKafkaConsumer(opts ...Option) Consumer {
	config := &Config{
		brokers: []string{"localhost:9092"},
		groupID: "default-group",
	}

	for _, opt := range opts {
		opt(config)
	}

	saramaConfig := sarama.NewConfig()
	saramaConfig.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumer, err := sarama.NewConsumerGroup(config.brokers, config.groupID, saramaConfig)
	if err != nil {
		log.Printf("Failed to create Kafka consumer: %v", err)
		return nil
	}

	return &KafkaConsumer{
		consumer: consumer,
		topic:    config.topic,
		groupID:  config.groupID,
	}
}

// SendMessage 发送消息
func (p *KafkaProducer) SendMessage(topic string, message []byte) error {
	if topic == "" {
		topic = p.topic
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(message),
	}

	_, _, err := p.producer.SendMessage(msg)
	return err
}

// Close 关闭生产者
func (p *KafkaProducer) Close() error {
	return p.producer.Close()
}

// ConsumerGroupHandler 消费者组处理器
type ConsumerGroupHandler struct {
	handler func([]byte) error
}

// Setup 设置
func (h *ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup 清理
func (h *ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim 消费消息
func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		if err := h.handler(message.Value); err != nil {
			log.Printf("Error processing message: %v", err)
			continue
		}
		session.MarkMessage(message, "")
	}
	return nil
}

// ConsumeMessages 消费消息
func (c *KafkaConsumer) ConsumeMessages(ctx context.Context, handler func([]byte) error) error {
	h := &ConsumerGroupHandler{handler: handler}
	topics := []string{c.topic}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := c.consumer.Consume(ctx, topics, h); err != nil {
				if strings.Contains(err.Error(), "context canceled") {
					return nil
				}
				log.Printf("Error consuming messages: %v", err)
				return err
			}
		}
	}
}

// Close 关闭消费者
func (c *KafkaConsumer) Close() error {
	return c.consumer.Close()
}
