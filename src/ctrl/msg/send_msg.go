package msg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/BitofferHub/msgcenter/src/config"
	"github.com/BitofferHub/msgcenter/src/constant"
	"github.com/BitofferHub/msgcenter/src/ctrl/ctrlmodel"
	"github.com/BitofferHub/msgcenter/src/ctrl/handler"
	"github.com/BitofferHub/msgcenter/src/ctrl/tools"
	"github.com/BitofferHub/msgcenter/src/data"
	"github.com/BitofferHub/pkg/middlewares/log"
	"github.com/BitofferHub/pkg/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SendMsgHandler 接口处理handler
type SendMsgHandler struct {
	Req    ctrlmodel.SendMsgReq
	Resp   ctrlmodel.SendMsgResp
	UserId string
}

// SendMsg 接口
func SendMsg(c *gin.Context) {
	var hd SendMsgHandler
	defer func() {
		hd.Resp.Msg = constant.GetErrMsg(hd.Resp.Code)
		c.JSON(http.StatusOK, hd.Resp)
	}()
	// 获取用户Id
	hd.UserId = c.Request.Header.Get(constant.HEADER_USERID)
	// 解析请求包
	if err := c.ShouldBind(&hd.Req); err != nil {
		log.Errorf("SendMsg shouldBind err %s", err.Error())
		hd.Resp.Code = constant.ERR_SHOULD_BIND
		return
	}
	// 执行处理函数, 这里会调用对应的HandleInput和HandleProcess，往下看
	if err := handler.Run(&hd); err != nil {
		log.Errorf("SendMsg handler.Run err %s", err.Error())
		// 如果Resp.Code未设置，设置为内部错误
		if hd.Resp.Code == 0 {
			hd.Resp.Code = constant.ERR_INTERNAL
		}
	}
}

// HandleInput 参数检查
func (p *SendMsgHandler) HandleInput() error {
	if p.Req.TemplateID == "" {
		p.Resp.Code = constant.ERR_INPUT_INVALID
		return nil
	}
	if p.Req.TemplateData == nil {
		p.Resp.Code = constant.ERR_INPUT_INVALID
		return nil
	}

	// 验证至少有一种接收者类型
	if p.Req.To == "" && len(p.Req.UserIDs) == 0 && len(p.Req.Tags) == 0 {
		p.Resp.Code = constant.ERR_INPUT_INVALID
		return nil
	}

	if p.Req.Priority == 0 {
		p.Req.Priority = int(data.PRIORITY_LOW)
	}
	return nil
}

// HandleProcess 处理函数
func (p *SendMsgHandler) HandleProcess() error {
	sourceID := p.UserId
	ctx := context.Background()
	log.Infof("into HandleProcess")
	dt := data.GetData()

	// 获取消息模板
	mt, err := dt.GetMsgTemplate(ctx, p.Req.TemplateID)
	if err != nil {
		log.Errorf("get msg template err %s", err.Error())
		p.Resp.Code = constant.ERR_TEMPLATE_NOT_READY
		return err
	}

	// 模板状态检查
	if mt.Status != int(data.TEMPLATE_STATUS_NORMAL) {
		p.Resp.Code = constant.ERR_TEMPLATE_NOT_READY
		return errors.New("template not ready")
	}

	// 解析接收者列表
	recipients, err := p.parseRecipients()
	if err != nil {
		log.Errorf("parse recipients err %s", err.Error())
		p.Resp.Code = constant.ERR_INPUT_INVALID
		return err
	}

	if len(recipients) == 0 {
		p.Resp.Code = constant.ERR_INPUT_INVALID
		return errors.New("no valid recipients found")
	}

	// 批量发送消息
	var successCount int
	var msgIDs []string

	for _, recipient := range recipients {
		// 为每个接收者创建单独的消息请求
		msgReq := p.Req
		msgReq.To = recipient

		msgID, err := p.sendSingleMessage(&msgReq, mt, sourceID)
		if err != nil {
			log.Errorf("send message to %s failed: %s", recipient, err.Error())
			continue
		}

		successCount++
		msgIDs = append(msgIDs, msgID)
	}

	if successCount == 0 {
		p.Resp.Code = constant.ERR_SEND_MSG
		return errors.New("all messages failed to send")
	}

	// 返回第一个消息ID作为主ID，实际场景中可能需要返回所有ID
	p.Resp.MsgID = msgIDs[0]
	log.Infof("batch send completed: %d/%d messages sent successfully", successCount, len(recipients))
	return nil
}

