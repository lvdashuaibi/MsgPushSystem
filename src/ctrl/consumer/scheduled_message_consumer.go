package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lvdashuaibi/MsgPushSystem/src/ctrl/ctrlmodel"
	"github.com/lvdashuaibi/MsgPushSystem/src/data"
	"github.com/lvdashuaibi/MsgPushSystem/src/pkg/utils"
	"github.com/redis/go-redis/v9"
)

// ScheduledMessageConsumer 定时消息消费者
type ScheduledMessageConsumer struct {
	stopChan   chan bool
	logger     *ScheduledLogger     // 控制台日志（保留用于重要信息）
	fileLogger *ScheduledFileLogger // 文件日志（详细的Redis轮询日志）
}

// NewScheduledMessageConsumer 创建定时消息消费者
func NewScheduledMessageConsumer() *ScheduledMessageConsumer {
	return &ScheduledMessageConsumer{
		stopChan:   make(chan bool),
		logger:     NewScheduledLogger(),
		fileLogger: NewScheduledFileLogger(),
	}
}

// Start 启动定时消息调度器
func (s *ScheduledMessageConsumer) Start() {
	// 控制台显示重要信息
	s.logger.LogSchedulerStart()
	// 文件记录详细信息
	if s.fileLogger != nil {
		s.fileLogger.LogSchedulerStart()
	}

	go func() {
		ticker := time.NewTicker(10 * time.Second) // 每10秒检查一次
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.processScheduledMessages()
			case <-s.stopChan:
				s.logger.LogSchedulerStop()
				if s.fileLogger != nil {
					s.fileLogger.LogSchedulerStop()
					s.fileLogger.Close()
				}
				return
			}
		}
	}()
}

// Stop 停止定时消息调度器
func (s *ScheduledMessageConsumer) Stop() {
	close(s.stopChan)
}

// processScheduledMessages 处理到期的定时消息（Redis轮询）
func (s *ScheduledMessageConsumer) processScheduledMessages() {
	dt := data.GetData()
	ctx := context.Background()

	// 文件日志记录详细扫描信息
	if s.fileLogger != nil {
		s.fileLogger.LogScanStart()
	}

	// 使用Redis ZSET查找到期的定时消息
	loc, _ := time.LoadLocation("Asia/Shanghai")
	now := time.Now().In(loc)
	currentScore := float64(now.Unix())

	// 从Redis ZSET中获取到期的消息ID（分数 <= 当前时间戳）
	scheduleIDs, err := dt.GetCache().ZRangeByScore(ctx, "Scheduled_Messages", &redis.ZRangeBy{
		Min: "0",
		Max: fmt.Sprintf("%.0f", currentScore),
	})
	if err != nil {
		s.logger.Error("从Redis获取待发送定时消息失败: %s", err.Error())
		if s.fileLogger != nil {
			s.fileLogger.Error("从Redis获取待发送定时消息失败: %s", err.Error())
		}
		return
	}

	// 文件日志记录详细扫描结果
	if s.fileLogger != nil {
		s.fileLogger.LogScanResult(len(scheduleIDs))
		s.fileLogger.LogRedisZSetScan(currentScore, scheduleIDs)
	}

	// 控制台只在有消息时显示
	if len(scheduleIDs) > 0 {
		s.logger.LogScanResult(len(scheduleIDs))
	}

	for _, scheduleID := range scheduleIDs {
		// 从数据库获取完整的消息信息
		message, err := data.ScheduledMessageNamespace.FindByScheduleID(dt.GetDB(), scheduleID)
		if err != nil {
			s.logger.Error("从数据库获取定时消息详情失败: %s, ID: %s", err.Error(), scheduleID)
			if s.fileLogger != nil {
				s.fileLogger.Error("从数据库获取定时消息详情失败: %s, ID: %s", err.Error(), scheduleID)
			}
			// 从Redis中移除无效的消息ID
			dt.GetCache().ZRem(ctx, "Scheduled_Messages", scheduleID)
			continue
		}

		// 检查消息状态，只处理待发送的消息
		if message.Status != int(data.ScheduledMessageStatusPending) {
			if s.fileLogger != nil {
				s.fileLogger.Debug("跳过非待发送状态的消息: %s, 状态: %d", scheduleID, message.Status)
			}
			s.logger.Debug("跳过非待发送状态的消息: %s, 状态: %d", scheduleID, message.Status)
			// 从Redis中移除已处理的消息
			dt.GetCache().ZRem(ctx, "Scheduled_Messages", scheduleID)
			continue
		}
		// 添加时间比较调试（只记录到文件）
		if s.fileLogger != nil {
			s.fileLogger.LogTimeComparison(message.ScheduledTime, now)
		}

		// 添加时间比较调试
		s.logger.LogTimeComparison(message.ScheduledTime, now)
		s.processOneScheduledMessage(ctx, message)
	}
}

