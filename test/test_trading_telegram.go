package main

import (
	"flag"
	"fmt"
	"log"
	"nofx/config"
	"nofx/logger"
	"time"
)

func main() {
	// 解析命令行参数
	var configFile string
	flag.StringVar(&configFile, "config", "config.json", "配置文件路径")
	flag.Parse()

	fmt.Printf("📨 正在测试交易系统Telegram通知功能...\n")
	fmt.Printf("   配置文件: %s\n", configFile)

	// 加载配置文件
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("❌ 加载配置文件失败: %v", err)
	}

	// 初始化Logger
	if err := logger.InitFromLogConfig(cfg.Log); err != nil {
		log.Fatalf("❌ 初始化Logger失败: %v", err)
	}

	fmt.Println("✅ Logger初始化成功")
	fmt.Println("📤 正在发送模拟交易通知...")
	fmt.Println("---")

	// 模拟开多仓通知
	simulateOpenLongNotification()
	time.Sleep(1 * time.Second) // 等待1秒，避免消息发送过快

	fmt.Println("---")

	// 模拟开空仓通知
	simulateOpenShortNotification()
	time.Sleep(1 * time.Second) // 等待1秒，避免消息发送过快

	fmt.Println("---")

	// 模拟平多仓通知
	simulateCloseLongNotification()
	time.Sleep(1 * time.Second) // 等待1秒，避免消息发送过快

	fmt.Println("---")

	// 模拟平空仓通知
	simulateCloseShortNotification()

	fmt.Println("---")
	fmt.Println("✅ 所有模拟交易通知发送完成！")
	fmt.Println("💡 检查你的Telegram聊天，应该已经收到4条模拟交易通知。")
}

// simulateOpenLongNotification 模拟开多仓通知
func simulateOpenLongNotification() {
	fmt.Println("📈 模拟开多仓通知...")

	// 构造开多仓消息，与trader/auto_trader.go中的格式一致
	tgMessage := fmt.Sprintf("📈 **开多仓成功**\n"+
		"📋 币种: `BTCUSDT`\n"+
		"💰 仓位大小: `1000.00 USDT`\n"+
		"📊 当前价格: `45000.0000`\n"+
		"⚙️ 杠杆: `50x`\n"+
		"🛑 止损: `44500.0000`\n"+
		"🎯 止盈: `46000.0000`\n"+
		"📝 订单ID: `123456789`\n"+
		"⏰ 时间: `%s`",
		time.Now().Format("2006-01-02 15:04:05"))

	// 发送通知
	logger.Info(tgMessage)
	fmt.Println("   ✅ 开多仓通知发送成功")
}

// simulateOpenShortNotification 模拟开空仓通知
func simulateOpenShortNotification() {
	fmt.Println("📉 模拟开空仓通知...")

	// 构造开空仓消息，与trader/auto_trader.go中的格式一致
	tgMessage := fmt.Sprintf("📉 **开空仓成功**\n"+
		"📋 币种: `ETHUSDT`\n"+
		"💰 仓位大小: `800.00 USDT`\n"+
		"📊 当前价格: `2200.0000`\n"+
		"⚙️ 杠杆: `40x`\n"+
		"🛑 止损: `2250.0000`\n"+
		"🎯 止盈: `2100.0000`\n"+
		"📝 订单ID: `987654321`\n"+
		"⏰ 时间: `%s`",
		time.Now().Format("2006-01-02 15:04:05"))

	// 发送通知
	logger.Info(tgMessage)
	fmt.Println("   ✅ 开空仓通知发送成功")
}

// simulateCloseLongNotification 模拟平多仓通知
func simulateCloseLongNotification() {
	fmt.Println("🔄 模拟平多仓通知...")

	// 构造平多仓消息，与trader/auto_trader.go中的格式一致
	tgMessage := fmt.Sprintf("🔄 **平多仓成功**\n"+
		"📋 币种: `BTCUSDT`\n"+
		"📊 平仓价格: `45500.0000`\n"+
		"📈 开仓价格: `45000.0000`\n"+
		"💰 盈亏: `111.11 USDT`\n"+
		"📝 订单ID: `567890123`\n"+
		"⏰ 时间: `%s`",
		time.Now().Format("2006-01-02 15:04:05"))

	// 发送通知
	logger.Info(tgMessage)
	fmt.Println("   ✅ 平多仓通知发送成功")
}

// simulateCloseShortNotification 模拟平空仓通知
func simulateCloseShortNotification() {
	fmt.Println("🔄 模拟平空仓通知...")

	// 构造平空仓消息，与trader/auto_trader.go中的格式一致
	tgMessage := fmt.Sprintf("🔄 **平空仓成功**\n"+
		"📋 币种: `ETHUSDT`\n"+
		"📊 平仓价格: `2150.0000`\n"+
		"📈 开仓价格: `2200.0000`\n"+
		"💰 盈亏: `72.73 USDT`\n"+
		"📝 订单ID: `321098765`\n"+
		"⏰ 时间: `%s`",
		time.Now().Format("2006-01-02 15:04:05"))

	// 发送通知
	logger.Info(tgMessage)
	fmt.Println("   ✅ 平空仓通知发送成功")
}
