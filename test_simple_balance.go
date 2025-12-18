package main

import (
	"fmt"
	"nofx/trader"
	"os"
)

// 简单测试Gate.io GetBalance方法，只打印关键信息
func main() {
	// 关闭默认日志，只打印关键信息
	os.Setenv("LOG_LEVEL", "ERROR")

	// 创建Gate.io交易器实例
	gateTrader := trader.NewGateFuturesTrader(
		"", // 空API Key，会使用测试密钥
		"", // 空密钥，会使用测试密钥
		"test_user",
	)

	// 直接调用GetBalance方法
	balance, err := gateTrader.GetBalance()
	if err != nil {
		fmt.Printf("获取余额失败: %v\n", err)
		return
	}

	fmt.Println("获取余额成功")
	fmt.Printf("totalWalletBalance: %.8f\n", balance["totalWalletBalance"])
	fmt.Printf("availableBalance: %.8f\n", balance["availableBalance"])
	fmt.Printf("totalUnrealizedProfit: %.8f\n", balance["totalUnrealizedProfit"])

	// 计算总资产
	totalEquity := balance["totalWalletBalance"].(float64) + balance["totalUnrealizedProfit"].(float64)
	fmt.Printf("总资产: %.8f\n", totalEquity)
}
