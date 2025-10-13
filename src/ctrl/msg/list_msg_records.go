package msg

import (
	"net/http"

	"github.com/lvdashuaibi/MsgPushSystem/src/constant"
	"github.com/lvdashuaibi/MsgPushSystem/src/ctrl/ctrlmodel"
	"github.com/lvdashuaibi/MsgPushSystem/src/ctrl/handler"
	"github.com/lvdashuaibi/MsgPushSystem/src/data"
	"github.com/lvdashuaibi/MsgPushSystem/src/pkg/log"
	"github.com/gin-gonic/gin"
)

// ListMsgRecordsHandler 消息记录列表处理handler
type ListMsgRecordsHandler struct {
	Req    ctrlmodel.ListMsgRecordsReq
	Resp   ctrlmodel.ListMsgRecordsResp
	UserId string
}

// ListMsgRecords 消息记录列表接口
func ListMsgRecords(c *gin.Context) {
	var hd ListMsgRecordsHandler
	defer func() {
		hd.Resp.Msg = constant.GetErrMsg(hd.Resp.Code)
		c.JSON(http.StatusOK, hd.Resp)
	}()

	// 获取用户Id
	hd.UserId = c.Request.Header.Get(constant.HEADER_USERID)

	// 解析请求参数
	if err := c.ShouldBindQuery(&hd.Req); err != nil {
		log.Errorf("ListMsgRecords shouldBindQuery err %s", err.Error())
		hd.Resp.Code = constant.ERR_SHOULD_BIND
		return
	}

	// 执行处理函数
	handler.Run(&hd)
}

// HandleInput 参数检查
func (p *ListMsgRecordsHandler) HandleInput() error {
	// 设置默认值
	if p.Req.Page <= 0 {
		p.Req.Page = 1
	}
	if p.Req.PageSize <= 0 {
		p.Req.PageSize = 10
	}
	if p.Req.PageSize > 100 {
		p.Req.PageSize = 100
	}
	return nil
}

// HandleProcess 处理函数
func (p *ListMsgRecordsHandler) HandleProcess() error {
	log.Infof("into ListMsgRecords HandleProcess")
	dt := data.GetData()

	// 计算偏移量
	offset := (p.Req.Page - 1) * p.Req.PageSize

	// 查询消息记录列表
	records, total, err := data.MsgRecordNsp.List(
		dt.GetDB(),
		offset,
		p.Req.PageSize,
		p.Req.MsgID,
		p.Req.To,
		p.Req.Status,
		p.Req.StartTime,
		p.Req.EndTime,
	)

	if err != nil {
		log.Errorf("查询消息记录列表失败: %s", err.Error())
		p.Resp.Code = constant.ERR_INTERNAL
		return err
	}

	// 设置响应数据
	p.Resp.Records = records
	p.Resp.Total = total
	p.Resp.Page = p.Req.Page

	log.Infof("查询消息记录列表成功，总数: %d, 当前页: %d", total, p.Req.Page)
	return nil
}
