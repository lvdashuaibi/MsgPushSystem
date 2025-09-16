package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/BitofferHub/msgcenter/src/config"
	"github.com/BitofferHub/msgcenter/src/ctrl/ctrlmodel"
	"github.com/BitofferHub/msgcenter/src/ctrl/tools"
	"github.com/BitofferHub/msgcenter/src/data"
	"github.com/BitofferHub/pkg/middlewares/lock"
	"github.com/BitofferHub/pkg/middlewares/log"
	"github.com/BitofferHub/pkg/middlewares/mq"
	"gorm.io/gorm"
)

type MsgConsume struct {
	// 分布式锁映射，每个优先级一个锁
	locks map[data.PriorityEnum]*lock.RedisLock
	// 是否是主节点的标志，每个优先级一个标志
	isLeader map[data.PriorityEnum]bool
}

const (
	// 锁的前缀
	LOCK_KEY_PREFIX = "MSG_LEADER_CONSUMER"

	// 锁的过期时间（秒）
	LOCK_EXPIRE_SECONDS = 30

	// 非主节点尝试获取锁的间隔（秒）
	LOCK_RETRY_INTERVAL_SECONDS = 30
)

var consumePriority = []data.PriorityEnum{
	data.PRIORITY_HIGH,
	data.PRIORITY_MIDDLE,
	data.PRIORITY_LOW,
	data.PRIORITY_RETRY,
}

// NewMsgConsume 创建一个新的消息消费实例
func NewMsgConsume() *MsgConsume {
	return &MsgConsume{
		locks:    make(map[data.PriorityEnum]*lock.RedisLock),
		isLeader: make(map[data.PriorityEnum]bool),
	}
}

// Consume 方法用于启动消息消费
func (s *MsgConsume) Consume() {
	// 初始化锁和领导状态
	for _, priority := range consumePriority {
		priorityStr := data.GetPriorityStr(priority)
		lockKey := fmt.Sprintf("%s_%s", LOCK_KEY_PREFIX, priorityStr)

		// 如果使用MySQL作为消息队列，则使用分布式锁
		if config.Conf.Common.MySQLAsMq {
			s.locks[priority] = lock.NewRedisLock(lockKey,
				lock.WithExpireSeconds(LOCK_EXPIRE_SECONDS),
				lock.WithWatchDogMode()) // 使用看门狗模式自动续期
		}
		s.isLeader[priority] = false
	}

	// 同时启动高、中、低三个优先级的消费者
	for _, priority := range consumePriority {
		log.Infof("启动%s优先级消息消费者", data.GetPriorityStr(priority))
		go s.startConsumer(priority)
	}
}

// tryBeLeader 尝试成为主节点
func (s *MsgConsume) tryBeLeader(ctx context.Context, priority data.PriorityEnum) bool {
	priorityStr := data.GetPriorityStr(priority)
	redisLock := s.locks[priority]

	// 尝试获取锁
	err := redisLock.Lock(ctx)
	if err != nil {
		log.Infof("%s优先级消费者未能获取到主节点锁: %v", priorityStr, err)
		return false
	}

	log.Infof("%s优先级消费者成功获取主节点锁，成为主消费者", priorityStr)
	return true
}

// startConsumer 启动指定优先级的消费者
// 根据配置决定是从MySQL还是消息队列中消费
func (s *MsgConsume) startConsumer(priority data.PriorityEnum) {
	var consumer mq.Consumer
	priorityStr := data.GetPriorityStr(priority)

	consumer = data.GetData().GetConsumer(priority)

	// 启动消费流程并设置恢复机制
	go func() {
		// 设置恢复机制，如果发生崩溃，尝试重启消费者
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("%s优先级消费者发生崩溃: %v，5秒后尝试重启", priorityStr, r)

				// 如果是leader，在崩溃时尝试释放锁
				if s.isLeader[priority] && s.locks[priority] != nil {
					ctx := context.Background()
					err := s.locks[priority].Unlock(ctx)
					if err != nil {
						log.Errorf("%s优先级消费者崩溃时解锁失败: %v", priorityStr, err)
					} else {
						log.Infof("%s优先级消费者崩溃时成功释放主节点锁", priorityStr)
					}
					s.isLeader[priority] = false
				}

				// 在一段时间后重新启动消费者
				time.Sleep(time.Second * 5)
				go s.startConsumer(priority)
			}
		}()

		// 启动实际的消费流程
		log.Infof("开始消费%s优先级消息", priorityStr)
		if config.Conf.Common.MySQLAsMq {
			// 使用MySQL作为消息中转站时，需要使用分布式锁
			s.consumeFromMySQLWithLock(priority)
		} else {
			s.consumeFromMQ(consumer, priority)
		}
	}()
}