// parseRecipients 解析接收者列表
func (p *SendMsgHandler) parseRecipients() ([]string, error) {
	var recipients []string
	dt := data.GetData()

	// 1. 直接指定的接收者
	if p.Req.To != "" {
		recipients = append(recipients, p.Req.To)
	}

	// 2. 按用户ID指定的接收者
	if len(p.Req.UserIDs) > 0 {
		for _, userID := range p.Req.UserIDs {
			user, err := data.UserNamespace.FindByUserID(dt.GetDB(), userID)
			if err != nil {
				log.Warnf("user %s not found: %s", userID, err.Error())
				continue
			}

			// 根据模板渠道选择合适的联系方式
			recipient := p.getRecipientByChannel(user)
			if recipient != "" {
				recipients = append(recipients, recipient)
			}
		}
	}

	// 3. 按标签指定的接收者
	if len(p.Req.Tags) > 0 {
		users, err := data.UserNamespace.FindByAnyTags(dt.GetDB(), p.Req.Tags)
		if err != nil {
			return nil, fmt.Errorf("find users by tags failed: %w", err)
		}

		for _, user := range users {
			recipient := p.getRecipientByChannel(user)
			if recipient != "" {
				recipients = append(recipients, recipient)
			}
		}
	}

	// 去重
	recipients = p.deduplicateRecipients(recipients)
	return recipients, nil
}

// getRecipientByChannel 根据模板渠道获取用户的联系方式
func (p *SendMsgHandler) getRecipientByChannel(user *data.User) string {
	ctx := context.Background()
	dt := data.GetData()

	// 获取模板信息以确定渠道
	mt, err := dt.GetMsgTemplate(ctx, p.Req.TemplateID)
	if err != nil {
		log.Errorf("get template failed: %s", err.Error())
		return ""
	}

	switch mt.Channel {
	case 1: // 邮件
		return user.Email
	case 2: // 短信
		return user.Mobile
	case 3: // 飞书
		return user.LarkID
	default:
		// 默认优先级：邮箱 > 手机号 > 飞书ID
		if user.Email != "" {
			return user.Email
		}
		if user.Mobile != "" {
			return user.Mobile
		}
		return user.LarkID
	}
}

// deduplicateRecipients 去重接收者列表
func (p *SendMsgHandler) deduplicateRecipients(recipients []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, recipient := range recipients {
		if recipient != "" && !seen[recipient] {
			seen[recipient] = true
			result = append(result, recipient)
		}
	}

	return result
}

