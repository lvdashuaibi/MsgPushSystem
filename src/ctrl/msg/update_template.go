package msg

import (
	"github.com/BitofferHub/msgcenter/src/constant"
	"github.com/BitofferHub/msgcenter/src/ctrl/ctrlmodel"
	"github.com/BitofferHub/msgcenter/src/ctrl/handler"
	"github.com/BitofferHub/msgcenter/src/data"
	"github.com/BitofferHub/pkg/middlewares/log"
	"github.com/gin-gonic/gin"
	"net/http"
)

// UpdateTemplateHandler 接口处理handler
type UpdateTemplateHandler struct {
	Req    ctrlmodel.UpdateTemplateReq
	Resp   ctrlmodel.UpdateTemplateResp
	UserId string
}

// UpdateTemplate 接口
func UpdateTemplate(c *gin.Context) {
	var hd UpdateTemplateHandler
	defer func() {
		hd.Resp.Msg = constant.GetErrMsg(hd.Resp.Code)
		c.JSON(http.StatusOK, hd.Resp)
	}()
	// 获取用户Id
	hd.UserId = c.Request.Header.Get(constant.HEADER_USERID)
	// 解析请求包
	if err := c.ShouldBind(&hd.Req); err != nil {
		log.Errorf("UpdateTemplate shouldBind err %s", err.Error())
		hd.Resp.Code = constant.ERR_SHOULD_BIND
		return
	}
	// 执行处理函数, 这里会调用对应的HandleInput和HandleProcess，往下看
	handler.Run(&hd)
}

// HandleInput 参数检查
func (p *UpdateTemplateHandler) HandleInput() error {
	if p.Req.TemplateID == "" {
		p.Resp.Code = constant.ERR_INPUT_INVALID
		return nil
	}
	return nil
}

// HandleProcess 处理函数
func (p *UpdateTemplateHandler) HandleProcess() error {
	log.Infof("into HandleProcess")
	dt := data.GetData()
	mt, err := data.MsgTemplateNsp.Find(dt.GetDB(), p.Req.TemplateID)
	if err != nil {
		return err
	}
	if p.Req.Name != "" {
		mt.Name = p.Req.Name
	}
	if mt.Content != "" {
		mt.Content = p.Req.Content
	}
	mt.Content = p.Req.Content
	if mt.Subject != "" {
		mt.Subject = p.Req.Subject
	}
	if mt.Channel != 0 {
		mt.Channel = p.Req.Channel
	}
	if p.Req.SourceID != "" {
		mt.SourceID = p.Req.SourceID
	}
	err = data.MsgTemplateNsp.Save(dt.GetDB(), mt)
	if err != nil {
		return err
	}
	return nil
}
