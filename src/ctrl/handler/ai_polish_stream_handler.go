package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lvdashuaibi/MsgPushSystem/src/constant"
	"github.com/lvdashuaibi/MsgPushSystem/src/pkg/ai"
	log "github.com/lvdashuaibi/MsgPushSystem/src/pkg/log"
)

// PolishStreamRequest 流式润色请求
type PolishStreamRequest struct {
	OriginalIntent string `json:"original_intent" binding:"required"` // 原始意图
	Channel        int    `json:"channel" binding:"required"`         // 渠道类型 (1:邮件, 2:短信, 3:飞书)
}

// StreamEvent SSE事件
type StreamEvent struct {
	Event string      `json:"event"` // 事件类型: start, chunk, complete, error
	Data  interface{} `json:"data"`  // 事件数据
}

// PolishContentStream 通用内容润色流式处理
// @Summary 通用内容润色
// @Description 使用SSE流式返回AI润色内容（优化文字表达）
// @Tags AI润色
// @Accept json
// @Produce text/event-stream
// @Param original_intent query string true "原始内容"
// @Success 200 {string} string "SSE流"
// @Router /ai/polish/content [get]
func (h *AIPolishHandler) PolishContentStream(c *gin.Context) {
	// 从query参数获取数据
	originalIntent := c.Query("original_intent")

	if originalIntent == "" {
		c.JSON(http.StatusOK, PolishResponse{
			Code: constant.ERR_INPUT_INVALID,
			Msg:  "原始内容不能为空",
		})
		return
	}

	log.Infof("✨ 收到内容润色请求，原始内容: %s", originalIntent)

	// 检查润色器是否可用
	if !h.polisher.IsAvailable() {
		log.Error("AI润色器不可用")
		c.JSON(http.StatusOK, PolishResponse{
			Code: constant.ERR_INTERNAL,
			Msg:  "AI服务暂时不可用，请稍后重试",
		})
		return
	}

	// 设置SSE响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// 发送开始事件
	sendSSE(c.Writer, "start", map[string]interface{}{
		"message": "开始优化内容...",
	})
	c.Writer.Flush()

	ctx := c.Request.Context()

	// 构建提示词
	prompt := buildPolishPrompt(originalIntent)

	log.Infof("📝 开始流式优化内容")

	// 累积的内容
	var accumulatedContent string
	var subject string
	var description string

	// 创建临时AI客户端用于流式调用
	logger := log.GetLogger()
	aiClient := ai.NewGPTUtilsClient(logger)
	defer aiClient.Close()

	// 使用流式API
	err := aiClient.SimpleChatStream(ctx, prompt, func(chunk string) error {
		accumulatedContent += chunk

		// 发送chunk事件
		sendSSE(c.Writer, "chunk", map[string]interface{}{
			"content": chunk,
			"total":   accumulatedContent,
		})
		c.Writer.Flush()

		// 检查客户端是否断开连接
		select {
		case <-ctx.Done():
			return io.EOF
		default:
		}

		return nil
	})

	if err != nil {
		log.Errorf("❌ 内容润色失败: %v", err)
		sendSSE(c.Writer, "error", map[string]interface{}{
			"message": "内容优化失败: " + err.Error(),
		})
		c.Writer.Flush()
		return
	}

	// 尝试解析JSON响应
	var result struct {
		Subject     string `json:"subject"`
		Content     string `json:"content"`
		Description string `json:"description"`
	}

	if err := json.Unmarshal([]byte(accumulatedContent), &result); err == nil {
		subject = result.Subject
		description = result.Description
		accumulatedContent = result.Content
	} else {
		// 如果不是JSON，使用默认值
		subject = "优化内容"
		description = "AI优化生成"
	}

	// 发送完成事件
	polishedContent := &ai.PolishedContent{
		Channel:     ai.ChannelSMS,
		Subject:     subject,
		Content:     accumulatedContent,
		Format:      "text",
		RawContent:  originalIntent,
		Description: description,
	}

	sendSSE(c.Writer, "complete", polishedContent)
	c.Writer.Flush()

	log.Infof("✅ 内容润色完成")
}

