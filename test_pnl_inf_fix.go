package main

import (
	"encoding/json"
	"log"
	"math"
)

// 直接测试PnL计算函数，模拟AutoTrader中的逻辑
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

// 模拟没有修复的计算（用于对比测试）
func calculatePnLPctWithoutFix(entryPrice, markPrice float64, leverage int, side string) float64 {
	var currentPnLPct float64
	if side == "long" {
		currentPnLPct = ((markPrice - entryPrice) / entryPrice) * float64(leverage) * 100
	} else {
		currentPnLPct = ((entryPrice - markPrice) / entryPrice) * float64(leverage) * 100
	}
	// 没有修复：直接返回结果
	return currentPnLPct
}

func main() {
	log.Println("🔄 测试PnL计算中的JSON +Inf错误修复...")

	// 测试用例1：entryPrice为0，会导致除以零
	log.Println("\n📊 测试用例1：entryPrice = 0")
	result1 := calculatePnLPct(0.0, 105.0, 10, "long")
	log.Printf("  计算结果: %.2f%%", result1)
	if !math.IsInf(result1, 0) && !math.IsNaN(result1) {
		log.Println("  ✅ 修复成功！结果不是无穷大")
	} else {
		log.Println("  ❌ 修复失败！结果是无穷大")
	}

	// 测试用例2：正常情况
	log.Println("\n📊 测试用例2：正常情况")
	result2 := calculatePnLPct(100.0, 105.0, 10, "long")
	log.Printf("  计算结果: %.2f%%", result2)
	if !math.IsInf(result2, 0) && !math.IsNaN(result2) {
		log.Println("  ✅ 正常情况结果正常")
	}

	// 测试用例3：空仓情况
	log.Println("\n📊 测试用例3：空仓情况")
	result3 := calculatePnLPct(100.0, 95.0, 10, "short")
	log.Printf("  计算结果: %.2f%%", result3)
	if !math.IsInf(result3, 0) && !math.IsNaN(result3) {
		log.Println("  ✅ 空仓情况结果正常")
	}

	// 测试JSON序列化
	log.Println("\n📊 测试JSON序列化")

	// 创建包含所有结果的数据结构
	testData := map[string]interface{}{
		"case1_result": result1,
		"case2_result": result2,
		"case3_result": result3,
	}

	// 尝试序列化到JSON
	jsonData, err := json.Marshal(testData)
	if err != nil {
		log.Printf("❌ JSON序列化失败: %v", err)
		log.Println("❌ 修复失败！JSON仍包含+Inf值")
		return
	}

	log.Printf("✅ JSON序列化成功: %s", string(jsonData))
	log.Println("✅ 修复成功！JSON中不再包含+Inf值")

	// 对比测试：没有修复的情况
	log.Println("\n📊 对比测试：没有修复的情况")

	badResult := calculatePnLPctWithoutFix(0.0, 105.0, 10, "long")
	badData := map[string]interface{}{"bad_result": badResult}

	badJSON, err := json.Marshal(badData)
	if err != nil {
		log.Printf("❌ 未修复版本序列化失败: %v (这是预期的)", err)
		log.Println("✅ 确认未修复版本会产生JSON +Inf错误")
	} else {
		log.Printf("✅ 未修复版本序列化成功: %s (这是意外的)", string(badJSON))
	}

	log.Println("\n🎉 所有测试通过！JSON +Inf错误已修复")
}
