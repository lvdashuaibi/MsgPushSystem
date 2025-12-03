package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
)

// ChannelType 渠道类型
type ChannelType int

const (
	ChannelEmail ChannelType = 1 // 邮件
	ChannelSMS   ChannelType = 2 // 短信
	ChannelLark  ChannelType = 3 // 飞书
)

// PolishedContent 润色后的内容
type PolishedContent struct {
	Channel     ChannelType `json:"channel"`     // 渠道类型
	Subject     string      `json:"subject"`     // 主题/标题
	Content     string      `json:"content"`     // 内容
	Format      string      `json:"format"`      // 格式类型 (html/text/json)
	RawContent  string      `json:"raw_content"` // 原始内容
	Description string      `json:"description"` // 内容描述
}

// MultiChannelContent 多渠道内容
type MultiChannelContent struct {
	OriginalIntent string           `json:"original_intent"` // 原始意图
	EmailContent   *PolishedContent `json:"email_content"`   // 邮件内容
	SMSContent     *PolishedContent `json:"sms_content"`     // 短信内容
	LarkContent    *PolishedContent `json:"lark_content"`    // 飞书内容
}

// ContentPolisher AI内容润色器
type ContentPolisher struct {
	client AIClient
	logger *logrus.Logger
}

// NewContentPolisher 创建内容润色器
func NewContentPolisher(client AIClient, logger *logrus.Logger) *ContentPolisher {
	if logger == nil {
		logger = logrus.New()
	}

	return &ContentPolisher{
		client: client,
		logger: logger,
	}
}

// PolishForEmail 为邮件渠道润色内容
func (p *ContentPolisher) PolishForEmail(ctx context.Context, originalIntent string) (*PolishedContent, error) {
	p.logger.Infof("📧 开始为邮件渠道润色内容")

	prompt := fmt.Sprintf(`请将以下文本内容转换为HTML格式的邮件内容：

原始内容：%s

⚠️ 重要要求：
- **严格保持原文的所有信息和含义，不要添加、删除或修改任何实质内容**
- 只进行格式转换，将文本内容结构化为HTML邮件格式
- 不要改写、润色或优化原文表达

格式转换要求：
1. 生成完整的HTML格式邮件内容
2. 包含适当的标题（使用<h2>标签）
3. 正文分段清晰，使用<p>标签
4. 重要信息使用<strong>标签加粗强调
5. 保持原文的语气和风格
6. 包含适当的问候语和落款
7. 使用合适的HTML样式，使邮件美观易读

请按以下JSON格式返回：
{
  "subject": "邮件主题",
  "content": "完整的HTML邮件内容",
  "description": "内容简要说明"
}

只返回JSON，不要其他说明。`, originalIntent)

	response, err := p.client.SimpleChat(ctx, prompt)
	if err != nil {
		p.logger.Errorf("❌ 邮件内容润色失败: %v", err)
		return nil, fmt.Errorf("邮件内容润色失败: %w", err)
	}

	// 解析JSON响应
	var result struct {
		Subject     string `json:"subject"`
		Content     string `json:"content"`
		Description string `json:"description"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		p.logger.Warnf("⚠️ JSON解析失败，使用原始响应: %v", err)
		// 如果JSON解析失败，返回原始响应
		return &PolishedContent{
			Channel:     ChannelEmail,
			Subject:     "通知",
			Content:     response,
			Format:      "html",
			RawContent:  originalIntent,
			Description: "AI生成的邮件内容",
		}, nil
	}

	p.logger.Infof("✅ 邮件内容润色成功")
	return &PolishedContent{
		Channel:     ChannelEmail,
		Subject:     result.Subject,
		Content:     result.Content,
		Format:      "html",
		RawContent:  originalIntent,
		Description: result.Description,
	}, nil
}

// PolishContent 通用内容润色 - 优化文字表达
func (p *ContentPolisher) PolishContent(ctx context.Context, originalIntent string) (*PolishedContent, error) {
	p.logger.Infof("✨ 开始润色内容")

	prompt := fmt.Sprintf(`请优化以下文本内容，使其表达更专业、更吸引人：

原始内容：%s

要求：
1. 保持原文的核心含义和信息
2. 优化语言表达，使其更专业、更吸引人
3. 改进句子结构，提高可读性
4. 适当添加过渡词，使逻辑更清晰
5. 使用更恰当的词汇和表述方式
6. 保持原文的语气和风格
7. 不添加任何签名或前缀
8. 不改变原文的长度过多

请按以下JSON格式返回：
{
  "subject": "内容主题或标题（如果有的话）",
  "content": "优化后的内容",
  "description": "优化说明"
}

只返回JSON，不要其他说明。`, originalIntent)

	response, err := p.client.SimpleChat(ctx, prompt)
	if err != nil {
		p.logger.Errorf("❌ 内容润色失败: %v", err)
		return nil, fmt.Errorf("内容润色失败: %w", err)
	}

	// 解析JSON响应
	var result struct {
		Subject     string `json:"subject"`
		Content     string `json:"content"`
		Description string `json:"description"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		p.logger.Warnf("⚠️ JSON解析失败，使用原始响应: %v", err)
		return &PolishedContent{
			Channel:     ChannelSMS,
			Subject:     "优化内容",
			Content:     response,
			Format:      "text",
			RawContent:  originalIntent,
			Description: "AI优化的内容",
		}, nil
	}

	p.logger.Infof("✅ 内容润色成功")
	return &PolishedContent{
		Channel:     ChannelSMS,
		Subject:     result.Subject,
		Content:     result.Content,
		Format:      "text",
		RawContent:  originalIntent,
		Description: result.Description,
	}, nil
}

