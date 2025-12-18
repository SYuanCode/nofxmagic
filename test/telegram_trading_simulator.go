package main

import (
	"fmt"
	"log"
	"nofx/config"
	"nofx/logger"
	"time"
)

func main() {
	fmt.Println("📨 正在测试交易操作的Telegram推送功能...")

	// 加载配置文件
	cfg, err := config.LoadConfig("../config.json")
	if err != nil {
		log.Fatalf("❌ 加载配置文件失败: %v", err)
	}

	// 初始化Logger
	if err := logger.InitFromLogConfig(cfg.Log); err != nil {
		log.Fatalf("❌ 初始化Logger失败: %v", err)
	}

	fmt.Println("✅ Logger初始化成功")
	fmt.Println("📤 正在发送模拟交易操作通知...")
	fmt.Println("---")

	// 模拟设置止损通知
	simulateSetStopLossNotification()
	time.Sleep(1 * time.Second) // 等待1秒，避免消息发送过快

	fmt.Println("---")

	// 模拟设置止盈通知
	simulateSetTakeProfitNotification()
	time.Sleep(1 * time.Second) // 等待1秒，避免消息发送过快

	fmt.Println("---")

	// 模拟平多仓通知
	simulateCloseLongNotification()
	time.Sleep(1 * time.Second) // 等待1秒，避免消息发送过快

	fmt.Println("---")

	// 模拟平空仓通知
	simulateCloseShortNotification()

	fmt.Println("---")
	fmt.Println("✅ 所有模拟交易操作通知发送完成！")
	fmt.Println("💡 检查你的Telegram聊天，应该已经收到4条模拟交易操作通知。")
}

// simulateSetStopLossNotification 模拟设置止损通知
func simulateSetStopLossNotification() {
	fmt.Println("🎯 模拟设置止损通知...")

	// 构造设置止损消息，与binance_futures.go中的格式一致
	tgMessage := fmt.Sprintf("🎯 **止损设置成功**\n"+
		"📋 币种: `BTCUSDT`\n"+
		"🔄 方向: `LONG`\n"+
		"🛑 止损价格: `44500.0000`\n"+
		"📊 数量: `0.0200`\n"+
		"⏰ 时间: `%s`",
		time.Now().Format("2006-01-02 15:04:05"))

	// 发送通知
	logger.Info(tgMessage)
	fmt.Println("   ✅ 止损设置通知发送成功")
}

// simulateSetTakeProfitNotification 模拟设置止盈通知
func simulateSetTakeProfitNotification() {
	fmt.Println("🎯 模拟设置止盈通知...")

	// 构造设置止盈消息，与binance_futures.go中的格式一致
	tgMessage := fmt.Sprintf("🎯 **止盈设置成功**\n"+
		"📋 币种: `BTCUSDT`\n"+
		"🔄 方向: `LONG`\n"+
		"🎯 止盈价格: `46000.0000`\n"+
		"📊 数量: `0.0200`\n"+
		"⏰ 时间: `%s`",
		time.Now().Format("2006-01-02 15:04:05"))

	// 发送通知
	logger.Info(tgMessage)
	fmt.Println("   ✅ 止盈设置通知发送成功")
}

// simulateCloseLongNotification 模拟平多仓通知
func simulateCloseLongNotification() {
	fmt.Println("🔄 模拟平多仓通知...")

	// 构造平多仓消息，与binance_futures.go中的格式一致
	tgMessage := fmt.Sprintf("🔄 **平多仓成功**\n"+
		"📋 币种: `BTCUSDT`\n"+
		"📊 平仓价格: `45500.0000`\n"+
		"📈 数量: `0.0200`\n"+
		"📝 订单ID: `123456789`\n"+
		"⏰ 时间: `%s`",
		time.Now().Format("2006-01-02 15:04:05"))

	// 发送通知
	logger.Info(tgMessage)
	fmt.Println("   ✅ 平多仓通知发送成功")
}

// simulateCloseShortNotification 模拟平空仓通知
func simulateCloseShortNotification() {
	fmt.Println("🔄 模拟平空仓通知...")

	// 构造平空仓消息，与binance_futures.go中的格式一致
	tgMessage := fmt.Sprintf("🔄 **平空仓成功**\n"+
		"📋 币种: `ETHUSDT`\n"+
		"📊 平仓价格: `2150.0000`\n"+
		"📉 数量: `0.1000`\n"+
		"📝 订单ID: `987654321`\n"+
		"⏰ 时间: `%s`",
		time.Now().Format("2006-01-02 15:04:05"))

	// 发送通知
	logger.Info(tgMessage)
	fmt.Println("   ✅ 平空仓通知发送成功")
}