// PolishForSingleChannelStream 单渠道流式润色
// @Summary 单渠道流式润色
// @Description 使用SSE流式返回AI润色内容
// @Tags AI润色
// @Accept json
// @Produce text/event-stream
// @Param request body PolishStreamRequest true "流式润色请求"
// @Success 200 {string} string "SSE流"
// @Router /ai/polish/stream [post]
func (h *AIPolishHandler) PolishForSingleChannelStream(c *gin.Context) {
	// 从query参数获取数据
	originalIntent := c.Query("original_intent")
	channel := c.DefaultQuery("channel", "1")

	if originalIntent == "" {
		c.JSON(http.StatusOK, PolishResponse{
			Code: constant.ERR_INPUT_INVALID,
			Msg:  "原始意图不能为空",
		})
		return
	}

	// 转换channel为int
	var channelInt int
	fmt.Sscanf(channel, "%d", &channelInt)

	if channelInt < 1 || channelInt > 3 {
		c.JSON(http.StatusOK, PolishResponse{
			Code: constant.ERR_INPUT_INVALID,
			Msg:  "渠道类型无效，必须是1(邮件)、2(短信)或3(飞书)",
		})
		return
	}

	req := PolishStreamRequest{
		OriginalIntent: originalIntent,
		Channel:        channelInt,
	}

	log.Infof("🎨 收到流式润色请求，渠道: %d，原始意图: %s", req.Channel, req.OriginalIntent)

	// 检查润色器是否可用
	if !h.polisher.IsAvailable() {
		log.Error("AI润色器不可用")
		c.JSON(http.StatusOK, PolishResponse{
			Code: constant.ERR_INTERNAL,
			Msg:  "AI服务暂时不可用，请稍后重试",
		})
		return
	}

	// 设置SSE响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// 发送开始事件
	sendSSE(c.Writer, "start", map[string]interface{}{
		"channel": req.Channel,
		"message": "开始生成内容...",
	})
	c.Writer.Flush()

	ctx := c.Request.Context()

	// 构建提示词
	var prompt string
	channelName := ""
	formatType := ""

	switch ai.ChannelType(req.Channel) {
	case ai.ChannelEmail:
		channelName = "邮件"
		formatType = "html"
		prompt = buildEmailPrompt(req.OriginalIntent)
	case ai.ChannelSMS:
		channelName = "短信"
		formatType = "text"
		prompt = buildSMSPrompt(req.OriginalIntent)
	case ai.ChannelLark:
		channelName = "飞书"
		formatType = "json"
		prompt = buildLarkPrompt(req.OriginalIntent)
	}

	log.Infof("📝 开始流式生成%s内容", channelName)

	// 累积的内容
	var accumulatedContent string
	var subject string
	var description string

	// 创建临时AI客户端用于流式调用
	logger := log.GetLogger()
	aiClient := ai.NewGPTUtilsClient(logger)
	defer aiClient.Close()

	// 使用流式API
	err := aiClient.SimpleChatStream(ctx, prompt, func(chunk string) error {
		accumulatedContent += chunk

		// 发送chunk事件
		sendSSE(c.Writer, "chunk", map[string]interface{}{
			"content": chunk,
			"total":   accumulatedContent,
		})
		c.Writer.Flush()

		// 检查客户端是否断开连接
		select {
		case <-ctx.Done():
			return io.EOF
		default:
		}

		return nil
	})

	if err != nil {
		log.Errorf("❌ 流式润色失败: %v", err)
		sendSSE(c.Writer, "error", map[string]interface{}{
			"message": "内容生成失败: " + err.Error(),
		})
		c.Writer.Flush()
		return
	}

	// 尝试解析JSON响应
	var result struct {
		Subject     string `json:"subject"`
		Content     string `json:"content"`
		Description string `json:"description"`
	}

	if err := json.Unmarshal([]byte(accumulatedContent), &result); err == nil {
		subject = result.Subject
		description = result.Description
		accumulatedContent = result.Content
	} else {
		// 如果不是JSON，使用默认值
		subject = "AI生成的" + channelName + "内容"
		description = "AI润色生成"
	}

	// 发送完成事件
	polishedContent := &ai.PolishedContent{
		Channel:     ai.ChannelType(req.Channel),
		Subject:     subject,
		Content:     accumulatedContent,
		Format:      formatType,
		RawContent:  req.OriginalIntent,
		Description: description,
	}

	sendSSE(c.Writer, "complete", polishedContent)
	c.Writer.Flush()

	log.Infof("✅ 流式润色完成")
}

// sendSSE 发送SSE事件
func sendSSE(w gin.ResponseWriter, event string, data interface{}) {
	eventData := StreamEvent{
		Event: event,
		Data:  data,
	}

	jsonData, err := json.Marshal(eventData)
	if err != nil {
		log.Errorf("JSON序列化失败: %v", err)
		return
	}

	fmt.Fprintf(w, "data: %s\n\n", jsonData)
}