// processOneScheduledMessage 处理单条定时消息
func (s *ScheduledMessageConsumer) processOneScheduledMessage(ctx context.Context, message *data.ScheduledMessage) {
	dt := data.GetData()

	s.logger.LogMessageProcessStart(message.ScheduleID, message.ScheduledTime)

	// 更新状态为处理中（防止重复处理）
	err := data.ScheduledMessageNamespace.UpdateStatus(dt.GetDB(), message.ScheduleID, data.ScheduledMessageStatusSent)
	if err != nil {
		s.logger.Error("更新定时消息状态失败: %s", err.Error())
		return
	}
	s.logger.LogStatusUpdate(message.ScheduleID, "已发送")

	// 解析模板数据
	var templateData map[string]string
	if message.TemplateData != "" {
		if err := json.Unmarshal([]byte(message.TemplateData), &templateData); err != nil {
			s.logger.Error("解析模板数据失败: %s", err.Error())
			data.ScheduledMessageNamespace.UpdateStatus(dt.GetDB(), message.ScheduleID, data.ScheduledMessageStatusFailed)
			s.logger.LogStatusUpdate(message.ScheduleID, "发送失败")
			return
		}
	}

	// 获取目标用户列表
	var targetUsers []*data.User

	// 如果指定了用户ID，直接获取用户
	if len(message.UserIDs) > 0 {
		for _, userID := range message.UserIDs {
			user, userErr := data.UserNamespace.FindByUserID(dt.GetDB(), userID)
			if userErr != nil {
				s.logger.Error("查找用户失败: %s, 用户ID: %s", userErr.Error(), userID)
				continue
			}
			targetUsers = append(targetUsers, user)
		}
	}

	// 如果指定了标签，根据标签查找用户
	if len(message.Tags) > 0 {
		tagUsers, tagErr := data.UserNamespace.FindByAnyTags(dt.GetDB(), message.Tags)
		if tagErr != nil {
			s.logger.Error("根据标签查找用户失败: %s", tagErr.Error())
		} else {
			// 合并用户列表，去重
			userMap := make(map[string]*data.User)
			for _, user := range targetUsers {
				userMap[user.UserID] = user
			}
			for _, user := range tagUsers {
				userMap[user.UserID] = user
			}

			targetUsers = make([]*data.User, 0, len(userMap))
			for _, user := range userMap {
				targetUsers = append(targetUsers, user)
			}
		}
	}

	s.logger.LogUserResolution(message.ScheduleID, len(targetUsers))

	// 为每个用户发送消息
	successCount := 0
	for _, user := range targetUsers {
		// 根据用户的联系方式选择发送渠道
		var to string
		if user.Email != "" {
			to = user.Email
		} else if user.Mobile != "" {
			to = user.Mobile
		} else if user.LarkID != "" {
			to = user.LarkID
		} else {
			s.logger.Warn("用户 %s 没有有效的联系方式", user.UserID)
			continue
		}

		// 创建发送消息请求
		sendReq := &ctrlmodel.SendMsgReq{
			To:           to,
			TemplateID:   message.TemplateID,
			TemplateData: templateData,
			Priority:     2, // 中等优先级
		}

		// 发送消息到队列
		sendErr := s.sendMessageToQueue(sendReq)
		if sendErr != nil {
			s.logger.LogSendError(user.UserID, sendErr)
		} else {
			s.logger.LogSendToQueue(user.UserID, to)
			successCount++
		}
	}

	s.logger.LogMessageProcessSuccess(message.ScheduleID, successCount, len(targetUsers))

	// 从Redis定时队列中移除
	s.logger.LogRedisOperation("ZREM", message.ScheduleID)
	dt.GetCache().ZRem(ctx, "Scheduled_Messages", message.ScheduleID)
}

// sendMessageToQueue 发送消息到队列
func (s *ScheduledMessageConsumer) sendMessageToQueue(req *ctrlmodel.SendMsgReq) error {
	dt := data.GetData()

	// 生成消息ID
	req.MsgID = utils.GenerateUUID()

	// 序列化消息
	msgJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}

	// 发送到中等优先级队列
	producer := dt.GetProducer(data.PRIORITY_MIDDLE)
	return producer.SendMessage("", msgJSON)
}
