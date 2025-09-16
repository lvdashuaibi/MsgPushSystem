package data

import (
	"time"

	"gorm.io/gorm"
)

var MsgTemplateNsp MsgTemplate

type MsgTemplate struct {
	ID            int64
	TemplateID    string
	RelTemplateID string
	Name          string
	Content       string
	Subject       string
	Channel       int
	SourceID      string
	SignName      string
	Status        int
	Ext           string
	CreateTime    *time.Time `gorm:"column:create_time;default:null"`
	ModifyTime    *time.Time `gorm:"column:modify_time;default:null"`
}

// TableName 表名
func (p *MsgTemplate) TableName() string {
	return "t_msg_template"
}

func (p *MsgTemplate) Find(db *gorm.DB, templateID string) (*MsgTemplate, error) {
	var data = &MsgTemplate{}
	err := db.Where("template_id= ?", templateID).First(data).Error
	return data, err
}

func (p *MsgTemplate) Create(db *gorm.DB, dt *MsgTemplate) error {
	data := dt
	err := db.Create(data).Error
	return err
}

func (p *MsgTemplate) Save(db *gorm.DB, dt *MsgTemplate) error {
	err := db.Save(dt).Error
	return err
}

func (p *MsgTemplate) Delete(db *gorm.DB, templateID string) error {
	err := db.Where("template_id = ?", templateID).Delete(&MsgTemplate{}).Error
	return err
}

// List 分页查询模板列表
func (p *MsgTemplate) List(db *gorm.DB, offset, limit int, sourceID string, channel int) ([]*MsgTemplate, int64, error) {
	var templates []*MsgTemplate
	var total int64

	query := db.Model(&MsgTemplate{})

	// 添加过滤条件
	if sourceID != "" {
		query = query.Where("source_id = ?", sourceID)
	}
	if channel > 0 {
		query = query.Where("channel = ?", channel)
	}

	// 查询总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询列表
	err := query.Offset(offset).
		Limit(limit).
		Order("create_time DESC").
		Find(&templates).Error

	return templates, total, err
}
