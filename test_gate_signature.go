package main

import (
	"log"
	"nofx/trader"
)

// 测试Gate.io API签名修复
func main() {
	log.Println("🔄 开始测试Gate.io API签名...")

	// 创建Gate.io交易器实例
	gateTrader := trader.NewGateFuturesTrader(
		"", // 空API Key，会使用测试密钥
		"", // 空密钥，会使用测试密钥
		"test_user",
	)

	// 测试获取余额（这个方法会调用签名生成逻辑）
	log.Println("📦 测试获取余额...")
	balance, err := gateTrader.GetBalance()
	if err != nil {
		log.Printf("❌ 获取余额失败: %v", err)
		log.Printf("🔍 错误详情: %v", err)
		log.Println("💡 提示：请检查API密钥是否正确，以及网络连接是否正常")
		return
	}

	log.Printf("✅ 获取余额成功: %+v", balance)
	log.Println("✅ Gate.io API签名修复成功！")
}
