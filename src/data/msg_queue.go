package data

import (
	"time"

	"gorm.io/gorm"
)

var MsgQueueNsp MsgQueue

type MsgQueue struct {
	ID           int64
	MsgId        string
	To           string
	Subject      string
	Channel      int
	TemplateID   string
	TemplateData string
	Priority     int
	Status       int
	CreateTime   *time.Time `gorm:"column:create_time;default:null"`
	ModifyTime   *time.Time `gorm:"column:modify_time;default:null"`
}

// TableName 表名
func (p *MsgQueue) TableName() string {
	return "t_msg_queue"
}

// Find 查找记录
func (p *MsgQueue) Find(db *gorm.DB, priorityStr string, msgID string) (*MsgQueue, error) {
	var data = &MsgQueue{}
	err := db.Table(p.TableName() + "_" + priorityStr).Where("msg_id= ?", msgID).First(data).Error
	return data, err
}

// Create 创建记录
func (p *MsgQueue) Create(db *gorm.DB, priorityStr string, dt *MsgQueue) error {
	err := db.Table(p.TableName() + "_" + priorityStr).Create(dt).Error
	return err
}

//

// GetTaskList 获取记录列表
// 1 待执行，2执行中 3.已完成
func (p *MsgQueue) GetMsgList(db *gorm.DB,
	priorityStr string, status int, limit int) ([]*MsgQueue, error) {
	var msgList = make([]*MsgQueue, 0)
	err := db.
		Table(p.TableName()+"_"+priorityStr).
		Where("status = ?", status).
		Order("create_time").
		Limit(limit).
		Find(&msgList).Error
	if err != nil {
		return nil, err
	}
	return msgList, nil
}

// BatchSetStatus batch set
func (p *MsgQueue) BatchSetStatus(db *gorm.DB, priorityStr string, msgIdList []string, status int) error {
	var dic = map[string]interface{}{
		"status": status,
	}
	db = db.Table(p.TableName()+"_"+priorityStr).Where("msg_id in (?)", msgIdList).
		UpdateColumns(dic)
	err := db.Error
	if err != nil {
		return err
	}
	return nil
}

func (p *MsgQueue) SetStatus(db *gorm.DB, priorityStr string, msgID string, status int) error {
	var dic = map[string]interface{}{
		"status": status,
	}
	db = db.Table(p.TableName()+"_"+priorityStr).Where("msg_id = ?", msgID).
		UpdateColumns(dic)
	err := db.Error
	if err != nil {
		return err
	}
	return nil
}
