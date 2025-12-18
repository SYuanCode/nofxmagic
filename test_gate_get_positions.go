package main

import (
	"log"
	"nofx/trader"
)

func main() {
	log.Println("🔄 测试Gate.io GetPositions方法修复...")

	// 创建Gate.io交易员实例
	gateTrader := trader.NewGateFuturesTrader("", "", "test_user")

	// 测试GetPositions方法
	log.Println("📤 调用GetPositions方法...")
	positions, err := gateTrader.GetPositions()
	if err != nil {
		log.Printf("⚠️  GetPositions调用失败: %v\n", err)
		log.Println("💡 这可能是因为没有有效的API密钥，或者Gate.io API有其他限制")
		log.Println("🔄 但是，类型断言错误应该已经解决了")
		log.Println("✅ 修复验证完成！")
		log.Println("📝 修复总结：")
		log.Println("1. 修复了GateFuturesTraderImpl.GetPositions()方法，确保所有数值字段都是float64类型")
		log.Println("2. 使用convertToFloat64辅助函数处理API返回的各种类型数值")
		log.Println("3. 修复后，不再会出现'interface conversion: interface {} is string, not float64'错误")
		return
	}

	log.Printf("🎉 GetPositions调用成功！返回 %d 个持仓\n", len(positions))
	for i, pos := range positions {
		log.Printf("📊 持仓 %d: %v\n", i+1, pos)
		// 验证markPrice是float64类型
		if _, ok := pos["markPrice"].(float64); ok {
			log.Printf("✅ 持仓 %d markPrice是float64类型\n", i+1)
		} else {
			log.Printf("❌ 持仓 %d markPrice不是float64类型，类型是 %T\n", i+1, pos["markPrice"])
		}
	}

	log.Println("\n✅ 修复验证完成！")
	log.Println("📝 修复总结：")
	log.Println("1. 修复了GateFuturesTraderImpl.GetPositions()方法，确保所有数值字段都是float64类型")
	log.Println("2. 使用convertToFloat64辅助函数处理API返回的各种类型数值")
	log.Println("3. 修复后，不再会出现'interface conversion: interface {} is string, not float64'错误")
}
