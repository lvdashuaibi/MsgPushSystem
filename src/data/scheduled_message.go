package data

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// StringSlice 字符串切片类型，用于JSON序列化
type StringSlice []string

// Value 实现 driver.Valuer 接口
func (ss StringSlice) Value() (driver.Value, error) {
	if len(ss) == 0 {
		return "[]", nil
	}
	return json.Marshal(ss)
}

// Scan 实现 sql.Scanner 接口
func (ss *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*ss = StringSlice{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("cannot scan into StringSlice")
	}

	return json.Unmarshal(bytes, ss)
}

// ScheduledMessage 定时消息模型
type ScheduledMessage struct {
	ID             int64       `gorm:"primaryKey;autoIncrement" json:"id"`
	ScheduleID     string      `gorm:"column:schedule_id;uniqueIndex;size:64;not null" json:"schedule_id"`
	UserIDs        StringSlice `gorm:"column:user_ids;type:text" json:"user_ids"`
	Tags           StringSlice `gorm:"column:tags;type:text" json:"tags"`
	TemplateID     string      `gorm:"column:template_id;size:64;not null" json:"template_id"`
	TemplateData   string      `gorm:"column:template_data;type:text" json:"template_data"`
	ScheduledTime  time.Time   `gorm:"column:scheduled_time;not null;index" json:"scheduled_time"`
	Status         int         `gorm:"column:status;default:1;index" json:"status"`
	ActualSendTime *time.Time  `gorm:"column:actual_send_time" json:"actual_send_time"`
	CreateTime     time.Time   `gorm:"column:create_time;autoCreateTime" json:"create_time"`
	ModifyTime     time.Time   `gorm:"column:modify_time;autoUpdateTime" json:"modify_time"`
}

// TableName 指定表名
func (ScheduledMessage) TableName() string {
	return "t_scheduled_message"
}

// ScheduledMessageStatus 定时消息状态枚举
type ScheduledMessageStatus int

const (
	ScheduledMessageStatusPending   ScheduledMessageStatus = 1 // 待发送
	ScheduledMessageStatusSent      ScheduledMessageStatus = 2 // 已发送
	ScheduledMessageStatusCancelled ScheduledMessageStatus = 3 // 已取消
	ScheduledMessageStatusFailed    ScheduledMessageStatus = 4 // 发送失败
)

// ScheduledMessageNsp 定时消息命名空间
type ScheduledMessageNsp struct{}

var ScheduledMessageNamespace = &ScheduledMessageNsp{}

// Create 创建定时消息
func (s *ScheduledMessageNsp) Create(db *gorm.DB, message *ScheduledMessage) error {
	return db.Create(message).Error
}

// FindByScheduleID 根据定时消息ID查找
func (s *ScheduledMessageNsp) FindByScheduleID(db *gorm.DB, scheduleID string) (*ScheduledMessage, error) {
	var message ScheduledMessage
	err := db.Where("schedule_id = ?", scheduleID).First(&message).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

// Update 更新定时消息
func (s *ScheduledMessageNsp) Update(db *gorm.DB, message *ScheduledMessage) error {
	return db.Save(message).Error
}

// UpdateStatus 更新定时消息状态
func (s *ScheduledMessageNsp) UpdateStatus(db *gorm.DB, scheduleID string, status ScheduledMessageStatus) error {
	updates := map[string]interface{}{
		"status": status,
	}

	// 如果是已发送状态，记录实际发送时间
	if status == ScheduledMessageStatusSent {
		now := time.Now()
		updates["actual_send_time"] = &now
	}

	return db.Model(&ScheduledMessage{}).Where("schedule_id = ?", scheduleID).Updates(updates).Error
}

// Cancel 取消定时消息
func (s *ScheduledMessageNsp) Cancel(db *gorm.DB, scheduleID string) error {
	return s.UpdateStatus(db, scheduleID, ScheduledMessageStatusCancelled)
}

// List 分页查询定时消息列表
func (s *ScheduledMessageNsp) List(db *gorm.DB, offset, limit int, status *ScheduledMessageStatus) ([]*ScheduledMessage, int64, error) {
	var messages []*ScheduledMessage
	var total int64

	query := db.Model(&ScheduledMessage{})
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	// 查询总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询列表
	err := query.Offset(offset).
		Limit(limit).
		Order("create_time DESC").
		Find(&messages).Error

	return messages, total, err
}

// GetPendingMessages 获取待发送的定时消息
func (s *ScheduledMessageNsp) GetPendingMessages(db *gorm.DB, beforeTime time.Time) ([]*ScheduledMessage, error) {
	var messages []*ScheduledMessage
	err := db.Where("status = ? AND scheduled_time <= ?",
		ScheduledMessageStatusPending, beforeTime).
		Order("scheduled_time ASC").
		Find(&messages).Error
	return messages, err
}

// Delete 删除定时消息
func (s *ScheduledMessageNsp) Delete(db *gorm.DB, scheduleID string) error {
	return db.Where("schedule_id = ?", scheduleID).Delete(&ScheduledMessage{}).Error
}
