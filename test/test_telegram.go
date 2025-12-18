package main

// import (
// 	"flag"
// 	"fmt"
// 	"log"
// 	"nofx/logger"
// 	"os"
// 	"time"
// )

// func main() {
// 	// 解析命令行参数
// 	var botToken string
// 	var chatID int64

// 	flag.StringVar(&botToken, "token", "", "Telegram Bot Token")
// 	flag.Int64Var(&chatID, "chat-id", 0, "Telegram Chat ID")
// 	flag.Parse()

// 	// 从环境变量读取（如果命令行参数未提供）
// 	if botToken == "" {
// 		botToken = os.Getenv("TELEGRAM_BOT_TOKEN")
// 	}

// 	if chatID == 0 {
// 		// 尝试从环境变量读取
// 		chatIDStr := os.Getenv("TELEGRAM_CHAT_ID")
// 		if chatIDStr != "" {
// 			fmt.Sscanf(chatIDStr, "%d", &chatID)
// 		}
// 	}

// 	// 验证参数
// 	if botToken == "" {
// 		log.Fatal("请提供Telegram Bot Token，使用 --token 或 TELEGRAM_BOT_TOKEN 环境变量")
// 	}

// 	if chatID == 0 {
// 		log.Fatal("请提供Telegram Chat ID，使用 --chat-id 或 TELEGRAM_CHAT_ID 环境变量")
// 	}

// 	fmt.Printf("📨 正在测试发送Telegram消息...\n")
// 	fmt.Printf("   Bot Token: %s\n", botToken)
// 	fmt.Printf("   Chat ID: %d\n", chatID)

// 	// 创建Telegram发送器
// 	sender, err := logger.NewTelegramSender(botToken, chatID)
// 	if err != nil {
// 		log.Fatalf("❌ 创建Telegram发送器失败: %v", err)
// 	}
// 	defer sender.Stop()

// 	// 发送测试消息
// 	testMessage := fmt.Sprintf("✅ **测试消息**\n"+
// 		"📋 这是一条来自AI交易系统的测试消息\n"+
// 		"🔧 功能: Telegram通知测试\n"+
// 		"📝 状态: 成功\n"+
// 		"⏰ 时间: %s",
// 		time.Now().Format("2006-01-02 15:04:05"))

// 	fmt.Println("📤 发送测试消息...")
// 	sender.SendAsync(testMessage)

// 	fmt.Println("✅ 测试消息发送成功！")
// 	fmt.Println("💡 检查你的Telegram聊天，应该已经收到测试消息。")
// }
