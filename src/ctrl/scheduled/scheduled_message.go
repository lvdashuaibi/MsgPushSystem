package scheduled

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/BitofferHub/msgcenter/src/constant"
	"github.com/BitofferHub/msgcenter/src/ctrl/ctrlmodel"
	"github.com/BitofferHub/msgcenter/src/ctrl/handler"
	"github.com/BitofferHub/msgcenter/src/data"
	"github.com/BitofferHub/pkg/middlewares/log"
	"github.com/BitofferHub/pkg/utils"
	"github.com/gin-gonic/gin"
)

// CreateScheduledMessageHandler 创建定时消息处理器
type CreateScheduledMessageHandler struct {
	Req  ctrlmodel.CreateScheduledMessageReq
	Resp ctrlmodel.CreateScheduledMessageResp
}

// CreateScheduledMessage 创建定时消息API
func CreateScheduledMessage(c *gin.Context) {
	var hd CreateScheduledMessageHandler
	defer func() {
		hd.Resp.Msg = constant.GetErrMsg(hd.Resp.Code)
		c.JSON(http.StatusOK, hd.Resp)
	}()

	if err := c.ShouldBind(&hd.Req); err != nil {
		log.Errorf("CreateScheduledMessage shouldBind err %s", err.Error())
		hd.Resp.Code = constant.ERR_SHOULD_BIND
		return
	}

	if err := handler.Run(&hd); err != nil {
		log.Errorf("CreateScheduledMessage handler.Run err %s", err.Error())
		if hd.Resp.Code == 0 {
			hd.Resp.Code = constant.ERR_INTERNAL
		}
	}
}

func (h *CreateScheduledMessageHandler) HandleInput() error {
	// 验证至少有一种接收者类型
	if h.Req.To == "" && len(h.Req.UserIDs) == 0 && len(h.Req.Tags) == 0 {
		h.Resp.Code = constant.ERR_INPUT_INVALID
		return nil
	}
	return nil
}

func (h *CreateScheduledMessageHandler) HandleProcess() error {
	dt := data.GetData()

	// 解析定时发送时间（使用本地时区）
	loc, _ := time.LoadLocation("Asia/Shanghai") // 使用中国时区
	scheduledTime, err := time.ParseInLocation("2006-01-02 15:04:05", h.Req.ScheduledTime, loc)
	if err != nil {
		log.Errorf("解析定时发送时间失败: %s", err.Error())
		h.Resp.Code = constant.ERR_INPUT_INVALID
		return err
	}

	// 检查时间不能是过去时间（使用相同时区比较）
	now := time.Now().In(loc)
	if scheduledTime.Before(now) {
		log.Warnf("定时发送时间 %s 早于当前时间 %s",
			scheduledTime.Format("2006-01-02 15:04:05"),
			now.Format("2006-01-02 15:04:05"))
		h.Resp.Code = constant.ERR_INPUT_INVALID
		return nil
	}

	// 验证模板是否存在
	ctx := context.Background()
	_, err = dt.GetMsgTemplate(ctx, h.Req.TemplateID)
	if err != nil {
		log.Errorf("获取消息模板失败: %s", err.Error())
		h.Resp.Code = constant.ERR_TEMPLATE_NOT_READY
		return err
	}

	// 解析接收者列表
	recipients, err := h.parseRecipients()
	if err != nil {
		log.Errorf("解析接收者失败: %s", err.Error())
		h.Resp.Code = constant.ERR_INPUT_INVALID
		return err
	}

	if len(recipients) == 0 {
		h.Resp.Code = constant.ERR_INPUT_INVALID
		return errors.New("no valid recipients found")
	}

	// 生成定时消息ID
	scheduleID := utils.NewUuid()

	// 序列化模板数据
	templateDataJSON, _ := json.Marshal(h.Req.TemplateData)

	// 创建定时消息对象
	scheduledMessage := &data.ScheduledMessage{
		ScheduleID:    scheduleID,
		UserIDs:       data.StringSlice(h.Req.UserIDs),
		Tags:          data.StringSlice(h.Req.Tags),
		TemplateID:    h.Req.TemplateID,
		TemplateData:  string(templateDataJSON),
		ScheduledTime: scheduledTime,
		Status:        int(data.ScheduledMessageStatusPending),
	}

	// 如果有直接指定的接收者，将其添加到UserIDs中
	if h.Req.To != "" {
		// 将直接指定的接收者作为特殊用户ID处理
		scheduledMessage.UserIDs = append(scheduledMessage.UserIDs, "direct:"+h.Req.To)
	}

	// 保存到数据库
	if err := data.ScheduledMessageNamespace.Create(dt.GetDB(), scheduledMessage); err != nil {
		log.Errorf("创建定时消息失败: %s", err.Error())
		h.Resp.Code = constant.ERR_INSERT
		return err
	}

	// 添加到Redis定时队列
	timeScore := float64(scheduledTime.Unix())
	member := scheduleID
	_, err = dt.GetCache().ZAdd(ctx, "Scheduled_Messages", timeScore, member)
	if err != nil {
		log.Errorf("添加到Redis定时队列失败: %s", err.Error())
		// 这里不返回错误，因为数据库已经保存成功，可以通过定时任务补偿
	}

	h.Resp.ScheduleID = scheduleID
	log.Infof("定时消息创建成功: %s, 计划发送时间: %s, 接收者数量: %d",
		scheduleID, scheduledTime.Format("2006-01-02 15:04:05"), len(recipients))
	return nil
}

