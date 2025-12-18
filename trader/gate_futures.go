package trader

import (
	"fmt"
	"log"
	"math"
)

// ================================
// 枚举 & 类型定义
// ================================

// FuturesAction 合约支持的动作（统一枚举，兼容Binance和Gate.io）
type FuturesAction string

const (
	ActionOpenLong   FuturesAction = "open_long"
	ActionOpenShort  FuturesAction = "open_short"
	ActionCloseLong  FuturesAction = "close_long"
	ActionCloseShort FuturesAction = "close_short"
	ActionPartial    FuturesAction = "partial_close"
)

// FuturesOrderRequest 统一下单请求（统一结构体，兼容Binance和Gate.io）
type FuturesOrderRequest struct {
	Symbol          string
	Action          FuturesAction
	PositionSizeUSD float64 // 只在开仓时使用
	ClosePercentage float64 // 只在 partial_close 使用
	Leverage        int

	StopLoss   float64
	TakeProfit float64
}

// ================================
// 核心：USD → 合约张数
// ================================

// CalcContracts
// 计算合约张数（兼容Binance和Gate.io）
// Gate USDT 永续：1 合约 = 1 USDT 名义价值
// Binance 永续：根据不同交易对计算合约张数（这里简化处理，实际会根据交易所调整）
func CalcContracts(positionSizeUSD float64) (int64, error) {
	if positionSizeUSD <= 0 {
		return 0, fmt.Errorf("positionSizeUSD 必须 > 0")
	}

	contracts := int64(math.Floor(positionSizeUSD))
	if contracts < 1 {
		return 0, fmt.Errorf("positionSizeUSD %.2f 太小，无法转换为合约张数", positionSizeUSD)
	}

	return contracts, nil
}

// CalcPartialContracts
func CalcPartialContracts(totalContracts int64, closePercent float64) (int64, error) {
	if totalContracts <= 0 {
		return 0, fmt.Errorf("当前无持仓")
	}
	if closePercent <= 0 || closePercent > 100 {
		return 0, fmt.Errorf("closePercentage 必须在 (0,100]")
	}

	closeContracts := int64(math.Floor(float64(totalContracts) * closePercent / 100))
	if closeContracts < 1 {
		closeContracts = 1
	}
	if closeContracts > totalContracts {
		closeContracts = totalContracts
	}

	return closeContracts, nil
}

// ================================
// 统一下单入口（兼容Binance和Gate.io）
// ================================

// PlaceFuturesOrder 统一下单函数（支持Binance和Gate.io）
// 策略层统一调用此函数，无需关心具体交易所实现
// 参数说明：
// - trader: 交易所实例，必须实现Trader接口
// - req: 统一的下单请求
// - currentPositionContracts: 当前持仓合约张数（平仓/部分平仓用）
func PlaceFuturesOrder(
	trader Trader, // 统一Trader接口，兼容所有交易所
	req FuturesOrderRequest,
	currentPositionContracts int64, // 平仓/部分平仓用
) (map[string]interface{}, error) {

	var (
		size         int64
		err          error
		order        map[string]interface{}
		quantity     float64
		positionSide string
	)

	// 1. 计算合约张数和方向
	switch req.Action {
	case ActionOpenLong:
		// 开多仓
		size, err = CalcContracts(req.PositionSizeUSD)
		if err != nil {
			return nil, err
		}
		size = +size
		quantity = float64(size)
		positionSide = "LONG"

		// 设置杠杆
		_ = trader.SetLeverage(req.Symbol, req.Leverage)

		// 开多仓
		order, err = trader.OpenLong(req.Symbol, quantity, req.Leverage)

	case ActionOpenShort:
		// 开空仓
		size, err = CalcContracts(req.PositionSizeUSD)
		if err != nil {
			return nil, err
		}
		size = -size
		quantity = float64(size)
		positionSide = "SHORT"

		// 设置杠杆
		_ = trader.SetLeverage(req.Symbol, req.Leverage)

		// 开空仓
		order, err = trader.OpenShort(req.Symbol, quantity, req.Leverage)

	case ActionCloseLong:
		// 平多仓
		size = -currentPositionContracts
		quantity = float64(math.Abs(float64(size)))
		positionSide = "LONG"

		// 平多仓
		order, err = trader.CloseLong(req.Symbol, quantity)

	case ActionCloseShort:
		// 平空仓
		size = +currentPositionContracts
		quantity = float64(math.Abs(float64(size)))
		positionSide = "SHORT"

		// 平空仓
		order, err = trader.CloseShort(req.Symbol, quantity)

	case ActionPartial:
		// 部分平仓
		closeContracts, err := CalcPartialContracts(currentPositionContracts, req.ClosePercentage)
		if err != nil {
			return nil, err
		}
		// 计算平仓方向
		if currentPositionContracts > 0 {
			size = -closeContracts
			quantity = float64(closeContracts)
			positionSide = "LONG"

			// 部分平多仓
			order, err = trader.CloseLong(req.Symbol, quantity)
		} else {
			size = +closeContracts
			quantity = float64(closeContracts)
			positionSide = "SHORT"

			// 部分平空仓
			order, err = trader.CloseShort(req.Symbol, quantity)
		}

	default:
		return nil, fmt.Errorf("不支持的 FuturesAction: %s", req.Action)
	}

	if err != nil {
		return nil, err
	}

	log.Printf(
		"📦 下单 | %s | %s | 数量=%.4f | SL=%.4f | TP=%.4f",
		req.Symbol, req.Action, quantity, req.StopLoss, req.TakeProfit,
	)

	// 2. 设置止损止盈
	absContracts := int64(math.Abs(float64(size)))

	if req.StopLoss > 0 {
		if err := trader.SetStopLoss(
			req.Symbol,
			positionSide,
			float64(absContracts),
			req.StopLoss,
		); err != nil {
			log.Printf("⚠ 设置止损失败: %v", err)
		}
	}

	if req.TakeProfit > 0 {
		if err := trader.SetTakeProfit(
			req.Symbol,
			positionSide,
			float64(absContracts),
			req.TakeProfit,
		); err != nil {
			log.Printf("⚠ 设置止盈失败: %v", err)
		}
	}

	return order, nil
}

// BinancePlaceFuturesOrder Binance专用下单函数（兼容旧代码）
// 此函数保持与Binance原有接口兼容，内部调用统一的PlaceFuturesOrder
func BinancePlaceFuturesOrder(
	trader Trader,
	req FuturesOrderRequest,
	currentPositionContracts int64,
) (map[string]interface{}, error) {
	return PlaceFuturesOrder(trader, req, currentPositionContracts)
}

// GatePlaceFuturesOrder Gate.io专用下单函数（兼容旧代码）
// 此函数保持与Gate.io原有接口兼容，内部调用统一的PlaceFuturesOrder
func GatePlaceFuturesOrder(
	trader Trader,
	req FuturesOrderRequest,
	currentPositionContracts int64,
) (map[string]interface{}, error) {
	return PlaceFuturesOrder(trader, req, currentPositionContracts)
}
