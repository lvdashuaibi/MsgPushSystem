package ctrlmodel

import (
	"github.com/lvdashuaibi/MsgPushSystem/src/data"
)

// CreateScheduledMessageReq 创建定时消息请求
type CreateScheduledMessageReq struct {
	To            string            `json:"to"`       // 直接指定接收者（手机号/邮箱等）
	UserIDs       []string          `json:"user_ids"` // 目标用户ID列表
	Tags          []string          `json:"tags"`     // 目标标签列表
	TemplateID    string            `json:"template_id" binding:"required"`
	TemplateData  map[string]string `json:"template_data"`
	ScheduledTime string            `json:"scheduled_time" binding:"required"` // 格式: "2025-09-11 17:30:00"
}

// CreateScheduledMessageResp 创建定时消息响应
type CreateScheduledMessageResp struct {
	RespComm
	ScheduleID string `json:"schedule_id"`
}

// GetScheduledMessageReq 获取定时消息请求
type GetScheduledMessageReq struct {
	ScheduleID string `form:"schedule_id" binding:"required"`
}

// GetScheduledMessageResp 获取定时消息响应
type GetScheduledMessageResp struct {
	RespComm
	Message *data.ScheduledMessage `json:"message"`
}

// ListScheduledMessagesReq 定时消息列表请求
type ListScheduledMessagesReq struct {
	Page     int `form:"page" binding:"min=1"`
	PageSize int `form:"page_size" binding:"min=1,max=100"`
	Status   int `form:"status"` // 可选：按状态过滤
}

// ListScheduledMessagesResp 定时消息列表响应
type ListScheduledMessagesResp struct {
	RespComm
	Messages []*data.ScheduledMessage `json:"messages"`
	Total    int64                    `json:"total"`
	Page     int                      `json:"page"`
}

// CancelScheduledMessageReq 取消定时消息请求
type CancelScheduledMessageReq struct {
	ScheduleID string `json:"schedule_id" binding:"required"`
}

// CancelScheduledMessageResp 取消定时消息响应
type CancelScheduledMessageResp struct {
	RespComm
}

// UpdateScheduledMessageReq 更新定时消息请求
type UpdateScheduledMessageReq struct {
	ScheduleID    string            `json:"schedule_id" binding:"required"`
	UserIDs       []string          `json:"user_ids"`
	Tags          []string          `json:"tags"`
	TemplateID    string            `json:"template_id"`
	TemplateData  map[string]string `json:"template_data"`
	ScheduledTime string            `json:"scheduled_time"` // 格式: "2025-09-11 17:30:00"
}

// UpdateScheduledMessageResp 更新定时消息响应
type UpdateScheduledMessageResp struct {
	RespComm
}

// DeleteScheduledMessageReq 删除定时消息请求
type DeleteScheduledMessageReq struct {
	ScheduleID string `json:"schedule_id" binding:"required"`
}

// DeleteScheduledMessageResp 删除定时消息响应
type DeleteScheduledMessageResp struct {
	RespComm
}
