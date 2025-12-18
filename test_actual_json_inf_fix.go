package main

import (
	"encoding/json"
	"log"

	"nofx/trader"
)

// 模拟GateFuturesTraderImpl的结构体，用于测试GetPositions返回inf值的情况
type MockGateTrader struct{}

func (m *MockGateTrader) GetBalance() (map[string]interface{}, error) {
	return map[string]interface{}{
		"totalWalletBalance":    10.51,
		"totalUnrealizedProfit": 0.0,
		"availableBalance":      10.51,
	}, nil
}

func (m *MockGateTrader) GetPositions() ([]map[string]interface{}, error) {
	return []map[string]interface{}{
		{
			"symbol":           "BTC_USDT",
			"side":             "long",
			"entryPrice":       0.0, // 这会导致+Inf错误
			"markPrice":        105.0,
			"positionAmt":      1.0,
			"leverage":         10.0,
			"unRealizedProfit": 5.0,
			"liquidationPrice": 0.0,
		},
	}, nil
}

// 实现其他必要的方法（返回空值或错误）
func (m *MockGateTrader) Init(apiKey, secretKey string) error           { return nil }
func (m *MockGateTrader) SetLeverage(symbol string, leverage int) error { return nil }
func (m *MockGateTrader) GetOrderBook(symbol string, limit int) (map[string]interface{}, error) {
	return nil, nil
}
func (m *MockGateTrader) RawPlaceOrder(params map[string]interface{}) (map[string]interface{}, error) {
	return nil, nil
}
func (m *MockGateTrader) CloseLong(symbol string, size float64) (map[string]interface{}, error) {
	return nil, nil
}
func (m *MockGateTrader) CloseShort(symbol string, size float64) (map[string]interface{}, error) {
	return nil, nil
}
func (m *MockGateTrader) GetTradeHistory(symbol string, limit int) ([]map[string]interface{}, error) {
	return nil, nil
}

func main() {
	log.Println("🔄 测试实际代码中的JSON +Inf错误修复...")

	// 创建AutoTrader实例
	at := trader.NewAutoTrader()
	at.Trader = &MockGateTrader{}
	at.InitialBalance = 10.51

	// 测试GetAccountInfo是否会导致JSON +Inf错误
	log.Println("\n📊 测试GetAccountInfo...")
	accountInfo, err := at.GetAccountInfo()
	if err != nil {
		log.Printf("❌ GetAccountInfo调用失败: %v", err)
		return
	}

	// 尝试序列化到JSON，这是最可能出现+Inf错误的地方
	log.Println("\n📊 测试JSON序列化...")
	jsonData, err := json.Marshal(accountInfo)
	if err != nil {
		log.Printf("❌ JSON序列化失败: %v", err)
		log.Println("❌ 修复失败！JSON仍包含+Inf值")
		return
	}

	log.Printf("✅ JSON序列化成功: %s", string(jsonData))
	log.Println("✅ 修复成功！JSON中不再包含+Inf值")

	// 测试GetPositions方法
	log.Println("\n📊 测试GetPositions...")
	positions, err := at.GetPositions()
	if err != nil {
		log.Printf("❌ GetPositions调用失败: %v", err)
		return
	}

	// 尝试序列化持仓数据到JSON
	posJSON, err := json.Marshal(positions)
	if err != nil {
		log.Printf("❌ 持仓JSON序列化失败: %v", err)
		log.Println("❌ 修复失败！持仓JSON仍包含+Inf值")
		return
	}

	log.Printf("✅ 持仓JSON序列化成功: %s", string(posJSON))
	log.Println("✅ 修复成功！持仓JSON中不再包含+Inf值")

	log.Println("\n🎉 所有测试通过！JSON +Inf错误已修复")
}
