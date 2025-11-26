package ai

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
)

// MessageAssistant 消息AI助手
type MessageAssistant struct {
	client AIClient
	logger *logrus.Logger
}

// NewMessageAssistant 创建消息AI助手
func NewMessageAssistant(client AIClient, logger *logrus.Logger) *MessageAssistant {
	if logger == nil {
		logger = logrus.New()
	}

	return &MessageAssistant{
		client: client,
		logger: logger,
	}
}

// GenerateEmailContent 生成邮件内容
func (m *MessageAssistant) GenerateEmailContent(ctx context.Context, subject string, requirements string) (string, error) {
	prompt := fmt.Sprintf(`请根据以下要求生成一封专业的邮件内容：

主题: %s
要求: %s

请生成邮件正文，格式清晰，内容专业。`, subject, requirements)

	return m.client.SimpleChat(ctx, prompt)
}

// GenerateSMSContent 生成短信内容
func (m *MessageAssistant) GenerateSMSContent(ctx context.Context, purpose string, maxLength int) (string, error) {
	prompt := fmt.Sprintf(`请根据以下要求生成一条短信内容：

目的: %s
最大长度: %d字符

请生成简洁有效的短信内容，不超过指定长度。`, purpose, maxLength)

	return m.client.SimpleChat(ctx, prompt)
}

// GenerateLarkMessage 生成飞书消息
func (m *MessageAssistant) GenerateLarkMessage(ctx context.Context, title string, content string) (string, error) {
	prompt := fmt.Sprintf(`请根据以下信息生成一条飞书消息：

标题: %s
内容: %s

请生成格式化的飞书消息，包含标题和详细内容。`, title, content)

	return m.client.SimpleChat(ctx, prompt)
}

// OptimizeMessageContent 优化消息内容
func (m *MessageAssistant) OptimizeMessageContent(ctx context.Context, originalContent string, targetAudience string) (string, error) {
	prompt := fmt.Sprintf(`请优化以下消息内容，使其更适合%s：

原始内容:
%s

请提供优化后的内容，保持原意但更加清晰、专业和有吸引力。`, targetAudience, originalContent)

	return m.client.SimpleChat(ctx, prompt)
}

// SummarizeContent 总结内容
func (m *MessageAssistant) SummarizeContent(ctx context.Context, content string, maxLength int) (string, error) {
	prompt := fmt.Sprintf(`请总结以下内容，不超过%d字符：

%s

请提供简洁的总结。`, maxLength, content)

	return m.client.SimpleChat(ctx, prompt)
}

// CheckContentQuality 检查内容质量
func (m *MessageAssistant) CheckContentQuality(ctx context.Context, content string) (map[string]interface{}, error) {
	prompt := fmt.Sprintf(`请检查以下内容的质量，并提供评分和建议：

%s

请从以下方面评估：
1. 清晰度 (1-10分)
2. 专业性 (1-10分)
3. 吸引力 (1-10分)
4. 改进建议

请以结构化的方式返回评估结果。`, content)

	result, err := m.client.SimpleChat(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// 返回结果作为字符串，实际应用中可以解析为结构化数据
	return map[string]interface{}{
		"evaluation": result,
	}, nil
}

// GeneratePersonalizedMessage 生成个性化消息
func (m *MessageAssistant) GeneratePersonalizedMessage(ctx context.Context, recipientName string, recipientInfo string, messageType string) (string, error) {
	prompt := fmt.Sprintf(`请为以下收件人生成一条个性化的%s消息：

收件人: %s
收件人信息: %s

请生成温暖、个性化且专业的消息。`, messageType, recipientName, recipientInfo)

	return m.client.SimpleChat(ctx, prompt)
}

// TranslateMessage 翻译消息
func (m *MessageAssistant) TranslateMessage(ctx context.Context, message string, targetLanguage string) (string, error) {
	return m.client.TranslateText(ctx, message, targetLanguage)
}

// IsAvailable 检查助手是否可用
func (m *MessageAssistant) IsAvailable() bool {
	return m.client != nil && m.client.IsAvailable()
}
