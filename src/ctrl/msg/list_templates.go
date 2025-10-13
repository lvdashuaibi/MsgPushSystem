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

// ListTemplatesHandler 模板列表处理器
type ListTemplatesHandler struct {
	Req    ctrlmodel.ListTemplatesReq
	Resp   ctrlmodel.ListTemplatesResp
	UserId string
}

// ListTemplates 模板列表API
func ListTemplates(c *gin.Context) {
	var hd ListTemplatesHandler
	defer func() {
		hd.Resp.Msg = constant.GetErrMsg(hd.Resp.Code)
		c.JSON(http.StatusOK, hd.Resp)
	}()

	hd.UserId = c.Request.Header.Get(constant.HEADER_USERID)

	if err := c.ShouldBind(&hd.Req); err != nil {
		log.Errorf("ListTemplates shouldBind err %s", err.Error())
		hd.Resp.Code = constant.ERR_SHOULD_BIND
		return
	}

	if err := handler.Run(&hd); err != nil {
		log.Errorf("ListTemplates handler.Run err %s", err.Error())
		if hd.Resp.Code == 0 {
			hd.Resp.Code = constant.ERR_INTERNAL
		}
	}
}

func (h *ListTemplatesHandler) HandleInput() error {
	// 设置默认值
	if h.Req.Page <= 0 {
		h.Req.Page = 1
	}
	if h.Req.PageSize <= 0 {
		h.Req.PageSize = 10
	}
	return nil
}

func (h *ListTemplatesHandler) HandleProcess() error {
	dt := data.GetData()

	offset := (h.Req.Page - 1) * h.Req.PageSize
	templates, total, err := data.MsgTemplateNsp.List(dt.GetDB(), offset, h.Req.PageSize, h.Req.SourceID, h.Req.Channel)
	if err != nil {
		log.Errorf("查询模板列表失败: %s", err.Error())
		h.Resp.Code = constant.ERR_QUERY
		return err
	}

	h.Resp.Templates = templates
	h.Resp.Total = total
	h.Resp.Page = h.Req.Page
	return nil
}