// sendSingleMessage 发送单条消息
func (p *SendMsgHandler) sendSingleMessage(msgReq *ctrlmodel.SendMsgReq, mt *data.MsgTemplate, sourceID string) (string, error) {
	ctx := context.Background()
	dt := data.GetData()

	// 获取配额
	var (
		limit, div int
		ready      bool
	)

	quatoCacheKey := fmt.Sprintf("%s%s%d", data.REDIS_KEY_SOURCE_QUOTA, mt.SourceID, mt.Channel)

	// 如果缓存开启，则从缓存中获取配额
	if config.Conf.Common.OpenCache {
		limitdiv, _, _ := dt.GetCache().Get(ctx, quatoCacheKey)
		if len(limitdiv) > 0 {
			ary := strings.Split(limitdiv, "_")
			limit, _ = strconv.Atoi(ary[0])
			div, _ = strconv.Atoi(ary[1])
			log.Infof("quota cache hit %d, %d", limit, div)
			ready = true
		}
	}

	// 如果缓存未命中，则从数据库中获取配额
	if !ready {
		log.Infof("quota cache miss")
		// 获取全局配额
		globalQuota, err := data.GlobalQuotaNsp.Find(dt.GetDB(), mt.Channel)
		if err != nil {
			return "", err
		}
		limit = globalQuota.Num
		div = globalQuota.Unit
		// 获取业务配额
		sourceQuota, err := data.SourceQuotaNsp.Find(dt.GetDB(), sourceID, mt.Channel)
		if err != nil {
			if err != gorm.ErrRecordNotFound {
				return "", err
			}
		} else {
			limit = sourceQuota.Num
			div = sourceQuota.Unit
		}
		value := fmt.Sprintf("%d_%d", limit, div)
		if config.Conf.Common.OpenCache {
			dt.GetCache().Set(ctx, quatoCacheKey, value, 30*time.Second)
		}
	}
	log.Infof("limit %d, div %d", limit, div)

	// 创建限流器
	lm := tools.NewRateLimiter(dt.GetCache().GetRedisBaseConn(), div, limit)
	keyID := fmt.Sprintf(data.REDIS_KEY_RATE_LIMIT_COUNT+":%s:%d", sourceID, mt.Channel)
	if msgReq.SendTimestamp > 0 {
		// 定时消息单独计数限频
		keyID = fmt.Sprintf(data.REDIS_KEY_RATE_LIMIT_COUNT_TIMER+":%s:%d", sourceID, mt.Channel)
	}

	// 判断用户的请求是否被允许
	allowed, err := lm.IsRequestAllowed(keyID)
	if err != nil {
		log.Errorf("IsRequestAllowed err %s", err.Error())
		return "", err
	}
	if !allowed {
		log.Infof("request denied for recipient %s", msgReq.To)
		return "", errors.New("request limit exceeded")
	}

	// 定时消息
	if msgReq.SendTimestamp > 0 {
		return p.sendSingleToTimer(msgReq)
	}

	// 确保消息在响应前持久化
	var msgErr error
	var msgID string
	if config.Conf.Common.MySQLAsMq {
		msgID, msgErr = p.sendSingleToMySQL(msgReq)
	} else {
		msgID, msgErr = p.sendSingleToMQ(msgReq)
	}

	status := int(data.MSG_STATUS_PENDING)
	if msgErr != nil {
		status = int(data.MSG_STATUS_FAILED)
	}

	// 创建消息记录，使用户可以立即查询到消息状态
	err = tools.CreateMsgRecord(dt.GetDB(), msgID, msgReq, mt, status)
	if err != nil {
		log.Errorf("创建消息记录失败：%s", err.Error())
		// 即使创建消息记录失败，我们也已经发送了消息到MQ，继续
	}

	// 如果持久化出错，则返回错误
	if msgErr != nil {
		log.Errorf("消息持久化失败: %s", msgErr.Error())
		return "", msgErr
	}

	log.Infof("消息 %s 已成功持久化", msgID)
	return msgID, nil
}

// sendSingleToTimer 将单条消息发送到定时队列
func (p *SendMsgHandler) sendSingleToTimer(msgReq *ctrlmodel.SendMsgReq) (string, error) {
	// 获取数据实例
	log.Infof("into sendSingleToTimer")
	dt := data.GetData()
	ctx := context.Background()

	// 生成唯一的消息ID
	msgID := utils.NewUuid()

	// 将消息ID赋值给请求结构体
	msgReq.MsgID = msgID

	// 将请求结构体转换为JSON格式
	msgJson, err := json.Marshal(msgReq)
	if err != nil {
		// 记录错误日志
		log.ErrorContextf(context.Background(), "json marshal err %s", err.Error())
		return "", err
	}

	// 根据消息优先级选择对应的定时队列 Todo 是否处理优先级？
	// 存入 MySQL 临时队列；
	// 创建一个新的消息队列实例
	var md = new(data.MsgTmpQueueTimer)

	// 设置消息的发送时间
	md.SendTimestamp = msgReq.SendTimestamp

	// 设置消息
	md.Req = string(msgJson)

	// 设置消息的ID
	md.MsgId = msgID

	// 设置消息的初始状态
	md.Status = int(data.TIMER_MSG_STATUS_PENDING)

	// 将消息插入到MySQL数据库中
	err = data.MsgTmpQueueTimerNsp.Create(dt.GetDB(), md)
	if err != nil {
		return "", err
	}

	// 存入 ZSET；
	timeSocre := float64(msgReq.SendTimestamp)
	member := fmt.Sprintf("%d", msgReq.SendTimestamp)
	_, err = dt.GetCache().ZAdd(ctx, "Timer_Msgs", timeSocre, member)
	if err != nil {
		return "", err
	}

	// 返回消息ID，表示发送成功
	return msgID, nil
}

