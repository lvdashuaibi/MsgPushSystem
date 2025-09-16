package msg

import (
	"net/http"

	"github.com/BitofferHub/msgcenter/src/constant"
	"github.com/BitofferHub/msgcenter/src/ctrl/ctrlmodel"
	"github.com/BitofferHub/msgcenter/src/ctrl/handler"
	"github.com/BitofferHub/msgcenter/src/data"
	"github.com/BitofferHub/pkg/middlewares/log"
	"github.com/BitofferHub/pkg/utils"
	"github.com/gin-gonic/gin"
)

// CreateTemplateHandler 接口处理handler
type CreateTemplateHandler struct {
	Req    ctrlmodel.CreateTemplateReq
	Resp   ctrlmodel.CreateTemplateResp
	UserId string
}

// CreateTemplate 接口
func CreateTemplate(c *gin.Context) {
	// 定义一个 CreateTemplateHandler 类型的变量 hd
	var hd CreateTemplateHandler
	// 使用 defer 关键字确保在函数返回前执行指定的函数
	defer func() {
		// 设置响应的消息为错误码对应的错误消息
		hd.Resp.Msg = constant.GetErrMsg(hd.Resp.Code)
		// 返回 JSON 格式的响应
		c.JSON(http.StatusOK, hd.Resp)
	}()
	// 从请求头中获取用户 ID
	hd.UserId = c.Request.Header.Get(constant.HEADER_USERID)
	// 解析请求包，将请求体绑定到 hd.Req 结构体上
	if err := c.ShouldBind(&hd.Req); err != nil {
		// 记录错误日志
		log.Errorf("CreateTemplate shouldBind err %s", err.Error())
		// 设置响应的错误码
		hd.Resp.Code = constant.ERR_SHOULD_BIND
		// 返回，结束函数执行
		return
	}
	// 执行处理函数，这里会调用对应的 HandleInput 和 HandleProcess
	handler.Run(&hd)
}

// HandleInput 参数检查
func (p *CreateTemplateHandler) HandleInput() error {
	// 检查模板名称是否为空
	if p.Req.Name == "" {
		p.Resp.Code = constant.ERR_INPUT_INVALID
		return constant.ERR_HANDLE_INPUT
	}
	// 检查模板内容是否为空
	if p.Req.Content == "" {
		p.Resp.Code = constant.ERR_INPUT_INVALID
		return constant.ERR_HANDLE_INPUT
	}
	// 检查模板主题是否为空 - 修改为可选字段
	// if p.Req.Subject == "" {
	// 	p.Resp.Code = constant.ERR_INPUT_INVALID
	// 	return constant.ERR_HANDLE_INPUT
	// }
	// 检查模板渠道是否为0
	if p.Req.Channel == 0 {
		p.Resp.Code = constant.ERR_INPUT_INVALID
		return constant.ERR_HANDLE_INPUT
	}
	// 检查模板来源ID是否为空
	if p.Req.SourceID == "" {
		p.Resp.Code = constant.ERR_INPUT_INVALID
		return constant.ERR_HANDLE_INPUT
	}
	return nil
}

// HandleProcess 处理函数
func (p *CreateTemplateHandler) HandleProcess() error {
	// 打印日志，表示进入 HandleProcess 方法
	log.Infof("into HandleProcess")
	// 从 data 包中获取数据实例
	dt := data.GetData()

	// 创建一个新的 MsgTemplate 实例
	var mt = new(data.MsgTemplate)
	// 生成一个新的 UUID 作为模板 ID
	templateID := utils.NewUuid()
	// 设置模板 ID
	mt.TemplateID = templateID
	// 设置模板名称
	mt.Name = p.Req.Name
	// 设置模板内容
	mt.Content = p.Req.Content
	// 设置模板主题
	mt.Subject = p.Req.Subject
	// 设置模板渠道
	mt.Channel = p.Req.Channel
	// 设置模板来源 ID
	mt.SourceID = p.Req.SourceID
	// 设置签名，主要用于短信
	mt.SignName = p.Req.SignName
	// 设置状态为正常状态，可以立即使用
	mt.Status = int(data.TEMPLATE_STATUS_NORMAL)
	// 将新模板保存到数据库中
	err := data.MsgTemplateNsp.Create(dt.GetDB(), mt)
	// 如果发生错误，返回错误
	if err != nil {
		p.Resp.Code = constant.ERR_INTERNAL
		return err
	}
	// 将新创建的模板 ID 设置到响应中
	p.Resp.TemplateID = templateID
	// 返回 nil，表示处理成功
	return nil
}
