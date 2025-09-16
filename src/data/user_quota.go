package data

import (
	"gorm.io/gorm"
	"time"
)

var UserQuotaNsp UserQuota

type UserQuota struct {
	ID         int64
	Num        int
	Unit       int
	Channel    int
	SourceID   string
	UserID     string
	CreateTime *time.Time `gorm:"column:create_time;default:null"`
	ModifyTime *time.Time `gorm:"column:modify_time;default:null"`
}

// Find 查找记录
func (p *UserQuota) Find(db *gorm.DB, sourceID string, channel int) (*UserQuota, error) {
	var data = &UserQuota{}
	err := db.Where("source_id  = ? and channel = ?", sourceID, channel).First(data).Error
	return data, err
}

// Create 创建记录
func (p *UserQuota) Create(db *gorm.DB, quota *UserQuota) error {
	var data = &UserQuota{}
	data = quota
	err := db.First(data).Error
	return err
}