// sendSingleToMySQL 将单条消息发送到MySQL数据库
func (p *SendMsgHandler) sendSingleToMySQL(msgReq *ctrlmodel.SendMsgReq) (string, error) {
	// 获取数据实例
	dt := data.GetData()

	// 生成唯一的消息ID
	msgID := utils.NewUuid()

	// 创建一个新的消息队列实例
	var md = new(data.MsgQueue)

	// 设置消息的主题
	md.Subject = msgReq.Subject

	// 设置消息的模板ID
	md.TemplateID = msgReq.TemplateID

	// 将模板数据转换为JSON格式
	td, err := json.Marshal(msgReq.TemplateData)
	if err != nil {
		return "", err
	}

	// 设置消息的模板数据
	md.TemplateData = string(td)

	// 设置消息的接收者
	md.To = msgReq.To

	// 设置消息的ID
	md.MsgId = msgID

	// 设置消息的初始状态为待处理状态
	// 消息状态流转: PENDING -> PROCESSING -> SUCC
	// 1. 初始状态为PENDING，表示消息已持久化，等待消费
	// 2. 消费者获取消息后，将状态更新为PROCESSING，表示消息正在处理中
	// 3. 消息成功处理后，将状态更新为SUCC，表示消息处理完成
	md.Status = int(data.TASK_STATUS_PENDING)

	// 设置消息的优先级
	md.Priority = msgReq.Priority

	// 获取消息优先级字符串
	priorityStr := data.GetPriorityStr(data.PriorityEnum(msgReq.Priority))

	// 将消息插入到MySQL数据库中
	err = data.MsgQueueNsp.Create(dt.GetDB(), priorityStr, md)
	if err != nil {
		return "", err
	}

	// 返回消息ID，表示发送成功
	return msgID, nil
}

// sendSingleToMQ 将单条消息发送到消息队列
func (p *SendMsgHandler) sendSingleToMQ(msgReq *ctrlmodel.SendMsgReq) (string, error) {
	// 获取数据实例
	log.Infof("into sendSingleToMQ")
	dt := data.GetData()

	// 生成唯一的消息ID
	msgID := utils.NewUuid()

	// 将消息ID赋值给请求结构体
	msgReq.MsgID = msgID

	// 将请求结构体转换为JSON格式
	msgJson, err := json.Marshal(msgReq)
	if err != nil {
		// 记录错误日志
		log.ErrorContextf(context.Background(), "json marshal err %s", err.Error())
		return "", err
	}

	// 消息队列处理流程：
	// 1. 消息发送到MQ并持久化
	// 2. 消费者从MQ获取消息并处理
	// 3. 处理成功后，消费者会更新MySQL中的消息状态为成功

	var sendErr error

	// 根据消息优先级选择对应的消息队列生产者
	producer := dt.GetProducer(data.PriorityEnum(msgReq.Priority))
	sendErr = producer.SendMessage(msgJson)
	if sendErr != nil {
		log.ErrorContextf(context.Background(), "发送消息到MQ失败: %s", sendErr.Error())
		return "", sendErr
	}

	log.Infof("消息 %s 已发送到%s优先级队列", msgID, data.GetPriorityStr(data.PriorityEnum(msgReq.Priority)))

	// 返回消息ID，表示发送成功
	return msgID, nil
}
