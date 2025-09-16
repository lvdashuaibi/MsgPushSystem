package data

import (
	"time"

	"gorm.io/gorm"
)

var MsgTmpQueueTimerNsp MsgTmpQueueTimer

type MsgTmpQueueTimer struct {
	ID            int64
	MsgId         string
	Req           string
	SendTimestamp int64
	Status        int
	CreateTime    *time.Time `gorm:"column:create_time;default:null"`
	ModifyTime    *time.Time `gorm:"column:modify_time;default:null"`
}

// TableName 表名
func (p *MsgTmpQueueTimer) TableName() string {
	return "t_msg_tmp_queue_timer"
}

// Find 查找记录
func (p *MsgTmpQueueTimer) Find(db *gorm.DB, msgID string) (*MsgTmpQueueTimer, error) {
	var data = &MsgTmpQueueTimer{}
	err := db.Where("msg_id= ?", msgID).First(data).Error
	return data, err
}

// Create 创建记录
func (p *MsgTmpQueueTimer) Create(db *gorm.DB, dt *MsgTmpQueueTimer) error {
	err := db.Table(p.TableName()).Create(dt).Error
	return err
}

//

// GetTaskList 获取记录列表
// 1 待执行，2执行中 3.已完成
func (p *MsgTmpQueueTimer) GetOnTimeMsgList(db *gorm.DB, status int, now int64) ([]*MsgTmpQueueTimer, error) {
	var msgList = make([]*MsgTmpQueueTimer, 0)
	err := db.
		Table(p.TableName()).
		Where("send_timestamp <= ?", now).
		Where("status = ?", status).
		Find(&msgList).Error
	if err != nil {
		return nil, err
	}
	return msgList, nil
}

// BatchSetStatus batch set
func (p *MsgTmpQueueTimer) BatchSetStatus(db *gorm.DB, msgIdList []string, status int) error {
	var dic = map[string]interface{}{
		"status": status,
	}
	db = db.Table(p.TableName()).Where("msg_id in (?)", msgIdList).
		UpdateColumns(dic)
	err := db.Error
	if err != nil {
		return err
	}
	return nil
}

func (p *MsgTmpQueueTimer) SetStatus(db *gorm.DB, msgID string, status int) error {
	var dic = map[string]interface{}{
		"status": status,
	}
	db = db.Table(p.TableName()).Where("msg_id = ?", msgID).
		UpdateColumns(dic)
	err := db.Error
	if err != nil {
		return err
	}
	return nil
}