// buildPolishPrompt 构建通用内容润色提示词
func buildPolishPrompt(originalIntent string) string {
	return fmt.Sprintf(`你是一个专业的内容润色助手。你的任务是优化用户提供的文本内容，使其表达更专业、更吸引人。

【核心原则】
1. 保持原文的核心含义和信息
2. 优化语言表达，使其更专业、更吸引人
3. 改进句子结构，提高可读性
4. 不添加任何签名或前缀
5. 不改变原文的长度过多

原始内容：%s

【具体要求】
1. 优化语言表达，使其更专业、更吸引人
2. 改进句子结构，提高可读性
3. 适当添加过渡词，使逻辑更清晰
4. 使用更恰当的词汇和表述方式
5. 保持原文的语气和风格
6. 不添加任何签名或前缀
7. 不改变原文的长度过多

【禁止事项】
- 不能改变原文的核心意思
- 不能添加原文中没有的信息
- 不能添加虚假的日期、时间、地点等信息
- 不能添加虚假的人名、部门名称等

请按以下JSON格式返回：
{
  "subject": "内容主题或标题（如果有的话）",
  "content": "优化后的内容",
  "description": "优化说明"
}

只返回JSON，不要其他说明。`, originalIntent)
}

// buildEmailPrompt 构建邮件提示词
func buildEmailPrompt(originalIntent string) string {
	return fmt.Sprintf(`你是一个专业的邮件内容润色助手。你的任务是将用户提供的原始意图转换为专业的邮件内容。

【核心原则】
1. 严格遵循原始意图，只进行润色和格式化，不能添加、删除或改变原意
2. 不能编造与原文无关的内容
3. 不能添加虚假信息或假设
4. 只能基于原始意图中已有的信息进行扩展和优化

原始意图：%s

【具体要求】
1. 生成完整的HTML格式邮件内容
2. 包含适当的标题（使用<h2>标签）
3. 正文分段清晰，使用<p>标签
4. 重要信息使用<strong>标签加粗强调
5. 语气正式、专业、详尽
6. 包含适当的问候语和落款
7. 使用合适的HTML样式，使邮件美观易读
8. 【严格要求】不能出现"XXX"、"某某"、"[待填写]"等占位符，所有内容必须具体明确
9. 【严格要求】不能添加原始意图中没有的信息，只能对已有信息进行润色

【禁止事项】
- 不能添加虚假的日期、时间、地点等信息
- 不能添加虚假的人名、部门名称等
- 不能添加原文中没有提及的细节
- 不能改变原文的核心意思

请按以下JSON格式返回：
{
  "subject": "邮件主题",
  "content": "完整的HTML邮件内容",
  "description": "内容简要说明"
}

只返回JSON，不要其他说明。`, originalIntent)
}

// buildSMSPrompt 构建短信提示词
func buildSMSPrompt(originalIntent string) string {
	return fmt.Sprintf(`你是一个专业的短信内容润色助手。你的任务是将用户提供的原始意图转换为简洁的短信内容。

【核心原则】
1. 严格遵循原始意图，只进行润色和格式化，不能添加、删除或改变原意
2. 不能编造与原文无关的内容
3. 不能添加虚假信息或假设
4. 只能基于原始意图中已有的信息进行优化

原始意图：%s

【具体要求】
1. 纯文本格式，不使用HTML或Markdown
2. 以【MsgMate】开头作为签名
3. 字数控制在70字以内
4. 保留所有关键信息（时间、地点、事项）
5. 语言简洁明了，易于理解
6. 使用温馨、友好的语气
7. 重要数字和时间使用阿拉伯数字
8. 适当使用标点符号分隔信息
9. 【严格要求】不能出现"XXX"、"某某"、"[待填写]"等占位符，所有内容必须具体明确
10. 【严格要求】不能添加原始意图中没有的信息，只能对已有信息进行优化

【禁止事项】
- 不能添加虚假的日期、时间、地点等信息
- 不能添加虚假的人名、部门名称等
- 不能添加原文中没有提及的细节
- 不能改变原文的核心意思

请按以下JSON格式返回：
{
  "subject": "短信主题（简短）",
  "content": "完整的短信内容（包含【MsgMate】签名）",
  "description": "内容简要说明"
}

只返回JSON，不要其他说明。`, originalIntent)
}

// buildLarkPrompt 构建飞书提示词
func buildLarkPrompt(originalIntent string) string {
	return fmt.Sprintf(`请将以下原始意图转换为飞书交互卡片的JSON结构：

原始意图：%s

要求：
1. 生成完整的飞书卡片JSON结构
2. 标题使用红色警告色（如果是重要通知）或蓝色（普通通知）
3. 正文使用Markdown格式，支持加粗、列表等
4. 包含一个"查看详情"或"了解更多"的按钮
5. 卡片结构清晰，信息层次分明
6. 使用飞书卡片的标准JSON格式

请按以下JSON格式返回：
{
  "subject": "卡片标题",
  "content": "完整的飞书卡片JSON结构（字符串形式）",
  "description": "内容简要说明"
}

只返回JSON，不要其他说明。`, originalIntent)
}
