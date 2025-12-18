package main

import (
	"encoding/json"
	"log"
	"math"
)

// 模拟持仓数据结构
type Position struct {
	EntryPrice float64 `json:"entryPrice"`
	MarkPrice  float64 `json:"markPrice"`
	Symbol     string  `json:"symbol"`
	Side       string  `json:"side"`
	Leverage   float64 `json:"leverage"`
}

// 模拟AutoTrader中的PnL计算
func calculatePnLPct(entryPrice, markPrice float64, leverage int, side string) float64 {
	var currentPnLPct float64
	if side == "long" {
		currentPnLPct = ((markPrice - entryPrice) / entryPrice) * float64(leverage) * 100
	} else {
		currentPnLPct = ((entryPrice - markPrice) / entryPrice) * float64(leverage) * 100
	}

	// 修复前：没有检查
	// 修复后：添加检查
	if math.IsNaN(currentPnLPct) || math.IsInf(currentPnLPct, 0) {
		return 0.0
	}
	return currentPnLPct
}

func main() {
	log.Println("🔄 测试JSON +Inf错误修复...")

	// 测试用例1：正常情况
	log.Println("\n📊 测试用例1：正常情况")
	pos1 := Position{
		EntryPrice: 100.0,
		MarkPrice:  105.0,
		Symbol:     "BTC_USDT",
		Side:       "long",
		Leverage:   10.0,
	}
	pnl1 := calculatePnLPct(pos1.EntryPrice, pos1.MarkPrice, int(pos1.Leverage), pos1.Side)
	log.Printf("  结果: %.2f%% (正常情况)", pnl1)

	// 测试用例2：entryPrice为0，会导致除以零，产生+Inf
	log.Println("\n📊 测试用例2：entryPrice为0 (会产生+Inf)")
	pos2 := Position{
		EntryPrice: 0.0,
		MarkPrice:  105.0,
		Symbol:     "BTC_USDT",
		Side:       "long",
		Leverage:   10.0,
	}
	pnl2 := calculatePnLPct(pos2.EntryPrice, pos2.MarkPrice, int(pos2.Leverage), pos2.Side)
	log.Printf("  结果: %.2f%% (修复后应该返回0.0)", pnl2)

	// 测试用例3：markPrice远大于entryPrice，可能产生很大数值
	log.Println("\n📊 测试用例3：markPrice远大于entryPrice")
	pos3 := Position{
		EntryPrice: 1.0,
		MarkPrice:  1000000.0,
		Symbol:     "BTC_USDT",
		Side:       "long",
		Leverage:   100.0,
	}
	pnl3 := calculatePnLPct(pos3.EntryPrice, pos3.MarkPrice, int(pos3.Leverage), pos3.Side)
	log.Printf("  结果: %.2f%% (正常大数值)", pnl3)

	// 测试JSON序列化
	log.Println("\n📊 测试JSON序列化")

	// 测试修复后的PnL可以正常序列化
	testData := map[string]interface{}{
		"symbol":     "BTC_USDT",
		"side":       "long",
		"pnl_pct":    pnl1,
		"fixed_pnl2": pnl2,
		"large_pnl":  pnl3,
	}

	jsonData, err := json.Marshal(testData)
	if err != nil {
		log.Printf("❌ JSON序列化失败: %v", err)
	} else {
		log.Printf("✅ JSON序列化成功: %s", string(jsonData))
	}

	// 测试直接使用+Inf会失败
	log.Println("\n📊 测试直接使用+Inf")
	infData := map[string]interface{}{
		"inf_value": math.Inf(1),
	}

	infJSON, err := json.Marshal(infData)
	if err != nil {
		log.Printf("❌ 直接使用+Inf序列化失败: %v (这是预期的)", err)
	} else {
		log.Printf("✅ 直接使用+Inf序列化成功: %s (这是意外的)", string(infJSON))
	}

	log.Println("\n✅ JSON +Inf错误修复测试完成！")
}