// PolishForSMS 为短信渠道润色内容
func (p *ContentPolisher) PolishForSMS(ctx context.Context, originalIntent string) (*PolishedContent, error) {
	p.logger.Infof("💬 开始为短信渠道润色内容")

	prompt := fmt.Sprintf(`请将以下原始意图转换为简洁的短信内容：

原始意图：%s

要求：
1. 纯文本格式，不使用HTML或Markdown
2. 以【MsgMate】开头作为签名
3. 字数控制在70字以内
4. 保留所有关键信息（时间、地点、事项）
5. 语言简洁明了，易于理解
6. 使用温馨、友好的语气
7. 重要数字和时间使用阿拉伯数字
8. 适当使用标点符号分隔信息

请按以下JSON格式返回：
{
  "subject": "短信主题（简短）",
  "content": "完整的短信内容（包含【MsgMate】签名）",
  "description": "内容简要说明"
}

只返回JSON，不要其他说明。`, originalIntent)

	response, err := p.client.SimpleChat(ctx, prompt)
	if err != nil {
		p.logger.Errorf("❌ 短信内容润色失败: %v", err)
		return nil, fmt.Errorf("短信内容润色失败: %w", err)
	}

	// 解析JSON响应
	var result struct {
		Subject     string `json:"subject"`
		Content     string `json:"content"`
		Description string `json:"description"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		p.logger.Warnf("⚠️ JSON解析失败，使用原始响应: %v", err)
		return &PolishedContent{
			Channel:     ChannelSMS,
			Subject:     "通知",
			Content:     response,
			Format:      "text",
			RawContent:  originalIntent,
			Description: "AI生成的短信内容",
		}, nil
	}

	p.logger.Infof("✅ 短信内容润色成功，字数: %d", len([]rune(result.Content)))
	return &PolishedContent{
		Channel:     ChannelSMS,
		Subject:     result.Subject,
		Content:     result.Content,
		Format:      "text",
		RawContent:  originalIntent,
		Description: result.Description,
	}, nil
}

// PolishForLark 为飞书渠道润色内容
func (p *ContentPolisher) PolishForLark(ctx context.Context, originalIntent string) (*PolishedContent, error) {
	p.logger.Infof("🦅 开始为飞书渠道润色内容")

	prompt := fmt.Sprintf(`请将以下文本内容转换为飞书交互卡片的JSON结构：

原始内容：%s

⚠️ 重要要求：
- **严格保持原文的所有信息和含义，不要添加、删除或修改任何实质内容**
- 只进行格式转换，将文本内容结构化为飞书卡片格式
- 不要改写、润色或优化原文表达

格式转换要求：
1. 生成完整的飞书卡片JSON结构，必须是有效的JSON格式
2. 根据内容类型选择合适的标题颜色：
   - "blue": 普通通知、信息提醒
   - "red": 紧急通知、警告信息
   - "green": 成功消息、完成通知
   - "orange": 待办事项、提醒
3. 正文使用Markdown格式，支持：
   - **加粗文本**
   - *斜体文本*
   - 列表（使用 - 或数字）
   - 链接 [文本](url)
4. 如果有多个信息字段，使用fields布局展示（如订单号、时间等）
5. 添加分割线(hr)分隔不同部分
6. 添加note元素显示提示信息
7. 添加交互按钮（如"查看详情"、"立即处理"等）
8. 使用emoji让内容更生动（📦📧💡✅⚠️等）

飞书卡片完整示例：
{
  "config": {
    "wide_screen_mode": true
  },
  "header": {
    "title": {
      "tag": "plain_text",
      "content": "📦 订单发货通知"
    },
    "template": "blue"
  },
  "elements": [
    {
      "tag": "div",
      "text": {
        "tag": "lark_md",
        "content": "尊敬的 **客户**，您好！\n\n您的订单已成功发货，请注意查收。"
      }
    },
    {
      "tag": "hr"
    },
    {
      "tag": "div",
      "fields": [
        {
          "is_short": true,
          "text": {
            "tag": "lark_md",
            "content": "**订单编号**\nORD123456"
          }
        },
        {
          "is_short": true,
          "text": {
            "tag": "lark_md",
            "content": "**下单时间**\n2025-12-03 10:30"
          }
        },
        {
          "is_short": true,
          "text": {
            "tag": "lark_md",
            "content": "**快递公司**\n顺丰速运"
          }
        },
        {
          "is_short": true,
          "text": {
            "tag": "lark_md",
            "content": "**运单号**\nSF1234567890"
          }
        }
      ]
    },
    {
      "tag": "hr"
    },
    {
      "tag": "note",
      "elements": [
        {
          "tag": "plain_text",
          "content": "💡 预计2-3个工作日送达，请保持手机畅通"
        }
      ]
    },
    {
      "tag": "action",
      "actions": [
        {
          "tag": "button",
          "text": {
            "tag": "plain_text",
            "content": "查看物流"
          },
          "type": "primary",
          "url": "https://example.com/track"
        },
        {
          "tag": "button",
          "text": {
            "tag": "plain_text",
            "content": "联系客服"
          },
          "type": "default",
          "url": "https://example.com/contact"
        }
      ]
    }
  ]
}

请按以下JSON格式返回：
{
  "subject": "卡片标题（简短，带emoji）",
  "content": "完整的飞书卡片JSON结构（必须是转义后的JSON字符串）",
  "description": "内容简要说明"
}

注意：
1. content字段必须是JSON字符串，不是JSON对象
2. 确保所有JSON格式正确，特别是引号和逗号
3. 根据原始意图提取关键信息，合理组织卡片结构
4. 如果有具体数据（如订单号、时间等），使用fields展示
5. 按钮URL可以使用占位符，如 https://example.com/action

只返回JSON，不要其他说明。`, originalIntent)

	response, err := p.client.SimpleChat(ctx, prompt)
	if err != nil {
		p.logger.Errorf("❌ 飞书内容润色失败: %v", err)
		return nil, fmt.Errorf("飞书内容润色失败: %w", err)
	}

	// 解析JSON响应
	var result struct {
		Subject     string `json:"subject"`
		Content     string `json:"content"`
		Description string `json:"description"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		p.logger.Warnf("⚠️ JSON解析失败，使用原始响应: %v", err)
		return &PolishedContent{
			Channel:     ChannelLark,
			Subject:     "通知",
			Content:     response,
			Format:      "json",
			RawContent:  originalIntent,
			Description: "AI生成的飞书卡片内容",
		}, nil
	}

	p.logger.Infof("✅ 飞书内容润色成功")
	return &PolishedContent{
		Channel:     ChannelLark,
		Subject:     result.Subject,
		Content:     result.Content,
		Format:      "json",
		RawContent:  originalIntent,
		Description: result.Description,
	}, nil
}

