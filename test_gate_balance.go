package main

import (
	"log"
	"nofx/trader"
)

// 测试Gate.io GetBalance方法，查看实际返回的数据结构
func main() {
	log.Println("🔄 开始测试Gate.io GetBalance方法...")

	// 创建Gate.io交易器实例
	gateTrader := trader.NewGateFuturesTrader(
		"", // 空API Key，会使用测试密钥
		"", // 空密钥，会使用测试密钥
		"test_user",
	)

	// 直接调用GetBalance方法
	log.Println("📦 测试获取余额...")
	balance, err := gateTrader.GetBalance()
	if err != nil {
		log.Printf("❌ 获取余额失败: %v", err)
		return
	}

	log.Printf("✅ 获取余额成功")
	log.Printf("  完整返回数据: %+v", balance)
	log.Printf("  totalWalletBalance: %v (类型: %T)", balance["totalWalletBalance"], balance["totalWalletBalance"])
	log.Printf("  availableBalance: %v (类型: %T)", balance["availableBalance"], balance["availableBalance"])
	log.Printf("  totalUnrealizedProfit: %v (类型: %T)", balance["totalUnrealizedProfit"], balance["totalUnrealizedProfit"])
}
