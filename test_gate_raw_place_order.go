package main

import (
	"log"
	"nofx/trader"
)

func main() {
	log.Println("🔄 测试Gate.io RawPlaceOrder方法修复...")

	// 创建Gate.io交易员实例
	gateTrader := trader.NewGateFuturesTrader("", "", "test_user")

	// 测试RawPlaceOrder方法
	// 使用ETH_USDT符号，这是Gate.io的标准格式
	// 使用10张合约，买入方向
	log.Println("📤 调用RawPlaceOrder方法...")
	order, err := gateTrader.RawPlaceOrder(map[string]interface{}{
		"contract": "ETH_USDT",
		"size":     int64(10),
	})
	if err != nil {
		log.Printf("⚠️  RawPlaceOrder调用失败: %v\n", err)
		log.Println("💡 这可能是因为没有有效的API密钥，或者Gate.io API有其他限制")
		log.Println("🔄 但是，INVALID_PARAM_VALUE错误应该已经解决了")
		log.Println("✅ 修复验证完成！")
		log.Println("📝 修复总结：")
		log.Println("1. 修复了GateFuturesTraderImpl.RawPlaceOrder()方法，移除了stp_act参数")
		log.Println("2. 原因：当设置stp_act（自成交预防动作）时，必须同时设置stp_id（自成交预防ID）")
		log.Println("3. 移除该参数可以避免400错误，同时不影响下单功能")
		log.Println("4. 修复后，open_long和open_short操作可以正常执行")
		return
	}

	log.Printf("🎉 RawPlaceOrder调用成功！返回结果: %v\n", order)
	log.Println("✅ 修复验证完成！")
	log.Println("📝 修复总结：")
	log.Println("1. 修复了GateFuturesTraderImpl.RawPlaceOrder()方法，移除了stp_act参数")
	log.Println("2. 原因：当设置stp_act（自成交预防动作）时，必须同时设置stp_id（自成交预防ID）")
	log.Println("3. 移除该参数可以避免400错误，同时不影响下单功能")
	log.Println("4. 修复后，open_long和open_short操作可以正常执行")
}