// consumeFromMySQLWithLock 使用分布式锁从MySQL消费消息
func (s *MsgConsume) consumeFromMySQLWithLock(priority data.PriorityEnum) {
	priorityStr := data.GetPriorityStr(priority)
	ctx := context.Background()

	// 确保函数退出时释放锁
	defer func() {
		if s.isLeader[priority] && s.locks[priority] != nil {
			err := s.locks[priority].Unlock(ctx)
			if err != nil {
				log.Errorf("%s优先级消费者函数退出时解锁失败: %v", priorityStr, err)
			} else {
				log.Infof("%s优先级消费者函数退出时成功释放主节点锁", priorityStr)
			}
			s.isLeader[priority] = false
		}
	}()

	// 首先尝试获取锁(消费者启动时，尝试获取锁)
	s.isLeader[priority] = s.tryBeLeader(ctx, priority)

	for {
		if s.isLeader[priority] {
			// 作为主节点，正常消费消息
			log.Debugf("%s优先级消费者作为主节点消费消息", priorityStr)
			s.consumeMySQLMsg(priority)

			// 对于重试队列，使用更长的消费间隔
			var step int64
			if priority == data.PRIORITY_RETRY {
				// 重试队列使用1000-2000ms的随机间隔
				step = RandNum(1000) + 1000
			} else {
				// 其他队列使用0-500ms的随机间隔
				step = RandNum(500)
			}

			internelTime := time.Duration(step) * time.Millisecond
			time.Sleep(internelTime)
		} else {
			// 作为备用节点，定期尝试获取锁
			log.Debugf("%s优先级消费者作为备用节点，等待成为主节点", priorityStr)
			time.Sleep(time.Second * LOCK_RETRY_INTERVAL_SECONDS)
			s.isLeader[priority] = s.tryBeLeader(ctx, priority)

			if s.isLeader[priority] {
				log.Infof("%s优先级消费者从备用节点升级为主节点", priorityStr)
			}
		}
	}
}

// consumeFromMQ 从消息队列中消费消息并处理
func (s *MsgConsume) consumeFromMQ(consumer mq.Consumer, priority data.PriorityEnum) {
	priorityStr := data.GetPriorityStr(priority)
	forNum := 1
	if priority == data.PRIORITY_HIGH {
		forNum = 6
	} else if priority == data.PRIORITY_MIDDLE {
		forNum = 3
	} else if priority == data.PRIORITY_LOW {
		forNum = 1
	} else if priority == data.PRIORITY_RETRY {
		forNum = 1
	}
	for i := 0; i < forNum; i++ {
		// 使用匿名函数启动一个新的 goroutine
		go func() {
			log.Infof("🚀 启动%s优先级消费者goroutine", priorityStr)
			// 消费消息
			consumer.ConsumeMessages(func(message []byte) error {
				// 创建一个新的上下文
				ctx := context.Background()
				// 记录日志
				log.InfoContextf(ctx, "📨 [%s] 收到消息: %s", priorityStr, string(message))

				// 创建一个新的 SendMsgReq 实例
				var req = new(ctrlmodel.SendMsgReq)
				// 反序列化消息
				err := json.Unmarshal(message, &req)
				if err != nil {
					log.ErrorContextf(ctx, "❌ [%s] 消息反序列化失败: %s, 原始消息: %s", priorityStr, err.Error(), string(message))
					return err
				}
				log.InfoContextf(ctx, "✅ [%s] 消息反序列化成功，MsgID: %s, To: %s, TemplateID: %s", priorityStr, req.MsgID, req.To, req.TemplateID)

				// 处理消息
				log.InfoContextf(ctx, "🔄 [%s] 开始处理消息，MsgID: %s", priorityStr, req.MsgID)
				err = dealOneMsg(ctx, req)
				if err != nil {
					log.ErrorContextf(ctx, "❌ [%s] 消息处理失败，MsgID: %s, 错误: %s", priorityStr, req.MsgID, err.Error())
					// 进入重试
					return s.handleMqRetryAfterFailure(ctx, req, message, priorityStr)
				}
				log.InfoContextf(ctx, "✅ [%s] 消息处理成功，MsgID: %s", priorityStr, req.MsgID)
				return nil
			})
		}()
	}
}

