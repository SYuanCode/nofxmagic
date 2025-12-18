package main

import (
	"log"
	"nofx/trader"
)

func main() {
	// 模拟Gate.io API返回的JSON响应，包含字符串类型的数值
	mockResponse := `{
		"cross_margin_balance": "10.512906242763",
		"available": "10.512906242763",
		"cross_unrealised_pnl": "0.0"
	}`

	log.Println("🔄 测试修复后的MockGateIOGetBalance方法...")
	log.Printf("📥 模拟API响应: %s\n", mockResponse)

	// 调用修复后的MockGateIOGetBalance方法
	balance, err := trader.MockGateIOGetBalance(mockResponse)
	if err != nil {
		log.Fatalf("❌ MockGateIOGetBalance调用失败: %v", err)
	}

	log.Printf("📊 修复后结果: %v\n", balance)

	// 检查结果是否正确
	if totalWalletBalance, ok := balance["totalWalletBalance"].(float64); ok {
		log.Printf("✅ totalWalletBalance: %.8f", totalWalletBalance)
		if totalWalletBalance == 10.51290624 {
			log.Println("🎉 修复成功！totalWalletBalance正确显示为10.51290624")
		} else {
			log.Printf("❌ 修复失败！totalWalletBalance应为10.51290624，但实际为%.8f", totalWalletBalance)
		}
	}

	// 模拟GetAccountInfo转换
	log.Println("\n🔄 模拟GetAccountInfo转换...")
	accountInfo, err := trader.MockGateIOGetAccountInfo(balance, 10.51)
	if err != nil {
		log.Fatalf("❌ MockGateIOGetAccountInfo调用失败: %v", err)
	}

	log.Printf("📊 GetAccountInfo结果: %v\n", accountInfo)

	if availableBalance, ok := accountInfo["available_balance"].(float64); ok {
		log.Printf("✅ available_balance: %.8f", availableBalance)
		if availableBalance == 10.51290624 {
			log.Println("🎉 修复成功！available_balance正确显示为10.51290624")
		} else {
			log.Printf("❌ 修复失败！available_balance应为10.51290624，但实际为%.8f", availableBalance)
		}
	}

	log.Println("\n✅ 测试完成！")
}