// PolishForAllChannels 为所有渠道润色内容
func (p *ContentPolisher) PolishForAllChannels(ctx context.Context, originalIntent string) (*MultiChannelContent, error) {
	p.logger.Infof("🎨 开始为所有渠道润色内容")
	p.logger.Infof("原始意图: %s", originalIntent)

	result := &MultiChannelContent{
		OriginalIntent: originalIntent,
	}

	// 并发生成三个渠道的内容
	type channelResult struct {
		channel ChannelType
		content *PolishedContent
		err     error
	}

	resultChan := make(chan channelResult, 3)

	// 邮件渠道
	go func() {
		content, err := p.PolishForEmail(ctx, originalIntent)
		resultChan <- channelResult{channel: ChannelEmail, content: content, err: err}
	}()

	// 短信渠道
	go func() {
		content, err := p.PolishForSMS(ctx, originalIntent)
		resultChan <- channelResult{channel: ChannelSMS, content: content, err: err}
	}()

	// 飞书渠道
	go func() {
		content, err := p.PolishForLark(ctx, originalIntent)
		resultChan <- channelResult{channel: ChannelLark, content: content, err: err}
	}()

	// 收集结果
	var errors []error
	for i := 0; i < 3; i++ {
		res := <-resultChan
		if res.err != nil {
			errors = append(errors, res.err)
			p.logger.Errorf("❌ 渠道 %d 润色失败: %v", res.channel, res.err)
			continue
		}

		switch res.channel {
		case ChannelEmail:
			result.EmailContent = res.content
		case ChannelSMS:
			result.SMSContent = res.content
		case ChannelLark:
			result.LarkContent = res.content
		}
	}

	if len(errors) == 3 {
		return nil, fmt.Errorf("所有渠道润色都失败了")
	}

	p.logger.Infof("✅ 多渠道内容润色完成，成功: %d/3", 3-len(errors))
	return result, nil
}