// handleMqRetryAfterFailure 处理mq消息处理失败后的重试逻辑
func (s *MsgConsume) handleMqRetryAfterFailure(ctx context.Context, req *ctrlmodel.SendMsgReq, message []byte, priorityStr string) error {
	// 获取数据实例
	dt := data.GetData()

	// 增加重试次数并检查是否达到上限
	newCount, retryErr := data.MsgRecordNsp.IncrementRetryCount(dt.GetDB(), req.MsgID)
	if retryErr != nil {
		log.Errorf("更新重试次数失败: %s", retryErr.Error())
		// 即使更新失败也要继续重试
	}

	// 检查重试次数是否达到上限
	if newCount >= config.Conf.Common.MaxRetryCount {
		log.Infof("消息 %s 已达到最大重试次数 %d，不再重试",
			req.MsgID, config.Conf.Common.MaxRetryCount)
		// 更新消息状态为最终失败
		data.MsgRecordNsp.UpdateStatus(dt.GetDB(), req.MsgID, int(data.MSG_STATUS_FAILED))
		// 更新队列状态为最终失败
		data.MsgQueueNsp.SetStatus(dt.GetDB(), priorityStr, req.MsgID, int(data.TASK_STATUS_FAILED))
		return nil
	}

	log.InfoContextf(ctx, "消息 %s 当前重试次数: %d/%d，加入重试队列",
		req.MsgID, newCount, config.Conf.Common.MaxRetryCount)
	// 扔进重试主题处理
	data.GetData().GetRetryMQProducer().SendMessage(message)
	return nil // 返回nil，避免消息被重复消费
}

// dealOneMsg 处理一条消息
func dealOneMsg(ctx context.Context, req *ctrlmodel.SendMsgReq) error {
	log.InfoContextf(ctx, "🔍 开始处理消息，MsgID: %s, TemplateID: %s", req.MsgID, req.TemplateID)

	// 获取数据实例
	dt := data.GetData()

	log.InfoContextf(ctx, "📋 获取消息模板，TemplateID: %s", req.TemplateID)
	tp, err := dt.GetMsgTemplate(ctx, req.TemplateID)
	if err != nil {
		log.ErrorContextf(ctx, "❌ 获取消息模板失败: %s", err.Error())
		return err
	}
	log.InfoContextf(ctx, "✅ 获取消息模板成功，Channel: %d, Subject: %s, Content长度: %d", tp.Channel, tp.Subject, len(tp.Content))

	// 替换模板中的变量
	var content string
	if tp.Channel == int(data.Channel_EMAIL) || tp.Channel == int(data.Channel_LARK) {
		log.InfoContextf(ctx, "🔄 开始模板变量替换，原内容: %s", tp.Content)
		content, err = tools.TemplateReplace(tp.Content, req.TemplateData)
		if err != nil {
			log.ErrorContextf(ctx, "❌ 模板变量替换失败: %s", err.Error())
			return err
		}
		log.InfoContextf(ctx, "✅ 模板变量替换成功，替换后内容: %s", content)
	}

	// 根据通道类型获取消息处理器
	log.InfoContextf(ctx, "🔍 查找消息处理器，Channel: %d", tp.Channel)
	handler, ok := msgProcMap[tp.Channel]
	if !ok {
		log.ErrorContextf(ctx, "❌ 不支持的渠道类型: %d", tp.Channel)
		return errors.New("channel not support")
	}
	log.InfoContextf(ctx, "✅ 找到消息处理器，Channel: %d", tp.Channel)

	// 创建消息处理器实例
	t := handler.NewProc()
	// 设置消息处理器的基本信息
	t.Base().To = req.To
	t.Base().Subject = tp.Subject
	t.Base().Content = content
	t.Base().Priority = req.Priority
	t.Base().TemplateID = req.TemplateID
	t.Base().TemplateData = req.TemplateData

	log.InfoContextf(ctx, "📧 准备发送消息，To: %s, Subject: %s, Content: %s", req.To, tp.Subject, content)

	// 发送消息
	err = t.SendMsg()
	if err != nil {
		log.ErrorContextf(ctx, "❌ 发送消息失败: %s", err.Error())
		return err
	}
	// 使用通用函数创建或更新消息记录
	// 如果记录存在则更新状态，如果不存在则创建新记录
	err = tools.CreateOrUpdateMsgRecord(dt.GetDB(), req.MsgID, req, tp, int(data.MSG_STATUS_SUCC))
	if err != nil {
		log.ErrorContextf(ctx, "创建或更新消息记录失败: %s", err.Error())
		// 消息记录操作失败不应影响消息队列状态更新
	}

	// 更新消息状态为成功
	priorityStr := data.GetPriorityStr(data.PriorityEnum(req.Priority))
	if config.Conf.Common.MySQLAsMq {
		err = data.MsgQueueNsp.SetStatus(dt.GetDB(), priorityStr, req.MsgID, int(data.TASK_STATUS_SUCC))
		if err != nil {
			log.ErrorContextf(ctx, "update msg status to success err %s", err.Error())
			// 更新状态失败不应该影响整个处理流程，所以这里只记录日志，不返回错误
		}
	}
	log.InfoContextf(ctx, "消息 %s 已成功处理并更新状态", req.MsgID)
	return nil
}

