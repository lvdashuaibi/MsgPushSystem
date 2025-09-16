package data

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

// UserTags 用户标签类型，用于JSON序列化
type UserTags []string

// Value 实现 driver.Valuer 接口
func (ut UserTags) Value() (driver.Value, error) {
	if len(ut) == 0 {
		return "[]", nil
	}
	return json.Marshal(ut)
}

// Scan 实现 sql.Scanner 接口
func (ut *UserTags) Scan(value interface{}) error {
	if value == nil {
		*ut = UserTags{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("cannot scan into UserTags")
	}

	return json.Unmarshal(bytes, ut)
}

// User 用户模型
type User struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     string    `gorm:"column:user_id;uniqueIndex;size:64;not null" json:"user_id"`
	Name       string    `gorm:"column:name;size:100;not null" json:"name"`
	Nickname   string    `gorm:"column:nickname;size:100" json:"nickname"`
	Mobile     string    `gorm:"column:mobile;size:20;index" json:"mobile"`
	Email      string    `gorm:"column:email;size:100;index" json:"email"`
	LarkID     string    `gorm:"column:lark_id;size:100" json:"lark_id"`
	Tags       UserTags  `gorm:"column:tags;type:json" json:"tags"`
	Status     int       `gorm:"column:status;default:1;index" json:"status"`
	CreateTime time.Time `gorm:"column:create_time;autoCreateTime" json:"create_time"`
	ModifyTime time.Time `gorm:"column:modify_time;autoUpdateTime" json:"modify_time"`
}

// TableName 指定表名
func (User) TableName() string {
	return "t_user"
}

// UserStatus 用户状态枚举
type UserStatus int

const (
	UserStatusDisabled UserStatus = 0 // 禁用
	UserStatusEnabled  UserStatus = 1 // 启用
)

// UserNsp 用户命名空间，包含用户相关的数据库操作
type UserNsp struct{}

var UserNamespace = &UserNsp{}

// Create 创建用户
func (u *UserNsp) Create(db *gorm.DB, user *User) error {
	return db.Create(user).Error
}

// FindByUserID 根据用户ID查找用户
func (u *UserNsp) FindByUserID(db *gorm.DB, userID string) (*User, error) {
	var user User
	err := db.Where("user_id = ? AND status = ?", userID, UserStatusEnabled).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByID 根据主键ID查找用户
func (u *UserNsp) FindByID(db *gorm.DB, id int64) (*User, error) {
	var user User
	err := db.Where("id = ? AND status = ?", id, UserStatusEnabled).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update 更新用户信息
func (u *UserNsp) Update(db *gorm.DB, user *User) error {
	return db.Save(user).Error
}

// Delete 软删除用户（设置状态为禁用）
func (u *UserNsp) Delete(db *gorm.DB, userID string) error {
	return db.Model(&User{}).Where("user_id = ?", userID).Update("status", UserStatusDisabled).Error
}

// List 分页查询用户列表
func (u *UserNsp) List(db *gorm.DB, offset, limit int) ([]*User, int64, error) {
	var users []*User
	var total int64

	// 查询总数
	if err := db.Model(&User{}).Where("status = ?", UserStatusEnabled).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询列表
	err := db.Where("status = ?", UserStatusEnabled).
		Offset(offset).
		Limit(limit).
		Order("create_time DESC").
		Find(&users).Error

	return users, total, err
}

// FindByTags 根据标签查询用户
func (u *UserNsp) FindByTags(db *gorm.DB, tags []string) ([]*User, error) {
	var users []*User
	query := db.Where("status = ?", UserStatusEnabled)

	// 使用JSON_CONTAINS查询包含指定标签的用户
	for _, tag := range tags {
		query = query.Where("JSON_CONTAINS(tags, ?)", `"`+tag+`"`)
	}

	err := query.Find(&users).Error
	return users, err
}

// FindByAnyTags 根据标签查询用户（包含任意一个标签即可）
func (u *UserNsp) FindByAnyTags(db *gorm.DB, tags []string) ([]*User, error) {
	var users []*User
	if len(tags) == 0 {
		return users, nil
	}

	query := db.Where("status = ?", UserStatusEnabled)

	// 构建OR条件，包含任意一个标签即可
	var conditions []string
	var args []interface{}
	for _, tag := range tags {
		conditions = append(conditions, "JSON_CONTAINS(tags, ?)")
		args = append(args, `"`+tag+`"`)
	}

	if len(conditions) > 0 {
		query = query.Where("("+strings.Join(conditions, " OR ")+")", args...)
	}

	err := query.Find(&users).Error
	return users, err
}

// AddTag 为用户添加标签
func (u *UserNsp) AddTag(db *gorm.DB, userID string, tag string) error {
	user, err := u.FindByUserID(db, userID)
	if err != nil {
		return err
	}

	// 检查标签是否已存在
	for _, existingTag := range user.Tags {
		if existingTag == tag {
			return nil // 标签已存在，不需要添加
		}
	}

	// 添加新标签
	user.Tags = append(user.Tags, tag)
	return u.Update(db, user)
}

// RemoveTag 为用户移除标签
func (u *UserNsp) RemoveTag(db *gorm.DB, userID string, tag string) error {
	user, err := u.FindByUserID(db, userID)
	if err != nil {
		return err
	}

	// 移除指定标签
	var newTags UserTags
	for _, existingTag := range user.Tags {
		if existingTag != tag {
			newTags = append(newTags, existingTag)
		}
	}

	user.Tags = newTags
	return u.Update(db, user)
}

// GetTagStatistics 获取标签统计信息
func (u *UserNsp) GetTagStatistics(db *gorm.DB) (map[string]int, error) {
	var users []*User
	err := db.Where("status = ?", UserStatusEnabled).Find(&users).Error
	if err != nil {
		return nil, err
	}

	tagCount := make(map[string]int)
	for _, user := range users {
		for _, tag := range user.Tags {
			tagCount[tag]++
		}
	}

	return tagCount, nil
}
