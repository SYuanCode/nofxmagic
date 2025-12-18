package main

import (
	"log"
	"nofx/trader"
)

// 测试Gate.io GetPositions方法的签名修复
func main() {
	log.Println("🔄 开始测试Gate.io GetPositions方法...")

	// 创建Gate.io交易器实例
	gateTrader := trader.NewGateFuturesTrader(
		"", // 空API Key，会使用测试密钥
		"", // 空密钥，会使用测试密钥
		"test_user",
	)

	// 测试获取持仓（这个方法之前报错签名错误）
	log.Println("📦 测试获取持仓...")
	positions, err := gateTrader.GetPositions()
	if err != nil {
		log.Printf("❌ 获取持仓失败: %v", err)
		log.Printf("🔍 错误详情: %v", err)
		log.Println("💡 提示：请检查API密钥是否正确，以及网络连接是否正常")
		return
	}

	log.Printf("✅ 获取持仓成功，共 %d 个持仓", len(positions))
	for i, pos := range positions {
		log.Printf("  [%d] %s %s: %.2f @ %.2f", i+1, pos["symbol"], pos["side"], pos["positionAmt"], pos["entryPrice"])
	}
	log.Println("✅ Gate.io GetPositions方法签名修复成功！")
}