// dealRetryMysqlQueue 将消息发送到重试队列
func dealRetryMysqlQueue(db *gorm.DB, req *ctrlmodel.SendMsgReq) error {

	// 增加重试次数
	newCount, retryErr := data.MsgRecordNsp.IncrementRetryCount(db, req.MsgID)
	if retryErr != nil {
		log.Errorf("更新重试次数失败: %s", retryErr.Error())
		// 即使更新失败也要继续重试
	} else {
		log.Infof("消息 %s 重试次数已更新为: %d (最大重试次数: %d)",
			req.MsgID, newCount, config.Conf.Common.MaxRetryCount)
	}

	// 检查重试次数是否达到上限
	if newCount >= config.Conf.Common.MaxRetryCount {
		log.Infof("消息 %s 已达到最大重试次数 %d，不再重试",
			req.MsgID, config.Conf.Common.MaxRetryCount)
		// 更新消息状态为最终失败
		data.MsgRecordNsp.UpdateStatus(db, req.MsgID, int(data.MSG_STATUS_FAILED))
		// 更新队列状态为最终失败
		priorityStr := data.GetPriorityStr(data.PriorityEnum(req.Priority))
		data.MsgQueueNsp.SetStatus(db, priorityStr, req.MsgID, int(data.TASK_STATUS_FAILED))
		return nil
	}

	// 生成唯一的消息ID
	msgID := req.MsgID

	// 检查消息是否已存在于重试队列
	retryPriorityStr := data.GetPriorityStr(data.PRIORITY_RETRY)
	existingMsg, err := data.MsgQueueNsp.Find(db, retryPriorityStr, msgID)

	if existingMsg != nil && existingMsg.ID != 0 && err != gorm.ErrRecordNotFound {
		log.Infof("消息 %s 已存在于重试队列，更新状态为待处理", msgID)
		// 更新状态为待处理，而不是创建新记录
		return data.MsgQueueNsp.SetStatus(db, retryPriorityStr, msgID, int(data.TASK_STATUS_PENDING))
	}

	// 创建一个新的消息队列实例
	var md = new(data.MsgQueue)

	// 设置消息的主题
	md.Subject = req.Subject

	// 设置消息的模板ID
	md.TemplateID = req.TemplateID

	// 将模板数据转换为JSON格式
	td, err := json.Marshal(req.TemplateData)
	if err != nil {
		return err
	}

	// 设置消息的模板数据
	md.TemplateData = string(td)

	// 设置消息的接收者
	md.To = req.To

	// 设置消息的ID
	md.MsgId = msgID

	// 设置消息的状态为待处理
	md.Status = int(data.TASK_STATUS_PENDING)

	// 设置消息的优先级为重试优先级
	md.Priority = int(data.PRIORITY_RETRY)

	// 将消息插入到MySQL重试队列表中
	err = data.MsgQueueNsp.Create(db, retryPriorityStr, md)
	if err != nil {
		// 处理可能的重复键错误
		if strings.Contains(err.Error(), "Duplicate entry") {
			log.Warnf("消息 %s 在重试队列中已存在（并发处理）", msgID)
			return nil
		}
		return err
	}

	log.Infof("消息 %s 已加入MySQL重试队列，当前重试次数: %d/%d",
		msgID, newCount, config.Conf.Common.MaxRetryCount)
	return nil
}

