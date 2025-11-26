package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/lvdashuaibi/MsgPushSystem/src/pkg/ai"
	"github.com/sirupsen/logrus"
)

func main() {
	// 命令行参数
	mode := flag.String("mode", "chat", "运行模式: chat(对话), analyze(分析), generate(生成), translate(翻译)")
	stream := flag.Bool("stream", false, "是否使用流式输出")
	flag.Parse()

	// 初始化日志
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})

	// 创建AI客户端
	client := ai.NewGPTUtilsClient(logger)
	defer client.Close()

	// 检查客户端是否可用
	if !client.IsAvailable() {
		fmt.Println("❌ GPTUtils客户端未初始化")
		fmt.Println("请设置API_KEY环境变量:")
		fmt.Println("  export API_KEY=\"your-api-key-here\"")
		os.Exit(1)
	}

	fmt.Println("✅ GPTUtils AI对话系统已启动")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("运行模式: %s\n", *mode)
	fmt.Printf("流式输出: %v\n", *stream)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("输入 'exit' 或 'quit' 退出程序")
	fmt.Println()

	ctx := context.Background()
	scanner := bufio.NewScanner(os.Stdin)

	switch *mode {
	case "chat":
		interactiveChat(ctx, client, scanner, *stream)
	case "analyze":
		analyzeMode(ctx, client, scanner)
	case "generate":
		generateMode(ctx, client, scanner)
	case "translate":
		translateMode(ctx, client, scanner)
	default:
		fmt.Printf("❌ 未知的运行模式: %s\n", *mode)
		os.Exit(1)
	}
}

// interactiveChat 交互式对话模式
func interactiveChat(ctx context.Context, client ai.AIClient, scanner *bufio.Scanner, stream bool) {
	fmt.Println("📝 进入对话模式")
	fmt.Println()

	for {
		fmt.Print("你: ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if input == "exit" || input == "quit" {
			fmt.Println("👋 再见！")
			break
		}

		if stream {
			fmt.Print("🤖 AI: ")
			err := client.SimpleChatStream(ctx, input, func(chunk string) error {
				fmt.Print(chunk)
				return nil
			})
			fmt.Println()
			if err != nil {
				fmt.Printf("❌ 错误: %v\n", err)
			}
		} else {
			response, err := client.SimpleChat(ctx, input)
			if err != nil {
				fmt.Printf("❌ 错误: %v\n", err)
				continue
			}
			fmt.Printf("🤖 AI: %s\n", response)
		}
		fmt.Println()
	}
}

// analyzeMode 分析模式
func analyzeMode(ctx context.Context, client ai.AIClient, scanner *bufio.Scanner) {
	fmt.Println("📊 进入分析模式")
	fmt.Println("请输入要分析的文本（输入 'END' 结束）:")
	fmt.Println()

	var text strings.Builder
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		line := scanner.Text()
		if line == "END" {
			break
		}

		if text.Len() > 0 {
			text.WriteString("\n")
		}
		text.WriteString(line)
	}

	if text.Len() == 0 {
		fmt.Println("❌ 没有输入文本")
		return
	}

	fmt.Println()
	fmt.Println("🔄 正在分析...")
	fmt.Println()

	result, err := client.AnalyzeText(ctx, text.String())
	if err != nil {
		fmt.Printf("❌ 分析失败: %v\n", err)
		return
	}

	fmt.Println("📋 分析结果:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println(result)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
}

// generateMode 生成模式
func generateMode(ctx context.Context, client ai.AIClient, scanner *bufio.Scanner) {
	fmt.Println("✍️  进入生成模式")
	fmt.Println()

	fmt.Print("请输入主题: ")
	if !scanner.Scan() {
		return
	}
	topic := strings.TrimSpace(scanner.Text())

	fmt.Print("请输入要求: ")
	if !scanner.Scan() {
		return
	}
	requirements := strings.TrimSpace(scanner.Text())

	fmt.Println()
	fmt.Println("🔄 正在生成...")
	fmt.Println()

	result, err := client.GenerateContent(ctx, topic, requirements)
	if err != nil {
		fmt.Printf("❌ 生成失败: %v\n", err)
		return
	}

	fmt.Println("📝 生成结果:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println(result)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
}

// translateMode 翻译模式
func translateMode(ctx context.Context, client ai.AIClient, scanner *bufio.Scanner) {
	fmt.Println("🌐 进入翻译模式")
	fmt.Println()

	fmt.Print("请输入目标语言 (例如: 英文, 日文, 法文): ")
	if !scanner.Scan() {
		return
	}
	targetLang := strings.TrimSpace(scanner.Text())

	fmt.Print("请输入要翻译的文本: ")
	if !scanner.Scan() {
		return
	}
	text := strings.TrimSpace(scanner.Text())

	fmt.Println()
	fmt.Println("🔄 正在翻译...")
	fmt.Println()

	result, err := client.TranslateText(ctx, text, targetLang)
	if err != nil {
		fmt.Printf("❌ 翻译失败: %v\n", err)
		return
	}

	fmt.Println("📝 翻译结果:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println(result)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
}
