package tools

import (
	"strings"

	"github.com/lvdashuaibi/MsgPushSystem/src/pkg/log"
)

func TemplateReplace(templateContent string, data map[string]string) (string, error) {
	log.Infof("templateContent: %s, data: %v", templateContent, data)

	// 使用简单的字符串替换方式，支持 {{key}} 格式
	result := templateContent
	for key, value := range data {
		placeholder := "{{" + key + "}}"
		log.Infof("替换占位符: %s -> %s", placeholder, value)
		// 使用strings.ReplaceAll进行替换
		result = strings.ReplaceAll(result, placeholder, value)
	}

	log.Infof("模板替换结果: %s", result)
	return result, nil
}
