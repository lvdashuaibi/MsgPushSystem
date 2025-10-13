package msgpush

import (
	"crypto/tls"
	"sync"

	"github.com/lvdashuaibi/MsgPushSystem/src/config"
	"github.com/lvdashuaibi/MsgPushSystem/src/pkg/log"
	"gopkg.in/gomail.v2"
)

const (
	// 端口
	port = 465
	// 邮箱服务器
	emailHost = "smtp.qq.com"
)

var (
	once = sync.Once{}
	d    *gomail.Dialer
)

// 发送给谁
func SendEmail(to string, subject string, text string) error {
	once.Do(func() {
		log.Infof("初始化邮件发送器，账号: %s", config.Conf.Common.EmailAccount)
		d = gomail.NewDialer(emailHost, port, config.Conf.Common.EmailAccount, config.Conf.Common.EmailAuthCode)
		d.TLSConfig = &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         emailHost,
		}
	})

	m := gomail.NewMessage()
	// 设置发送者
	m.SetHeader("From", config.Conf.Common.EmailAccount)
	// 设置接收者
	m.SetHeader("To", to)
	// 设置主题
	m.SetHeader("Subject", subject)

	// 检查内容是否包含HTML标签，如果包含则使用HTML格式
	if len(text) > 0 && (text[0] == '<' || text[len(text)-1] == '>') {
		m.SetBody("text/html", text)
		log.Infof("发送HTML格式邮件到: %s", to)
	} else {
		m.SetBody("text/plain", text)
		log.Infof("发送纯文本邮件到: %s", to)
	}

	// Send the email
	log.Infof("开始发送邮件，发送者: %s，接收者: %s，主题: %s", config.Conf.Common.EmailAccount, to, subject)
	if err := d.DialAndSend(m); err != nil {
		log.Errorf("发送邮件失败，发送者: %s，接收者: %s，错误: %s", config.Conf.Common.EmailAccount, to, err.Error())
		return err
	}
	log.Infof("发送邮件成功，发送者: %s，接收者: %s", config.Conf.Common.EmailAccount, to)
	return nil
}
