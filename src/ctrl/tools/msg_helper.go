package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/BitofferHub/msgcenter/src/config"
	"github.com/BitofferHub/msgcenter/src/ctrl/ctrlmodel"
	"github.com/BitofferHub/msgcenter/src/data"
	"github.com/BitofferHub/pkg/middlewares/log"
	"gorm.io/gorm"
)

// CreateMsgRecord 创建消息记录的通用函数
// 参数:
//   - db: 数据库连接
//   - msgID: 消息ID
//   - req: 消息请求结构体或其包含的字段内容
//   - mt: 消息模板信息
//   - status: 消息状态
//
// 返回:
//   - error: 如果有错误发生则返回，否则返回nil
func CreateMsgRecord(db *gorm.DB, msgID string, req *ctrlmodel.SendMsgReq, mt *data.MsgTemplate, status int) error {
	ctx := context.Background()

	// 创建消息记录，使用户可以立即查询到消息状态
	var msgRecord = new(data.MsgRecord)
	msgRecord.Subject = req.Subject
	msgRecord.MsgId = msgID
	msgRecord.TemplateID = req.TemplateID
	msgRecord.To = req.To
	msgRecord.Status = status

	// 将模板数据转换为JSON格式
	if req.TemplateData != nil {
		td, err := json.Marshal(req.TemplateData)
		if err != nil {
			log.ErrorContextf(ctx, "marshal template data err %s", err.Error())
			return err
		}
		msgRecord.TemplateData = string(td)
	}

	// 设置渠道和来源ID
	if mt != nil {
		msgRecord.Channel = mt.Channel
		msgRecord.SourceID = mt.SourceID
	}

	// 将消息记录保存到数据库
	err := data.MsgRecordNsp.Create(db, msgRecord)
	if err != nil {
		log.ErrorContextf(ctx, "创建消息记录失败：%s", err.Error())
		return err
	}

	// 保存到缓存
	if config.Conf.Common.OpenCache {
		jsonData, _ := json.Marshal(msgRecord)
		cacheKey := fmt.Sprintf("%s%s", data.REDIS_KEY_MES_RECORD, msgID)
		data.GetData().GetCache().Set(ctx, cacheKey, string(jsonData), 10000*time.Second)
	}

	log.InfoContextf(ctx, "消息记录 %s 已创建，状态为：%d", msgID, status)
	return nil
}

// CreateOrUpdateMsgRecord 创建或更新消息记录
// 如果消息记录已存在，则更新状态；否则创建新记录
func CreateOrUpdateMsgRecord(db *gorm.DB, msgID string, req *ctrlmodel.SendMsgReq, mt *data.MsgTemplate, status int) error {
	ctx := context.Background()

	// 尝试查找记录
	record, err := data.MsgRecordNsp.Find(db, msgID)
	if err != nil {
		// 只有在记录真的不存在时才创建新记录
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.InfoContextf(ctx, "消息记录 %s 不存在，将创建新记录", msgID)
			return CreateMsgRecord(db, msgID, req, mt, status)
		}
		// 其他查询错误，返回错误
		log.ErrorContextf(ctx, "查询消息记录时发生错误: %s", err.Error())
		return err
	}

	// 记录存在，才更新状态
	log.InfoContextf(ctx, "消息记录 %s 已存在，当前状态:%d，将更新为:%d",
		msgID, record.Status, status)

	err = data.MsgRecordNsp.UpdateStatus(db, msgID, status)
	if err != nil {
		log.ErrorContextf(ctx, "更新消息记录状态失败: %s", err.Error())
		return err
	}

	log.InfoContextf(ctx, "消息记录 %s 状态已更新为：%d", msgID, status)
	return nil
}
