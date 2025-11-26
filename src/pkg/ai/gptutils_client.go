package ai

import (
	"context"
	"fmt"
	"sync"

	gptutils "github.com/lvdashuaibi/GPTUtils"
	"github.com/sirupsen/logrus"
)

// GPTUtilsClient GPTUtils AI客户端包装器
type GPTUtilsClient struct {
	client interface{} // 使用interface{}来存储*client.HTTPClient
	logger *logrus.Logger
	mu     sync.RWMutex
}

// NewGPTUtilsClient 创建新的GPTUtils客户端
func NewGPTUtilsClient(logger *logrus.Logger) *GPTUtilsClient {
	if logger == nil {
		logger = logrus.New()
	}

	client := gptutils.NewDefaultClient()
	if client == nil {
		logger.Warn("GPTUtils客户端初始化失败，请检查API_KEY环境变量")
	}

	return &GPTUtilsClient{
		client: client,
		logger: logger,
	}
}

// SimpleChat 简单对话
func (g *GPTUtilsClient) SimpleChat(ctx context.Context, message string) (string, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.client == nil {
		return "", fmt.Errorf("GPTUtils客户端未初始化")
	}

	g.logger.Infof("🤖 GPTUtils简单对话: %s", message)

	// 类型转换
	httpClient, ok := g.client.(*gptutils.HTTPClient)
	if !ok {
		return "", fmt.Errorf("客户端类型转换失败")
	}

	response, err := httpClient.SimpleChat(ctx, message)
	if err != nil {
		g.logger.Errorf("❌ GPTUtils对话失败: %v", err)
		return "", err
	}

	g.logger.Infof("✅ GPTUtils对话成功，响应长度: %d", len(response))
	return response, nil
}

// SimpleChatStream 流式对话
func (g *GPTUtilsClient) SimpleChatStream(ctx context.Context, message string, callback func(chunk string) error) error {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.client == nil {
		return fmt.Errorf("GPTUtils客户端未初始化")
	}

	g.logger.Infof("🤖 GPTUtils流式对话: %s", message)

	// 类型转换
	httpClient, ok := g.client.(*gptutils.HTTPClient)
	if !ok {
		return fmt.Errorf("客户端类型转换失败")
	}

	err := httpClient.SimpleChatStream(ctx, message, callback)
	if err != nil {
		g.logger.Errorf("❌ GPTUtils流式对话失败: %v", err)
		return err
	}

	g.logger.Infof("✅ GPTUtils流式对话完成")
	return nil
}

// Chat 完整对话接口
func (g *GPTUtilsClient) Chat(ctx context.Context, req gptutils.ChatRequest) (*gptutils.ChatResponse, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.client == nil {
		return nil, fmt.Errorf("GPTUtils客户端未初始化")
	}

	g.logger.Infof("🤖 GPTUtils完整对话，消息数: %d", len(req.Messages))

	// 类型转换
	httpClient, ok := g.client.(*gptutils.HTTPClient)
	if !ok {
		return nil, fmt.Errorf("客户端类型转换失败")
	}

	resp, err := httpClient.Chat(ctx, req)
	if err != nil {
		g.logger.Errorf("❌ GPTUtils完整对话失败: %v", err)
		return nil, err
	}

	g.logger.Infof("✅ GPTUtils完整对话成功")
	return resp, nil
}

// AnalyzeText 分析文本
func (g *GPTUtilsClient) AnalyzeText(ctx context.Context, text string) (string, error) {
	prompt := fmt.Sprintf(`请分析以下文本内容，提供关键信息总结和建议：

%s

请提供：
1. 内容摘要
2. 关键要点
3. 建议或改进方向`, text)

	return g.SimpleChat(ctx, prompt)
}

// GenerateContent 生成内容
func (g *GPTUtilsClient) GenerateContent(ctx context.Context, topic string, requirements string) (string, error) {
	prompt := fmt.Sprintf(`请根据以下要求生成内容：

主题: %s
要求: %s

请生成高质量的内容。`, topic, requirements)

	return g.SimpleChat(ctx, prompt)
}

// TranslateText 翻译文本
func (g *GPTUtilsClient) TranslateText(ctx context.Context, text string, targetLanguage string) (string, error) {
	prompt := fmt.Sprintf(`请将以下文本翻译成%s：

%s

只返回翻译结果，不需要其他说明。`, targetLanguage, text)

	return g.SimpleChat(ctx, prompt)
}

// IsAvailable 检查客户端是否可用
func (g *GPTUtilsClient) IsAvailable() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.client != nil
}

// Close 关闭客户端
func (g *GPTUtilsClient) Close() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.logger.Info("关闭GPTUtils客户端")
	g.client = nil
	return nil
}
