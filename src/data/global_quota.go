package data

import (
	"gorm.io/gorm"
	"time"
)

var GlobalQuotaNsp GlobalQuota

type GlobalQuota struct {
	ID         int64
	Num        int
	Unit       int
	Channel    int
	CreateTime *time.Time `gorm:"column:create_time;default:null"`
	ModifyTime *time.Time `gorm:"column:modify_time;default:null"`
}

// TableName 表名
func (p *GlobalQuota) TableName() string {
	return "t_global_quota"
}

// Find 查找记录
func (p *GlobalQuota) Find(db *gorm.DB, channel int) (*GlobalQuota, error) {
	var data = &GlobalQuota{}
	err := db.Where("channel = ?", channel).First(data).Error
	return data, err
}

// Create 创建记录
func (p *GlobalQuota) Create(db *gorm.DB, quota *GlobalQuota) error {
	err := db.Create(quota).Error
	return err
}
