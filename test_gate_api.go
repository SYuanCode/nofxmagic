package main

import (
	"fmt"
	"log"
	"os"

	"E:/study/gocode/src/npfx/solo-nofx/nofx/trader"
)

func main() {
	// 设置日志格式
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(os.Stdout)

	fmt.Println("=== Gate.io API 测试脚本 ===")

	// 创建Gate.io交易器实例
	gateTrader := trader.NewGateFuturesTrader(
		"", // 空API Key，会被内部硬编码覆盖
		"", // 空密钥，会被内部硬编码覆盖
		"test_user",
	)

	fmt.Println("🔄 开始测试获取余额...")
	// 测试获取余额
	balance, err := gateTrader.GetBalance()
	if err != nil {
		fmt.Printf("❌ 获取余额失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 获取余额成功: %+v\n", balance)
	fmt.Println("=== 测试完成 ===")
}
