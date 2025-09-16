package constant

import (
	"errors"
	"fmt"
)

var (
	ERR_HANDLE_INPUT = errors.New("handle input error")
)

var (
	SUCCESS           = 0
	ERR_INPUT_INVALID = 8020
	ERR_SHOULD_BIND   = 8021
	ERR_JSON_MARSHAL  = 8022
	ERR_INTERNAL      = 8023

	ERR_TEMPLATE_NOT_READY       = 8035
	ERR_SEND_MSG                 = 8036
	ERR_INSERT                   = 8037
	ERR_REQUEST_LIMIT            = 8038
	ERR_GET_TASK_SET_POS_FROM_DB = 8039
	ERR_INCREASE_CRT_RETRY_NUM   = 8040
	ERR_SET_TASK                 = 8041
	ERR_GET_TASK_POS             = 8042
	ERR_GET_PROCESSING_COUNT     = 8043
	ERR_SET_USER_PRIORITY        = 8045
	ERR_GET_TASK_CFG_FROM_DB     = 8039
	ERR_INSERT_TIMER             = 8047

	// 用户管理相关错误码
	ERR_USER_ALREADY_EXISTS = 9001
	ERR_USER_NOT_FOUND      = 9002
	ERR_UPDATE              = 9003
	ERR_DELETE              = 9004
	ERR_QUERY               = 9005
)

var errMsgDic = map[int]string{
	SUCCESS:                      "ok",
	ERR_INPUT_INVALID:            "input invalid",
	ERR_SHOULD_BIND:              "should bind failed",
	ERR_JSON_MARSHAL:             "json marshal failed",
	ERR_INSERT:                   "数据库插入失败",
	ERR_SEND_MSG:                 "投递消息失败",
	ERR_TEMPLATE_NOT_READY:       "模板尚未准备好，检查模板状态",
	ERR_REQUEST_LIMIT:            "请求频繁，请稍后重试",
	ERR_GET_TASK_SET_POS_FROM_DB: "get msg set pos from db failed",
	ERR_SET_TASK:                 "set msg failed",
	ERR_INCREASE_CRT_RETRY_NUM:   "increase crt retry num failed",
	ERR_GET_TASK_POS:             "get msg pos failed",
	ERR_GET_PROCESSING_COUNT:     "get processing count failed",
	ERR_SET_USER_PRIORITY:        "set user priority failed",
	ERR_GET_TASK_CFG_FROM_DB:     "get msg cfg failed",
	ERR_INSERT_TIMER:             "数据库插入失败",

	// 用户管理错误描述
	ERR_USER_ALREADY_EXISTS: "用户已存在",
	ERR_USER_NOT_FOUND:      "用户不存在",
	ERR_UPDATE:              "更新失败",
	ERR_DELETE:              "删除失败",
	ERR_QUERY:               "查询失败",
}

// GetErrMsg 获取错误描述
func GetErrMsg(code int) string {
	if msg, ok := errMsgDic[code]; ok {
		return msg
	}
	return fmt.Sprintf("unknown error code %d", code)
}