// parseRecipients 解析接收者列表
func (h *CreateScheduledMessageHandler) parseRecipients() ([]string, error) {
	var recipients []string
	dt := data.GetData()

	// 1. 直接指定的接收者
	if h.Req.To != "" {
		recipients = append(recipients, h.Req.To)
	}

	// 2. 按用户ID指定的接收者
	if len(h.Req.UserIDs) > 0 {
		for _, userID := range h.Req.UserIDs {
			user, err := data.UserNamespace.FindByUserID(dt.GetDB(), userID)
			if err != nil {
				log.Warnf("user %s not found: %s", userID, err.Error())
				continue
			}

			// 根据模板渠道选择合适的联系方式
			recipient := h.getRecipientByChannel(user)
			if recipient != "" {
				recipients = append(recipients, recipient)
			}
		}
	}

	// 3. 按标签指定的接收者
	if len(h.Req.Tags) > 0 {
		users, err := data.UserNamespace.FindByAnyTags(dt.GetDB(), h.Req.Tags)
		if err != nil {
			return nil, fmt.Errorf("find users by tags failed: %w", err)
		}

		for _, user := range users {
			recipient := h.getRecipientByChannel(user)
			if recipient != "" {
				recipients = append(recipients, recipient)
			}
		}
	}

	// 去重
	recipients = h.deduplicateRecipients(recipients)
	return recipients, nil
}

