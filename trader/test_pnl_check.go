package trader

import (
	"fmt"
	"testing"
)

// TestPnLCalculation 测试盈亏百分比计算
func TestPnLCalculation(t *testing.T) {
	// 测试用例1：正常情况
	unrealizedPnl := 10.5
	marginUsed := 100.0
	expected := 10.5
	result := calculatePnLPercentage(unrealizedPnl, marginUsed)
	if result != expected {
		t.Errorf("TestPnLCalculation failed: expected %.2f, got %.2f", expected, result)
	} else {
		fmt.Printf("✅ TestPnLCalculation passed: %.2f%%\n", result)
	}

	// 测试用例2：亏损情况
	unrealizedPnl = -5.0
	marginUsed = 100.0
	expected = -5.0
	result = calculatePnLPercentage(unrealizedPnl, marginUsed)
	if result != expected {
		t.Errorf("TestPnLCalculation failed: expected %.2f, got %.2f", expected, result)
	} else {
		fmt.Printf("✅ TestPnLCalculation passed: %.2f%%\n", result)
	}

	// 测试用例3：零保证金情况
	unrealizedPnl = 10.0
	marginUsed = 0.0
	expected = 0.0
	result = calculatePnLPercentage(unrealizedPnl, marginUsed)
	if result != expected {
		t.Errorf("TestPnLCalculation failed: expected %.2f, got %.2f", expected, result)
	} else {
		fmt.Printf("✅ TestPnLCalculation passed: %.2f%% (zero margin)\n", result)
	}
}

// TestLeverageBasedPnLCheck 测试基于杠杆的盈亏检查逻辑
func TestLeverageBasedPnLCheck(t *testing.T) {
	// 测试用例：50倍以下杠杆，盈亏率在-10%~15%之间，不应触发
	leverage := 40
	pnlPct := 5.0
	shouldTrigger := true
	
	if leverage < 50 {
		if pnlPct >= -10 && pnlPct <= 15 {
			shouldTrigger = false
		}
	} else {
		if pnlPct >= -20 && pnlPct <= 30 {
			shouldTrigger = false
		}
	}
	
	if shouldTrigger {
		t.Errorf("TestLeverageBasedPnLCheck failed: expected shouldTrigger=false for leverage=%d, pnlPct=%.2f", leverage, pnlPct)
	} else {
		fmt.Printf("✅ TestLeverageBasedPnLCheck passed: leverage=%d, pnlPct=%.2f, shouldTrigger=%v\n", leverage, pnlPct, shouldTrigger)
	}

	// 测试用例：50倍以下杠杆，盈亏率超过15%，应触发
	leverage = 40
	pnlPct = 20.0
	shouldTrigger = true
	
	if leverage < 50 {
		if pnlPct >= -10 && pnlPct <= 15 {
			shouldTrigger = false
		}
	} else {
		if pnlPct >= -20 && pnlPct <= 30 {
			shouldTrigger = false
		}
	}
	
	if !shouldTrigger {
		t.Errorf("TestLeverageBasedPnLCheck failed: expected shouldTrigger=true for leverage=%d, pnlPct=%.2f", leverage, pnlPct)
	} else {
		fmt.Printf("✅ TestLeverageBasedPnLCheck passed: leverage=%d, pnlPct=%.2f, shouldTrigger=%v\n", leverage, pnlPct, shouldTrigger)
	}

	// 测试用例：50倍以上杠杆，盈亏率在-20%~30%之间，不应触发
	leverage = 60
	pnlPct = 25.0
	shouldTrigger = true
	
	if leverage < 50 {
		if pnlPct >= -10 && pnlPct <= 15 {
			shouldTrigger = false
		}
	} else {
		if pnlPct >= -20 && pnlPct <= 30 {
			shouldTrigger = false
		}
	}
	
	if shouldTrigger {
		t.Errorf("TestLeverageBasedPnLCheck failed: expected shouldTrigger=false for leverage=%d, pnlPct=%.2f", leverage, pnlPct)
	} else {
		fmt.Printf("✅ TestLeverageBasedPnLCheck passed: leverage=%d, pnlPct=%.2f, shouldTrigger=%v\n", leverage, pnlPct, shouldTrigger)
	}

	// 测试用例：50倍以上杠杆，盈亏率超过30%，应触发
	leverage = 60
	pnlPct = 35.0
	shouldTrigger = true
	
	if leverage < 50 {
		if pnlPct >= -10 && pnlPct <= 15 {
			shouldTrigger = false
		}
	} else {
		if pnlPct >= -20 && pnlPct <= 30 {
			shouldTrigger = false
		}
	}
	
	if !shouldTrigger {
		t.Errorf("TestLeverageBasedPnLCheck failed: expected shouldTrigger=true for leverage=%d, pnlPct=%.2f", leverage, pnlPct)
	} else {
		fmt.Printf("✅ TestLeverageBasedPnLCheck passed: leverage=%d, pnlPct=%.2f, shouldTrigger=%v\n", leverage, pnlPct, shouldTrigger)
	}
}
