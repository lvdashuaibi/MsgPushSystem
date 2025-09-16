package data

import (
	"gorm.io/gorm"
	"time"
)

var SourceQuotaNsp SourceQuota

type SourceQuota struct {
	ID         int64
	Num        int
	Unit       int
	Channel    int
	SourceID   string
	CreateTime *time.Time `gorm:"column:create_time;default:null"`
	ModifyTime *time.Time `gorm:"column:modify_time;default:null"`
}

// TableName 表名
func (p *SourceQuota) TableName() string {
	return "t_source_quota"
}

// Find 查找记录
func (p *SourceQuota) Find(db *gorm.DB, sourceID string, channel int) (*SourceQuota, error) {
	var data = &SourceQuota{}
	err := db.Where("source_id  = ? and channel = ?", sourceID, channel).First(data).Error
	return data, err
}

// Create 创建记录
func (p *SourceQuota) Create(db *gorm.DB, quota *SourceQuota) error {
	err := db.Create(quota).Error
	return err
}
