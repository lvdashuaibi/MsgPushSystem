package ai

import (
	"context"

	gptutils "github.com/lvdashuaibi/GPTUtils"
)

// AIClient AI客户端接口
type AIClient interface {
	// SimpleChat 简单对话
	SimpleChat(ctx context.Context, message string) (string, error)

	// SimpleChatStream 流式对话
	SimpleChatStream(ctx context.Context, message string, callback func(chunk string) error) error

	// Chat 完整对话接口
	Chat(ctx context.Context, req gptutils.ChatRequest) (*gptutils.ChatResponse, error)

	// AnalyzeText 分析文本
	AnalyzeText(ctx context.Context, text string) (string, error)

	// GenerateContent 生成内容
	GenerateContent(ctx context.Context, topic string, requirements string) (string, error)

	// TranslateText 翻译文本
	TranslateText(ctx context.Context, text string, targetLanguage string) (string, error)

	// IsAvailable 检查客户端是否可用
	IsAvailable() bool

	// Close 关闭客户端
	Close() error
}
