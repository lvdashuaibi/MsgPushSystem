package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/lvdashuaibi/MsgPushSystem/src/pkg/ai"
	"github.com/sirupsen/logrus"
)

func main() {
	// 命令行参数
	intent := flag.String("intent", "", "原始意图")
	channel := flag.Int("channel", 0, "渠道类型 (0:全部, 1:邮件, 2:短信, 3:飞书)")
	flag.Parse()

	// 如果没有提供意图，使用默认示例
	if *intent == "" {
		*intent = "本周五凌晨2点到4点系统维护，无法登录，请提前保存数据。"
	}

	// 初始化日志
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})

	// 创建AI客户端
	aiClient := ai.NewGPTUtilsClient(logger)
	defer aiClient.Close()

	// 检查客户端是否可用
	if !aiClient.IsAvailable() {
		fmt.Println("❌ GPTUtils客户端未初始化")
		fmt.Println("请设置API_KEY环境变量:")
		fmt.Println("  export API_KEY=\"your-api-key-here\"")
		os.Exit(1)
	}

	// 创建内容润色器
	polisher := ai.NewContentPolisher(aiClient, logger)

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("🎨 MsgMate AI内容润色系统")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("📝 原始意图: %s\n", *intent)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	ctx := context.Background()

	if *channel == 0 {
		// 为所有渠道润色
		fmt.Println("🔄 正在为所有渠道生成润色内容...")
		fmt.Println()

		result, err := polisher.PolishForAllChannels(ctx, *intent)
		if err != nil {
			fmt.Printf("❌ 润色失败: %v\n", err)
			os.Exit(1)
		}

		// 显示邮件内容
		if result.EmailContent != nil {
			printChannelContent("📧 邮件版本 (HTML)", result.EmailContent)
		}

		// 显示短信内容
		if result.SMSContent != nil {
			printChannelContent("💬 短信版本 (纯文本)", result.SMSContent)
		}

		// 显示飞书内容
		if result.LarkContent != nil {
			printChannelContent("🦅 飞书版本 (JSON卡片)", result.LarkContent)
		}

	} else {
		// 为单个渠道润色
		var channelName string
		var content *ai.PolishedContent
		var err error

		switch *channel {
		case 1:
			channelName = "📧 邮件"
			fmt.Printf("🔄 正在为%s渠道生成润色内容...\n\n", channelName)
			content, err = polisher.PolishForEmail(ctx, *intent)
		case 2:
			channelName = "💬 短信"
			fmt.Printf("🔄 正在为%s渠道生成润色内容...\n\n", channelName)
			content, err = polisher.PolishForSMS(ctx, *intent)
		case 3:
			channelName = "🦅 飞书"
			fmt.Printf("🔄 正在为%s渠道生成润色内容...\n\n", channelName)
			content, err = polisher.PolishForLark(ctx, *intent)
		default:
			fmt.Println("❌ 无效的渠道类型")
			os.Exit(1)
		}

		if err != nil {
			fmt.Printf("❌ 润色失败: %v\n", err)
			os.Exit(1)
		}

		printChannelContent(channelName, content)
	}

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✅ 内容润色完成！")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func printChannelContent(title string, content *ai.PolishedContent) {
	fmt.Println("┌────────────────────────────────────────────────────────────┐")
	fmt.Printf("│ %s\n", title)
	fmt.Println("├────────────────────────────────────────────────────────────┤")
	fmt.Printf("│ 📌 主题: %s\n", content.Subject)
	fmt.Printf("│ 📄 格式: %s\n", content.Format)
	fmt.Printf("│ 💡 说明: %s\n", content.Description)
	fmt.Println("├────────────────────────────────────────────────────────────┤")
	fmt.Println("│ 📝 内容:")
	fmt.Println("│")

	// 如果是JSON格式，尝试格式化输出
	if content.Format == "json" {
		var jsonData interface{}
		if err := json.Unmarshal([]byte(content.Content), &jsonData); err == nil {
			formatted, _ := json.MarshalIndent(jsonData, "│   ", "  ")
			fmt.Printf("│   %s\n", string(formatted))
		} else {
			// 如果不是有效的JSON，直接输出
			printMultilineContent(content.Content)
		}
	} else {
		printMultilineContent(content.Content)
	}

	fmt.Println("│")
	fmt.Println("└────────────────────────────────────────────────────────────┘")
	fmt.Println()
}

func printMultilineContent(content string) {
	lines := splitLines(content)
	for _, line := range lines {
		if len(line) > 55 {
			// 长行分割
			for i := 0; i < len(line); i += 55 {
				end := i + 55
				if end > len(line) {
					end = len(line)
				}
				fmt.Printf("│   %s\n", line[i:end])
			}
		} else {
			fmt.Printf("│   %s\n", line)
		}
	}
}

func splitLines(s string) []string {
	var lines []string
	current := ""
	for _, r := range s {
		if r == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(r)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}
