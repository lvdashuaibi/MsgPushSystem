package initialize

import (
	"github.com/BitofferHub/msgcenter/src/ctrl/msg"
	"github.com/BitofferHub/msgcenter/src/ctrl/scheduled"
	"github.com/BitofferHub/msgcenter/src/ctrl/user"
	"github.com/gin-gonic/gin"
)

// RegisterRouter 注册路由
func RegisterRouter(router *gin.Engine) {
	{
		// 消息相关接口
		router.POST("/msg/send_msg", msg.SendMsg)
		router.GET("/msg/get_msg_record", msg.GetMsgRecord)
		router.GET("/msg/list_msg_records", msg.ListMsgRecords)
		router.POST("/msg/create_template", msg.CreateTemplate)
		router.GET("/msg/get_template", msg.GetTemplate)
		router.GET("/msg/list_templates", msg.ListTemplates)
		router.POST("/msg/update_template", msg.UpdateTemplate)
		router.POST("/msg/del_template", msg.DelTemplate)

		// 用户管理接口
		router.POST("/user/create", user.CreateUser)
		router.GET("/user/get", user.GetUser)
		router.POST("/user/update", user.UpdateUser)
		router.GET("/user/list", user.ListUsers)
		router.POST("/user/delete", user.DeleteUser)
		router.POST("/user/find_by_tags", user.FindUsersByTags)
		router.GET("/user/tag_statistics", user.GetTagStatistics)

		// 定时消息接口
		router.POST("/scheduled/create", scheduled.CreateScheduledMessage)
		router.GET("/scheduled/get", scheduled.GetScheduledMessage)
		router.GET("/scheduled/list", scheduled.ListScheduledMessages)
		router.POST("/scheduled/cancel", scheduled.CancelScheduledMessage)
	}
}