// OptimizeContent 优化已有内容
func (p *ContentPolisher) OptimizeContent(ctx context.Context, content string, channel ChannelType, requirements string) (*PolishedContent, error) {
	p.logger.Infof("✨ 开始优化内容，渠道: %d", channel)

	channelName := "邮件"
	formatType := "html"
	switch channel {
	case ChannelSMS:
		channelName = "短信"
		formatType = "text"
	case ChannelLark:
		channelName = "飞书"
		formatType = "json"
	}

	prompt := fmt.Sprintf(`请优化以下%s内容：

原始内容：
%s

优化要求：
%s

请保持原有格式类型（%s），只优化内容质量，使其更加：
1. 清晰易懂
2. 专业规范
3. 吸引人
4. 符合渠道特点

请按以下JSON格式返回：
{
  "subject": "优化后的主题",
  "content": "优化后的完整内容",
  "description": "优化说明"
}

只返回JSON，不要其他说明。`, channelName, content, requirements, formatType)

	response, err := p.client.SimpleChat(ctx, prompt)
	if err != nil {
		p.logger.Errorf("❌ 内容优化失败: %v", err)
		return nil, fmt.Errorf("内容优化失败: %w", err)
	}

	// 解析JSON响应
	var result struct {
		Subject     string `json:"subject"`
		Content     string `json:"content"`
		Description string `json:"description"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		p.logger.Warnf("⚠️ JSON解析失败，使用原始响应: %v", err)
		return &PolishedContent{
			Channel:     channel,
			Subject:     "优化后的内容",
			Content:     response,
			Format:      formatType,
			RawContent:  content,
			Description: "AI优化的内容",
		}, nil
	}

	p.logger.Infof("✅ 内容优化成功")
	return &PolishedContent{
		Channel:     channel,
		Subject:     result.Subject,
		Content:     result.Content,
		Format:      formatType,
		RawContent:  content,
		Description: result.Description,
	}, nil
}

// IsAvailable 检查润色器是否可用
func (p *ContentPolisher) IsAvailable() bool {
	return p.client != nil && p.client.IsAvailable()
}
