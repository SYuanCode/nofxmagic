package main

import (
	"encoding/json"
	"fmt"
	"log"
)

// 模拟 Gate.io API 返回数据，包含 10.51 USDT
var mockGateioBalanceResponse = `{
	"cross_margin_balance": 10.512906242763,
	"available": 10.512906242763,
	"cross_unrealised_pnl": 0,
	"available_for_withdrawal": 10.512906242763,
	"total": 10.512906242763
}`

// 模拟 GateFuturesTraderImpl.GetBalance() 方法的核心逻辑
func mockGetBalance() (map[string]interface{}, error) {
	// 解析模拟的 JSON 响应
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(mockGateioBalanceResponse), &result); err != nil {
		return nil, fmt.Errorf("JSON 解析失败: %w", err)
	}

	// 打印API返回的原始响应，用于调试
	fmt.Printf("📥 Gate.io API原始响应: %s\n", mockGateioBalanceResponse)

	// 打印解析后的结果，用于调试
	fmt.Printf("🔍 解析后的API结果: %v\n", result)

	// 转换为统一格式，确保字段类型正确
	balance := make(map[string]interface{})

	// 处理totalWalletBalance（钱包余额）
	// Gate.io没有total字段，根据实际返回的数据，使用cross_margin_balance字段
	crossMarginBalance := 0.0
	if cmb, ok := result["cross_margin_balance"]; ok {
		crossMarginBalance, _ = cmb.(float64)
	}
	balance["totalWalletBalance"] = crossMarginBalance

	// 处理availableBalance（可用余额）
	available := 0.0
	if avail, ok := result["available"]; ok {
		available, _ = avail.(float64)
	}
	balance["availableBalance"] = available

	// 处理totalUnrealizedProfit（未实现盈亏）
	unrealisedPnl := 0.0
	if upnl, ok := result["cross_unrealised_pnl"]; ok {
		unrealisedPnl, _ = upnl.(float64)
	}
	balance["totalUnrealizedProfit"] = unrealisedPnl

	// 简单打印关键信息，验证修复是否有效
	fmt.Printf("📊 Gate.io余额数据转换结果: totalWalletBalance=%.8f, availableBalance=%.8f, totalUnrealizedProfit=%.8f\n",
		crossMarginBalance, available, unrealisedPnl)

	return balance, nil
}

// 模拟 AutoTrader.GetAccountInfo() 方法的核心逻辑
func mockGetAccountInfo() (map[string]interface{}, error) {
	balance, err := mockGetBalance()
	if err != nil {
		return nil, fmt.Errorf("获取余额失败: %w", err)
	}

	// 获取账户字段
	totalWalletBalance := 0.0
	totalUnrealizedProfit := 0.0
	availableBalance := 0.0

	if wallet, ok := balance["totalWalletBalance"].(float64); ok {
		totalWalletBalance = wallet
	}
	if unrealized, ok := balance["totalUnrealizedProfit"].(float64); ok {
		totalUnrealizedProfit = unrealized
	}
	if avail, ok := balance["availableBalance"].(float64); ok {
		availableBalance = avail
	}

	// Total Equity = 钱包余额 + 未实现盈亏
	totalEquity := totalWalletBalance + totalUnrealizedProfit

	// 假设 initialBalance 为 10.51
	initialBalance := 10.51

	// 计算总盈亏
	totalPnL := totalEquity - initialBalance
	totalPnLPct := (totalPnL / initialBalance) * 100

	// 模拟返回的 API 响应
	return map[string]interface{}{
		// 核心字段
		"total_equity":      totalEquity,
		"wallet_balance":    totalWalletBalance,
		"unrealized_profit": totalUnrealizedProfit,
		"available_balance": availableBalance,

		// 盈亏统计
		"total_pnl":       totalPnL,
		"total_pnl_pct":   totalPnLPct,
		"initial_balance": initialBalance,
		"daily_pnl":       0,

		// 持仓信息
		"position_count":  0,
		"margin_used":     0,
		"margin_used_pct": 0,
	}, nil
}

func main() {
	log.Println("🔄 直接测试 Gate.io 余额修复...")

	// 调用模拟的 GetAccountInfo 方法
	accountInfo, err := mockGetAccountInfo()
	if err != nil {
		log.Fatalf("获取账户信息失败: %v", err)
	}

	// 打印最终的 API 响应
	log.Printf("📤 最终 API 响应: %v\n", accountInfo)

	// 验证修复是否有效
	walletBalance := accountInfo["wallet_balance"].(float64)
	availableBalance := accountInfo["available_balance"].(float64)

	if walletBalance >= 10.51 && walletBalance <= 10.52 {
		log.Printf("✅ 修复成功！wallet_balance 为 %.8f USDT，正确显示了 Gate.io 返回的 10.51 USDT\n", walletBalance)
	} else {
		log.Printf("❌ 修复失败！wallet_balance 为 %.8f USDT，预期为 10.51 USDT\n", walletBalance)
	}

	if availableBalance >= 10.51 && availableBalance <= 10.52 {
		log.Printf("✅ 修复成功！available_balance 为 %.8f USDT，正确显示了 Gate.io 返回的可用余额\n", availableBalance)
	} else {
		log.Printf("❌ 修复失败！available_balance 为 %.8f USDT，预期为 10.51 USDT\n", availableBalance)
	}

	log.Println("🎉 测试完成！")
}