// getRecipientByChannel 根据模板渠道获取用户的联系方式
func (h *CreateScheduledMessageHandler) getRecipientByChannel(user *data.User) string {
	ctx := context.Background()
	dt := data.GetData()

	// 获取模板信息以确定渠道
	mt, err := dt.GetMsgTemplate(ctx, h.Req.TemplateID)
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
func (h *CreateScheduledMessageHandler) deduplicateRecipients(recipients []string) []string {
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

// GetScheduledMessageHandler 获取定时消息处理器
type GetScheduledMessageHandler struct {
	Req  ctrlmodel.GetScheduledMessageReq
	Resp ctrlmodel.GetScheduledMessageResp
}

// GetScheduledMessage 获取定时消息API
func GetScheduledMessage(c *gin.Context) {
	var hd GetScheduledMessageHandler
	defer func() {
		hd.Resp.Msg = constant.GetErrMsg(hd.Resp.Code)
		c.JSON(http.StatusOK, hd.Resp)
	}()

	if err := c.ShouldBind(&hd.Req); err != nil {
		log.Errorf("GetScheduledMessage shouldBind err %s", err.Error())
		hd.Resp.Code = constant.ERR_SHOULD_BIND
		return
	}

	if err := handler.Run(&hd); err != nil {
		log.Errorf("GetScheduledMessage handler.Run err %s", err.Error())
		if hd.Resp.Code == 0 {
			hd.Resp.Code = constant.ERR_INTERNAL
		}
	}
}

func (h *GetScheduledMessageHandler) HandleInput() error {
	return nil
}

func (h *GetScheduledMessageHandler) HandleProcess() error {
	dt := data.GetData()

	message, err := data.ScheduledMessageNamespace.FindByScheduleID(dt.GetDB(), h.Req.ScheduleID)
	if err != nil {
		log.Errorf("查找定时消息失败: %s", err.Error())
		h.Resp.Code = constant.ERR_QUERY
		return err
	}

	h.Resp.Message = message
	return nil
}

// ListScheduledMessagesHandler 定时消息列表处理器
type ListScheduledMessagesHandler struct {
	Req  ctrlmodel.ListScheduledMessagesReq
	Resp ctrlmodel.ListScheduledMessagesResp
}

// ListScheduledMessages 定时消息列表API
func ListScheduledMessages(c *gin.Context) {
	var hd ListScheduledMessagesHandler
	defer func() {
		hd.Resp.Msg = constant.GetErrMsg(hd.Resp.Code)
		c.JSON(http.StatusOK, hd.Resp)
	}()

	if err := c.ShouldBind(&hd.Req); err != nil {
		log.Errorf("ListScheduledMessages shouldBind err %s", err.Error())
		hd.Resp.Code = constant.ERR_SHOULD_BIND
		return
	}

	if err := handler.Run(&hd); err != nil {
		log.Errorf("ListScheduledMessages handler.Run err %s", err.Error())
		if hd.Resp.Code == 0 {
			hd.Resp.Code = constant.ERR_INTERNAL
		}
	}
}

func (h *ListScheduledMessagesHandler) HandleInput() error {
	if h.Req.Page <= 0 {
		h.Req.Page = 1
	}
	if h.Req.PageSize <= 0 {
		h.Req.PageSize = 10
	}
	return nil
}

func (h *ListScheduledMessagesHandler) HandleProcess() error {
	dt := data.GetData()

	offset := (h.Req.Page - 1) * h.Req.PageSize
	var status *data.ScheduledMessageStatus
	if h.Req.Status > 0 {
		s := data.ScheduledMessageStatus(h.Req.Status)
		status = &s
	}

	messages, total, err := data.ScheduledMessageNamespace.List(dt.GetDB(), offset, h.Req.PageSize, status)
	if err != nil {
		log.Errorf("查询定时消息列表失败: %s", err.Error())
		h.Resp.Code = constant.ERR_QUERY
		return err
	}

	h.Resp.Messages = messages
	h.Resp.Total = total
	h.Resp.Page = h.Req.Page
	return nil
}

// CancelScheduledMessageHandler 取消定时消息处理器
type CancelScheduledMessageHandler struct {
	Req  ctrlmodel.CancelScheduledMessageReq
	Resp ctrlmodel.CancelScheduledMessageResp
}

// CancelScheduledMessage 取消定时消息API
func CancelScheduledMessage(c *gin.Context) {
	var hd CancelScheduledMessageHandler
	defer func() {
		hd.Resp.Msg = constant.GetErrMsg(hd.Resp.Code)
		c.JSON(http.StatusOK, hd.Resp)
	}()

	if err := c.ShouldBind(&hd.Req); err != nil {
		log.Errorf("CancelScheduledMessage shouldBind err %s", err.Error())
		hd.Resp.Code = constant.ERR_SHOULD_BIND
		return
	}

	if err := handler.Run(&hd); err != nil {
		log.Errorf("CancelScheduledMessage handler.Run err %s", err.Error())
		if hd.Resp.Code == 0 {
			hd.Resp.Code = constant.ERR_INTERNAL
		}
	}
}

func (h *CancelScheduledMessageHandler) HandleInput() error {
	return nil
}

func (h *CancelScheduledMessageHandler) HandleProcess() error {
	dt := data.GetData()

	// 取消定时消息
	if err := data.ScheduledMessageNamespace.Cancel(dt.GetDB(), h.Req.ScheduleID); err != nil {
		log.Errorf("取消定时消息失败: %s", err.Error())
		h.Resp.Code = constant.ERR_UPDATE
		return err
	}

	// 从Redis定时队列中移除
	ctx := context.Background()
	_, err := dt.GetCache().ZRem(ctx, "Scheduled_Messages", h.Req.ScheduleID)
	if err != nil {
		log.Errorf("从Redis定时队列移除失败: %s", err.Error())
		// 这里不返回错误，因为数据库状态已经更新
	}

	log.Infof("定时消息取消成功: %s", h.Req.ScheduleID)
	return nil
}
