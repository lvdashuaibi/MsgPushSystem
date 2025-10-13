package ctrlmodel

import "github.com/lvdashuaibi/MsgPushSystem/src/data"

// CreateUserReq 创建用户请求
type CreateUserReq struct {
	UserID   string   `json:"user_id" binding:"required"`
	Name     string   `json:"name" binding:"required"`
	Nickname string   `json:"nickname"`
	Mobile   string   `json:"mobile"`
	Email    string   `json:"email"`
	LarkID   string   `json:"lark_id"`
	Tags     []string `json:"tags"`
}

// CreateUserResp 创建用户响应
type CreateUserResp struct {
	RespComm
	UserID string `json:"user_id"`
}

// UpdateUserReq 更新用户请求
type UpdateUserReq struct {
	UserID   string   `json:"user_id" binding:"required"`
	Name     string   `json:"name"`
	Nickname string   `json:"nickname"`
	Mobile   string   `json:"mobile"`
	Email    string   `json:"email"`
	LarkID   string   `json:"lark_id"`
	Tags     []string `json:"tags"`
}

// UpdateUserResp 更新用户响应
type UpdateUserResp struct {
	RespComm
}

// GetUserReq 获取用户请求
type GetUserReq struct {
	UserID string `form:"user_id" binding:"required"`
}

// GetUserResp 获取用户响应
type GetUserResp struct {
	RespComm
	User *data.User `json:"user"`
}

// ListUsersReq 用户列表请求
type ListUsersReq struct {
	Page     int `form:"page" binding:"min=1"`
	PageSize int `form:"page_size" binding:"min=1,max=100"`
}

// ListUsersResp 用户列表响应
type ListUsersResp struct {
	RespComm
	Users []*data.User `json:"users"`
	Total int64        `json:"total"`
	Page  int          `json:"page"`
}

// DeleteUserReq 删除用户请求
type DeleteUserReq struct {
	UserID string `json:"user_id" binding:"required"`
}

// DeleteUserResp 删除用户响应
type DeleteUserResp struct {
	RespComm
}

// AddUserTagReq 添加用户标签请求
type AddUserTagReq struct {
	UserID string `json:"user_id" binding:"required"`
	Tag    string `json:"tag" binding:"required"`
}

// AddUserTagResp 添加用户标签响应
type AddUserTagResp struct {
	RespComm
}

// RemoveUserTagReq 移除用户标签请求
type RemoveUserTagReq struct {
	UserID string `json:"user_id" binding:"required"`
	Tag    string `json:"tag" binding:"required"`
}

// RemoveUserTagResp 移除用户标签响应
type RemoveUserTagResp struct {
	RespComm
}

// FindUsersByTagsReq 根据标签查找用户请求
type FindUsersByTagsReq struct {
	Tags      []string `json:"tags" binding:"required"`
	MatchType string   `json:"match_type"` // "all": 必须包含所有标签, "any": 包含任意标签即可
}

// FindUsersByTagsResp 根据标签查找用户响应
type FindUsersByTagsResp struct {
	RespComm
	Users []*data.User `json:"users"`
	Count int          `json:"count"`
}

// GetTagStatisticsReq 获取标签统计请求
type GetTagStatisticsReq struct {
	// 暂无参数
}

// TagStatistic 标签统计信息
type TagStatistic struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

// GetTagStatisticsResp 获取标签统计响应
type GetTagStatisticsResp struct {
	RespComm
	Data []TagStatistic `json:"data"`
}
