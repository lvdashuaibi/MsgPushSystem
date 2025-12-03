package consumer

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lvdashuaibi/MsgPushSystem/src/ctrl/msgpush"
	"github.com/lvdashuaibi/MsgPushSystem/src/data"
	"github.com/lvdashuaibi/MsgPushSystem/src/pkg/log"
)

type MsgIntf interface {
	SendMsg() error
	Base() *MsgBase
}

type MsgHandler struct {
	Channel int
	NewProc func() MsgIntf
}

type MsgBase struct {
	To           string            `json:"to" form:"to"`
	Subject      string            `json:"subject" form:"subject"`
	Content      string            `json:"content" form:"content"`
	Priority     int               `json:"priority" form:"priority"`
	TemplateID   string            `json:"templateID" form:"templateID"`
	TemplateData map[string]string `json:"templateData" form:"templateData"`
	NotifyURL    string            `json:"notifyUrl" form:"notifyUrl"`
}

// Base func get base struct
func (p *MsgBase) Base() *MsgBase {
	return p
}

func InitMsgProc() {
	emailMsgProc := MsgHandler{
		Channel: int(data.Channel_EMAIL),
		NewProc: func() MsgIntf { return new(EmailMsgProc) },
	}
	RegisterHandler(&emailMsgProc)
	smsMsgProc := MsgHandler{
		Channel: int(data.Channel_SMS),
		NewProc: func() MsgIntf { return new(SMSMsgProc) },
	}
	RegisterHandler(&smsMsgProc)
	larkProc := MsgHandler{
		Channel: int(data.Channel_LARK),
		NewProc: func() MsgIntf { return new(LarkProc) },
	}
	RegisterHandler(&larkProc)
}

var msgProcMap = make(map[int]*MsgHandler, 0)

// RegisterHandler func RegisterHandler
func RegisterHandler(handler *MsgHandler) {
	msgProcMap[handler.Channel] = handler
}

type EmailMsgProc struct {
	MsgBase
}

func (p *EmailMsgProc) SendMsg() error {
	// 发送对应消息
	log.Infof("📧 EmailMsgProc开始发送邮件，To: %s, Subject: %s, Content: %s", p.To, p.Subject, p.Content)
	err := msgpush.SendEmail(p.To, p.Subject, p.Content)
	if err != nil {
		log.Errorf("❌ EmailMsgProc发送邮件失败: %s", err.Error())
		return err
	}
	log.Infof("✅ EmailMsgProc发送邮件成功，To: %s", p.To)
	return nil
}

type SMSMsgProc struct {
	MsgBase
}

func (p *SMSMsgProc) SendMsg() error {
	// 发送对应消息
	dt := data.GetData()
	mt, err := data.MsgTemplateNsp.Find(dt.GetDB(), p.TemplateID)
	if err != nil {
		return err
	}
	templateParam, _ := json.Marshal(p.TemplateData)
	err = msgpush.SendSMS(p.To, mt.SignName, mt.RelTemplateID, string(templateParam))
	if err != nil {
		return err
	}
	return nil
}

type LarkProc struct {
	MsgBase
}

func (p *LarkProc) SendMsg() error {
	// 发送对应消息
	accessToken, err := msgpush.GetAccessToken()
	if err != nil {
		fmt.Println("Error getting access token:", err)
		return err
	}

	// 检查内容是否为JSON格式的卡片（AI润色生成的）
	// 如果内容以 { 开头并包含 "config" 或 "header"，则认为是卡片JSON
	content := p.Content
	if len(content) > 0 && content[0] == '{' &&
		(strings.Contains(content, `"config"`) || strings.Contains(content, `"header"`)) {
		// 使用卡片消息发送
		log.Infof("🎨 检测到飞书卡片格式，使用卡片消息发送")
		err = msgpush.SendCardMessageFromJSON(accessToken, p.To, content)
	} else {
		// 使用普通文本消息发送
		err = msgpush.SendMessage(accessToken, p.To, content)
	}

	if err != nil {
		return err
	}
	return nil
}
