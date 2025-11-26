package ctrlmodel

// RespComm 通用的响应消息
type RespComm struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// SendMsgReq 请求消息
type SendMsgReq struct {
	To            string            `json:"to" form:"to"`             // 直接指定接收者（手机号/邮箱等）
	UserIDs       []string          `json:"user_ids" form:"user_ids"` // 目标用户ID列表
	Tags          []string          `json:"tags" form:"tags"`         // 目标标签列表
	Subject       string            `json:"subject" form:"subject"`
	Priority      int               `json:"priority" form:"priority"`
	TemplateID    string            `json:"templateID" form:"templateID"`
	TemplateData  map[string]string `json:"templateData" form:"templateData"`
	SendTimestamp int64             `json:"sendTimestamp" form:"sendTimestamp"`
	MsgID         string
	// 直接编写消息模式字段
	Channels []int  `json:"channels" form:"channels"` // 消息渠道列表 (1:邮件, 2:短信, 3:飞书, 4:微信, 5:钉钉) - 支持多选
	Content  string `json:"content" form:"content"`   // 消息内容（直接编写模式）
}

// SendMsgResp 响应消息
type SendMsgResp struct {
	RespComm
	MsgID string `json:"msgID"`
}

// GetMsgResult 请求消息
type GetMsgRecordReq struct {
	MsgID string `json:"msgID" form:"msgID"`
}

// GetMsgResult 响应消息
type GetMsgRecordResp struct {
	RespComm
	To           string            `json:"to" form:"to"`
	Subject      string            `json:"subject" form:"subject"`
	TemplateID   string            `json:"templateID" form:"templateID"`
	TemplateData map[string]string `json:"templateData" form:"templateData"`
}

// ListMsgRecordsReq 消息记录列表请求
type ListMsgRecordsReq struct {
	Page      int    `form:"page" binding:"min=1"`
	PageSize  int    `form:"page_size" binding:"min=1,max=100"`
	MsgID     string `form:"msg_id"`
	To        string `form:"to"`
	Status    int    `form:"status"`
	StartTime string `form:"start_time"`
	EndTime   string `form:"end_time"`
}

// ListMsgRecordsResp 消息记录列表响应
type ListMsgRecordsResp struct {
	RespComm
	Records interface{} `json:"records"`
	Total   int64       `json:"total"`
	Page    int         `json:"page"`
}

type CreateTemplateReq struct {
	SourceID string `json:"sourceID" form:"sourceID"`
	Name     string `json:"name" form:"name"`
	Subject  string `json:"subject" form:"subject"`
	SignName string `json:"signName" form:"signName"`
	Channel  int    `json:"channel" form:"channel"`
	Content  string `json:"content" form:"content"`
}

type CreateTemplateResp struct {
	RespComm
	TemplateID string `json:"templateID"`
}

type GetTemplateReq struct {
	TemplateID string `json:"templateID" form:"templateID"`
}

type GetTemplateResp struct {
	RespComm
	RelTemplateID string `json:"relTemplateID"`
	SourceID      string `json:"sourceID" form:"sourceID"`
	SignName      string `json:"signName" form:"signName"`
	Name          string `json:"name" form:"name"`
	Subject       string `json:"subject" form:"subject"`
	Channel       int    `json:"channel" form:"channel"`
	Content       string `json:"content" form:"content"`
}

type UpdateTemplateReq struct {
	TemplateID string `json:"templateID"`
	Name       string `json:"name" form:"name"`
	SourceID   string `json:"sourceID" form:"sourceID"`
	Subject    string `json:"subject" form:"subject"`
	Channel    int    `json:"channel" form:"channel"`
	Content    string `json:"content" form:"content"`
}

type UpdateTemplateResp struct {
	RespComm
}

type DelTemplateReq struct {
	TemplateID string `json:"templateID"`
}

type DelTemplateResp struct {
	RespComm
}

// ListTemplatesReq 模板列表请求
type ListTemplatesReq struct {
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"page_size" binding:"min=1,max=100"`
	SourceID string `form:"source_id"`
	Channel  int    `form:"channel"`
}

// ListTemplatesResp 模板列表响应
type ListTemplatesResp struct {
	RespComm
	Templates interface{} `json:"templates"`
	Total     int64       `json:"total"`
	Page      int         `json:"page"`
}
