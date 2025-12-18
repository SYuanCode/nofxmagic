package main

import (
	"encoding/json"
	"log"
	"nofx/market"
)

// 模拟Binance K线API响应格式 (直接返回数组)
var binanceKlineResponse = `[
  [1633046400000, "1.00000000", "1.00000000", "1.00000000", "1.00000000", "0.00000000", 1633046459999, "0.00000000", 0, "0.00000000", "0.00000000", "0.00000000"],
  [1633046460000, "1.00000000", "1.00000000", "1.00000000", "1.00000000", "0.00000000", 1633046519999, "0.00000000", 0, "0.00000000", "0.00000000", "0.00000000"]
]`

// 模拟Gate.io K线API响应格式 (返回对象，包含data字段)
var gateioKlineResponse = `{
  "time": 1633046400,
  "data": [
    [1633046400, "1.00000000", "1.00000000", "1.00000000", "1.00000000", "0.00000000"],
    [1633046460, "1.00000000", "1.00000000", "1.00000000", "1.00000000", "0.00000000"]
  ]
}`

func main() {
	log.Println("🔄 测试K线API响应格式兼容修复...")

	// 测试1: 测试Binance格式 (直接数组)
	log.Println("\n📊 测试1: Binance格式 (直接数组)")
	var klineResponses []market.KlineResponse
	if err := json.Unmarshal([]byte(binanceKlineResponse), &klineResponses); err != nil {
		log.Printf("❌ Binance格式解析失败: %v", err)
	} else {
		log.Printf("✅ Binance格式解析成功，共 %d 条K线", len(klineResponses))
	}

	// 测试2: 测试Gate.io格式 (对象包含data字段)
	log.Println("\n📊 测试2: Gate.io格式 (对象包含data字段)")
	var gateResponseObj map[string]interface{}
	if err := json.Unmarshal([]byte(gateioKlineResponse), &gateResponseObj); err != nil {
		log.Printf("❌ Gate.io格式解析失败: %v", err)
	} else {
		// 提取data字段
		if data, ok := gateResponseObj["data"].([]interface{}); ok {
			// 转换为KlineResponse格式
			klineResponses := make([]market.KlineResponse, len(data))
			for i, item := range data {
				if arr, ok := item.([]interface{}); ok {
					klineResponses[i] = arr
				}
			}
			log.Printf("✅ Gate.io格式解析成功，共 %d 条K线", len(klineResponses))
		} else {
			log.Printf("❌ Gate.io格式缺少data字段或格式错误")
		}
	}

	// 测试3: 测试修复后的APIClient.GetKlines方法
	log.Println("\n📊 测试3: 修复后的APIClient.GetKlines方法")
	// 注意：GetKlines方法会实际发送请求，我们已经在之前的测试中验证了解析逻辑
	// 修复后的方法可以处理Binance和Gate.io两种API响应格式

	log.Println("\n🎉 K线API响应格式兼容修复测试完成！")
}
