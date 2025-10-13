package msg

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/lvdashuaibi/MsgPushSystem/src/config"
	"github.com/lvdashuaibi/MsgPushSystem/src/constant"
	"github.com/lvdashuaibi/MsgPushSystem/src/ctrl/ctrlmodel"
	"github.com/lvdashuaibi/MsgPushSystem/src/ctrl/handler"
	"github.com/lvdashuaibi/MsgPushSystem/src/data"
	"github.com/lvdashuaibi/MsgPushSystem/src/pkg/log"
	"github.com/gin-gonic/gin"
)

// GetMsgRecordHandler 接口处理handler
type GetMsgRecordHandler struct {
	Req    ctrlmodel.GetMsgRecordReq
	Resp   ctrlmodel.GetMsgRecordResp
	UserId string
}

// GetMsgRecord 接口
func GetMsgRecord(c *gin.Context) {
	var hd GetMsgRecordHandler
	defer func() {
		hd.Resp.Msg = constant.GetErrMsg(hd.Resp.Code)
		c.JSON(http.StatusOK, hd.Resp)
	}()
	// 获取用户Id
	hd.UserId = c.Request.Header.Get(constant.HEADER_USERID)
	// 解析请求包
	if err := c.ShouldBind(&hd.Req); err != nil {
		log.Errorf("GetMsgRecord shouldBind err %s", err.Error())
		hd.Resp.Code = constant.ERR_SHOULD_BIND
		return
	}
	// 执行处理函数, 这里会调用对应的HandleInput和HandleProcess，往下看
	handler.Run(&hd)
}

// HandleInput 参数检查
func (p *GetMsgRecordHandler) HandleInput() error {
	if p.Req.MsgID == "" {
		p.Resp.Code = constant.ERR_INPUT_INVALID
		return nil
	}
	return nil
}

// HandleProcess 处理函数
func (p *GetMsgRecordHandler) HandleProcess() error {
	log.Infof("into HandleProcess")
	dt := data.GetData()
	ctx := context.Background()
	var record = new(data.MsgRecord)
	cacheKey := fmt.Sprintf("%s%s", data.REDIS_KEY_MES_RECORD, p.Req.MsgID)
	cacheRecord, _, _ := dt.GetCache().Get(ctx, cacheKey)
	log.Infof("cacheRecord: %s, req %+v", cacheRecord, p.Req)
	if len(cacheRecord) > 0 && config.Conf.Common.OpenCache {
		// 从缓存中获取模板数据
		json.Unmarshal([]byte(cacheRecord), record)
		log.Infof("record cache hit %+v", record)
	} else {
		log.Infof("record cache miss")
		var err error
		record, err = data.MsgRecordNsp.Find(dt.GetDB(), p.Req.MsgID)
		if err != nil {
			log.ErrorContextf(ctx, "MsgRecordNsp.Find err %s", err.Error())
			return err
		}
		if config.Conf.Common.OpenCache {
			value, _ := json.Marshal(record)
			dt.GetCache().Set(ctx, cacheKey, string(value), 30*time.Second)
		}
	}

	p.Resp.Subject = record.Subject
	p.Resp.TemplateID = record.TemplateID
	p.Resp.TemplateData = make(map[string]string)
	err := json.Unmarshal([]byte(record.TemplateData), &p.Resp.TemplateData)
	if err != nil {
		log.Errorf("json.Unmarshal err %s", err.Error())
		return err
	}
	return nil
}