// consumeFromMySQL 从MySQL消费消息（不使用分布式锁，已废弃）
func (s *MsgConsume) consumeFromMySQL(priority data.PriorityEnum) {
	priorityStr := data.GetPriorityStr(priority)
	log.Infof("开始从MySQL消费%s优先级消息", priorityStr)

	for {
		// 对于重试队列，使用更长的消费间隔
		var step int64
		if priority == data.PRIORITY_RETRY {
			// 重试队列使用1000-2000ms的随机间隔
			step = RandNum(1000) + 1000
		} else {
			// 其他队列使用0-500ms的随机间隔
			step = RandNum(500)
		}

		internelTime := time.Duration(step) * time.Millisecond
		t := time.NewTimer(internelTime)
		// 等待定时器触发
		<-t.C
		// 消费MySQL消息
		s.consumeMySQLMsg(priority)
	}
}

func (s *MsgConsume) consumeMySQLMsg(priority data.PriorityEnum) {

	dt := data.GetData()
	priorityStr := data.GetPriorityStr(priority)
	pullNum := 60
	if priority == data.PRIORITY_MIDDLE {
		pullNum = 30
	} else if priority == data.PRIORITY_LOW {
		pullNum = 10
	}
	msgList, err := data.MsgQueueNsp.GetMsgList(dt.GetDB(),
		priorityStr, int(data.TASK_STATUS_PENDING), pullNum)
	// 如果获取消息列表时发生错误，则返回
	if err != nil {
		return
	}
	// 创建一个字符串切片，用于存储消息ID
	msgIdList := make([]string, len(msgList))
	// 遍历消息列表，将每个消息的ID添加到msgIdList中
	for i, dbMsg := range msgList {
		msgIdList[i] = dbMsg.MsgId
	}
	// 如果msgIdList不为空，则批量设置消息状态为处理中
	if len(msgIdList) != 0 {
		err = data.MsgQueueNsp.BatchSetStatus(dt.GetDB(), priorityStr, msgIdList,
			int(data.TASK_STATUS_PROCESSING))
		// 如果批量设置消息状态时发生错误，则返回
		if err != nil {
			return
		}
	}
	ctx := context.Background()
	// 遍历消息列表，处理每个消息
	for _, dbMsg := range msgList {
		// 创建一个新的SendMsgReq实例
		var req = new(ctrlmodel.SendMsgReq)
		req.MsgID = dbMsg.MsgId
		req.Priority = dbMsg.Priority
		// 设置消息的接收者
		req.To = dbMsg.To
		// 设置消息的主题
		req.Subject = dbMsg.Subject
		// 设置消息的模板ID
		req.TemplateID = dbMsg.TemplateID
		// 反序列化消息的模板数据
		req.TemplateData = make(map[string]string, 0)
		err := json.Unmarshal([]byte(dbMsg.TemplateData), &req.TemplateData)
		// 如果反序列化时发生错误，则返回
		if err != nil {
			log.ErrorContextf(ctx, "unmarshal template data err %s", err.Error())
			return
		}
		// 处理单个消息
		if err := dealOneMsg(ctx, req); err != nil {
			// 如果处理失败，则将消息发送到重试队列
			log.ErrorContextf(ctx, "处理消息 %s 失败，准备加入重试队列: %s", req.MsgID, err.Error())

			if err := dealRetryMysqlQueue(dt.GetDB(), req); err != nil {
				log.ErrorContextf(ctx, "发送消息 %s 到重试队列失败: %s", req.MsgID, err.Error())
				return
			}
		}
	}
}

// RandNum func for rand num
func RandNum(num int64) int64 {
	step := rand.Int63n(num) + int64(1)
	flag := rand.Int63n(2)
	if flag == 0 {
		return -step
	}
	return step
}

// UnlockAll 释放所有持有的分布式锁
func (s *MsgConsume) UnlockAll() {
	ctx := context.Background()
	for priority, isLeader := range s.isLeader {
		// 只需要解锁自己是leader的锁
		if isLeader && s.locks[priority] != nil {
			priorityStr := data.GetPriorityStr(priority)
			err := s.locks[priority].Unlock(ctx)
			if err != nil {
				log.Errorf("%s优先级消费者解锁失败: %v", priorityStr, err)
			} else {
				log.Infof("%s优先级消费者成功释放主节点锁", priorityStr)
			}
			// 更新状态
			s.isLeader[priority] = false
		}
	}
}
