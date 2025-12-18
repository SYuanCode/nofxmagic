package trader

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"nofx/hook"
	"nofx/logger"
	"nofx/market"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/adshao/go-binance/v2/futures"
)

// getBrOrderID 生成唯一订单ID（合约专用）
// 格式: x-{BR_ID}{TIMESTAMP}{RANDOM}
// 合约限制32字符，统一使用此限制以保持一致性
// 使用纳秒时间戳+随机数确保全局唯一性（冲突概率 < 10^-20）
func getBrOrderID() string {
	brID := "KzrpZaP9" // 合约br ID

	// 计算可用空间: 32 - len("x-KzrpZaP9") = 32 - 11 = 21字符
	// 分配: 13位时间戳 + 8位随机数 = 21字符（完美利用）
	timestamp := time.Now().UnixNano() % 10000000000000 // 13位纳秒时间戳

	// 生成4字节随机数（8位十六进制）
	randomBytes := make([]byte, 4)
	rand.Read(randomBytes)
	randomHex := hex.EncodeToString(randomBytes)

	// 格式: x-KzrpZaP9{13位时间戳}{8位随机}
	// 示例: x-KzrpZaP91234567890123abcdef12 (正好31字符)
	orderID := fmt.Sprintf("x-%s%d%s", brID, timestamp, randomHex)

	// 确保不超过32字符限制（理论上正好31字符）
	if len(orderID) > 32 {
		orderID = orderID[:32]
	}

	return orderID
}

// StopLossTakeProfitCondition 止盈止损条件
type StopLossTakeProfitCondition struct {
	Symbol          string  `json:"symbol"`
	PositionSide    string  `json:"position_side"` // "LONG" or "SHORT"
	Quantity        float64 `json:"quantity"`
	StopLossPrice   float64 `json:"stop_loss_price"`
	TakeProfitPrice float64 `json:"take_profit_price"`
	Active          bool    `json:"active"`
}

// FuturesTrader 币安合约交易器
type FuturesTrader struct {
	client *futures.Client

	// 余额缓存
	cachedBalance     map[string]interface{}
	balanceCacheTime  time.Time
	balanceCacheMutex sync.RWMutex

	// 持仓缓存
	cachedPositions     []map[string]interface{}
	positionsCacheTime  time.Time
	positionsCacheMutex sync.RWMutex

	// 缓存有效期（15秒）
	cacheDuration time.Duration

	// 止盈止损条件管理
	slTpConditions     map[string]*StopLossTakeProfitCondition // key: symbol_positionSide (e.g., "BTCUSDT_LONG")
	slTpMutex          sync.RWMutex
	slTpCheckerRunning bool
	slTpCheckerStopCh  chan struct{}
}

// NewFuturesTrader 创建合约交易器
func NewFuturesTrader(apiKey, secretKey string, userId string) *FuturesTrader {
	client := futures.NewClient(apiKey, secretKey)

	hookRes := hook.HookExec[hook.NewBinanceTraderResult](hook.NEW_BINANCE_TRADER, userId, client)
	if hookRes != nil && hookRes.GetResult() != nil {
		client = hookRes.GetResult()
	}

	// 同步时间，避免 Timestamp ahead 错误
	syncBinanceServerTime(client)
	trader := &FuturesTrader{
		client:            client,
		cacheDuration:     5 * time.Second, // 15秒缓存
		slTpConditions:    make(map[string]*StopLossTakeProfitCondition),
		slTpCheckerStopCh: make(chan struct{}),
	}

	// 设置双向持仓模式（Hedge Mode）
	// 这是必需的，因为代码中使用了 PositionSide (LONG/SHORT)
	if err := trader.setDualSidePosition(); err != nil {
		log.Printf("⚠️ 设置双向持仓模式失败: %v (如果已是双向模式则忽略此警告)", err)
	}

	// 启动止盈止损检查器
	trader.startStopLossTakeProfitChecker()

	return trader
}

// setDualSidePosition 设置双向持仓模式（初始化时调用）
func (t *FuturesTrader) setDualSidePosition() error {
	// 尝试设置双向持仓模式
	err := t.client.NewChangePositionModeService().
		DualSide(true). // true = 双向持仓（Hedge Mode）
		Do(context.Background())

	if err != nil {
		// 如果错误信息包含"No need to change"，说明已经是双向持仓模式
		if strings.Contains(err.Error(), "No need to change position side") {
			log.Printf("  ✓ 账户已是双向持仓模式（Hedge Mode）")
			return nil
		}
		// 其他错误则返回（但在调用方不会中断初始化）
		return err
	}

	log.Printf("  ✓ 账户已切换为双向持仓模式（Hedge Mode）")
	log.Printf("  ℹ️  双向持仓模式允许同时持有多单和空单")
	return nil
}

// syncBinanceServerTime 同步币安服务器时间，确保请求时间戳合法
func syncBinanceServerTime(client *futures.Client) {
	serverTime, err := client.NewServerTimeService().Do(context.Background())
	if err != nil {
		log.Printf("⚠️ 同步币安服务器时间失败: %v", err)
		return
	}

	now := time.Now().UnixMilli()
	offset := now - serverTime
	client.TimeOffset = offset
	log.Printf("⏱ 已同步币安服务器时间，偏移 %dms", offset)
}

// GetBalance 获取账户余额（带缓存）
func (t *FuturesTrader) GetBalance() (map[string]interface{}, error) {
	// 先检查缓存是否有效
	t.balanceCacheMutex.RLock()
	if t.cachedBalance != nil && time.Since(t.balanceCacheTime) < t.cacheDuration {
		cacheAge := time.Since(t.balanceCacheTime)
		t.balanceCacheMutex.RUnlock()
		log.Printf("✓ 使用缓存的账户余额（缓存时间: %.1f秒前）", cacheAge.Seconds())
		return t.cachedBalance, nil
	}
	t.balanceCacheMutex.RUnlock()

	// 缓存过期或不存在，调用API
	log.Printf("🔄 缓存过期，正在调用币安API获取账户余额...")
	account, err := t.client.NewGetAccountService().Do(context.Background())
	if err != nil {
		log.Printf("❌ 币安API调用失败: %v", err)
		return nil, fmt.Errorf("获取账户信息失败: %w", err)
	}

	result := make(map[string]interface{})
	result["totalWalletBalance"], _ = strconv.ParseFloat(account.TotalWalletBalance, 64)
	result["availableBalance"], _ = strconv.ParseFloat(account.AvailableBalance, 64)
	result["totalUnrealizedProfit"], _ = strconv.ParseFloat(account.TotalUnrealizedProfit, 64)

	log.Printf("✓ 币安API返回: 总余额=%s, 可用=%s, 未实现盈亏=%s",
		account.TotalWalletBalance,
		account.AvailableBalance,
		account.TotalUnrealizedProfit)
	// 更新缓存
	t.balanceCacheMutex.Lock()
	t.cachedBalance = result
	t.balanceCacheTime = time.Now()
	t.balanceCacheMutex.Unlock()

	return result, nil
}

// GetPositions 获取所有持仓（带缓存）
func (t *FuturesTrader) GetPositions() ([]map[string]interface{}, error) {
	// 先检查缓存是否有效
	t.positionsCacheMutex.RLock()
	if t.cachedPositions != nil && time.Since(t.positionsCacheTime) < t.cacheDuration {
		cacheAge := time.Since(t.positionsCacheTime)
		t.positionsCacheMutex.RUnlock()
		log.Printf("✓ 使用缓存的持仓信息（缓存时间: %.1f秒前）", cacheAge.Seconds())
		return t.cachedPositions, nil
	}
	t.positionsCacheMutex.RUnlock()

	// 缓存过期或不存在，调用API
	log.Printf("🔄 缓存过期，正在调用币安API获取持仓信息...")
	positions, err := t.client.NewGetPositionRiskService().Do(context.Background())
	if err != nil {
		return nil, fmt.Errorf("获取持仓失败: %w", err)
	}

	var result []map[string]interface{}
	for _, pos := range positions {
		posAmt, _ := strconv.ParseFloat(pos.PositionAmt, 64)
		if posAmt == 0 {
			continue // 跳过无持仓的
		}

		posMap := make(map[string]interface{})
		posMap["symbol"] = pos.Symbol
		posMap["positionAmt"], _ = strconv.ParseFloat(pos.PositionAmt, 64)
		posMap["entryPrice"], _ = strconv.ParseFloat(pos.EntryPrice, 64)
		posMap["markPrice"], _ = strconv.ParseFloat(pos.MarkPrice, 64)
		posMap["unRealizedProfit"], _ = strconv.ParseFloat(pos.UnRealizedProfit, 64)
		posMap["leverage"], _ = strconv.ParseFloat(pos.Leverage, 64)
		posMap["liquidationPrice"], _ = strconv.ParseFloat(pos.LiquidationPrice, 64)

		// 判断方向
		if posAmt > 0 {
			posMap["side"] = "long"
		} else {
			posMap["side"] = "short"
		}

		result = append(result, posMap)
	}

	// 更新缓存
	t.positionsCacheMutex.Lock()
	t.cachedPositions = result
	t.positionsCacheTime = time.Now()
	t.positionsCacheMutex.Unlock()

	return result, nil
}

// SetMarginMode 设置仓位模式
func (t *FuturesTrader) SetMarginMode(symbol string, isCrossMargin bool) error {
	var marginType futures.MarginType
	if isCrossMargin {
		marginType = futures.MarginTypeCrossed
	} else {
		marginType = futures.MarginTypeIsolated
	}

	// 尝试设置仓位模式
	err := t.client.NewChangeMarginTypeService().
		Symbol(symbol).
		MarginType(marginType).
		Do(context.Background())

	marginModeStr := "全仓"
	if !isCrossMargin {
		marginModeStr = "逐仓"
	}

	if err != nil {
		// 如果错误信息包含"No need to change"，说明仓位模式已经是目标值
		if contains(err.Error(), "No need to change margin type") {
			log.Printf("  ✓ %s 仓位模式已是 %s", symbol, marginModeStr)
			return nil
		}
		// 如果有持仓，无法更改仓位模式，但不影响交易
		if contains(err.Error(), "Margin type cannot be changed if there exists position") {
			log.Printf("  ⚠️ %s 有持仓，无法更改仓位模式，继续使用当前模式", symbol)
			return nil
		}
		// 检测多资产模式（错误码 -4168）
		if contains(err.Error(), "Multi-Assets mode") || contains(err.Error(), "-4168") || contains(err.Error(), "4168") {
			log.Printf("  ⚠️ %s 检测到多资产模式，强制使用全仓模式", symbol)
			log.Printf("  💡 提示：如需使用逐仓模式，请在币安关闭多资产模式")
			return nil
		}
		// 检测统一账户 API（Portfolio Margin）
		if contains(err.Error(), "unified") || contains(err.Error(), "portfolio") || contains(err.Error(), "Portfolio") {
			log.Printf("  ❌ %s 检测到统一账户 API，无法进行合约交易", symbol)
			return fmt.Errorf("请使用「现货与合约交易」API 权限，不要使用「统一账户 API」")
		}
		log.Printf("  ⚠️ 设置仓位模式失败: %v", err)
		// 不返回错误，让交易继续
		return nil
	}

	log.Printf("  ✓ %s 仓位模式已设置为 %s", symbol, marginModeStr)
	return nil
}

// SetLeverage 设置杠杆（智能判断+冷却期）
func (t *FuturesTrader) SetLeverage(symbol string, leverage int) error {
	// 先尝试获取当前杠杆（从持仓信息）
	currentLeverage := 0
	positions, err := t.GetPositions()
	if err == nil {
		for _, pos := range positions {
			if pos["symbol"] == symbol {
				if lev, ok := pos["leverage"].(float64); ok {
					currentLeverage = int(lev)
					break
				}
			}
		}
	}

	// 如果当前杠杆已经是目标杠杆，跳过
	if currentLeverage == leverage && currentLeverage > 0 {
		log.Printf("  ✓ %s 杠杆已是 %dx，无需切换", symbol, leverage)
		return nil
	}

	// 切换杠杆
	_, err = t.client.NewChangeLeverageService().
		Symbol(symbol).
		Leverage(leverage).
		Do(context.Background())

	if err != nil {
		// 如果错误信息包含"No need to change"，说明杠杆已经是目标值
		if contains(err.Error(), "No need to change") {
			log.Printf("  ✓ %s 杠杆已是 %dx", symbol, leverage)
			return nil
		}
		return fmt.Errorf("设置杠杆失败: %w", err)
	}

	log.Printf("  ✓ %s 杠杆已切换为 %dx", symbol, leverage)

	// 切换杠杆后等待5秒（避免冷却期错误）
	log.Printf("  ⏱ 等待5秒冷却期...")
	time.Sleep(5 * time.Second)

	return nil
}

// OpenLong 开多仓
func (t *FuturesTrader) OpenLong(symbol string, quantity float64, leverage int) (map[string]interface{}, error) {
	// 先取消该币种的所有委托单（清理旧的止损止盈单）
	if err := t.CancelAllOrders(symbol); err != nil {
		log.Printf("  ⚠ 取消旧委托单失败（可能没有委托单）: %v", err)
	}
	if err := t.CancelStopOrders(symbol); err != nil {
		log.Printf("  ⚠ 取消止盈止损单失败（可能没有止盈止损单）: %v", err)
	}
	// 设置杠杆
	if err := t.SetLeverage(symbol, leverage); err != nil {
		return nil, err
	}

	// 注意：仓位模式应该由调用方（AutoTrader）在开仓前通过 SetMarginMode 设置

	// 格式化数量到正确精度
	quantityStr, err := t.FormatQuantity(symbol, quantity)
	if err != nil {
		return nil, err
	}

	// ✅ 检查格式化后的数量是否为 0（防止四舍五入导致的错误）
	quantityFloat, parseErr := strconv.ParseFloat(quantityStr, 64)
	if parseErr != nil || quantityFloat <= 0 {
		return nil, fmt.Errorf("开仓数量过小，格式化后为 0 (原始: %.8f → 格式化: %s)。建议增加开仓金额或选择价格更低的币种", quantity, quantityStr)
	}

	// ✅ 检查最小名义价值（Binance 要求至少 10 USDT）
	if err := t.CheckMinNotional(symbol, quantityFloat); err != nil {
		return nil, err
	}

	// 创建市价买入订单（使用br ID）
	order, err := t.client.NewCreateOrderService().
		Symbol(symbol).
		Side(futures.SideTypeBuy).
		PositionSide(futures.PositionSideTypeLong).
		Type(futures.OrderTypeMarket).
		Quantity(quantityStr).
		NewClientOrderID(getBrOrderID()).
		Do(context.Background())

	if err != nil {
		return nil, fmt.Errorf("开多仓失败: %w", err)
	}

	log.Printf("✓ 开多仓成功: %s 数量: %s", symbol, quantityStr)
	log.Printf("  订单ID: %d", order.OrderID)

	result := make(map[string]interface{})
	result["orderId"] = order.OrderID
	result["symbol"] = order.Symbol
	result["status"] = order.Status

	// 开仓后获取最新市场价格作为实际成交价格
	// 这种方式可以避免依赖交易所返回的订单详情中的具体字段
	marketData, err := market.Get(symbol)
	if err == nil {
		result["price"] = marketData.CurrentPrice
		log.Printf("  实际成交价格: %.4f", marketData.CurrentPrice)
	}

	return result, nil
}

// OpenShort 开空仓
func (t *FuturesTrader) OpenShort(symbol string, quantity float64, leverage int) (map[string]interface{}, error) {
	// 先取消该币种的所有委托单（清理旧的止损止盈单）
	if err := t.CancelAllOrders(symbol); err != nil {
		log.Printf("  ⚠ 取消旧委托单失败（可能没有委托单）: %v", err)
	}
	// 平仓前：取消该币种的所有止盈/止损订单（避免平仓过程中发生意外）
	if err := t.CancelStopOrders(symbol); err != nil {
		log.Printf("  ⚠ 平仓前取消止盈/止损订单失败: %v", err)
		// 继续执行，不中断平仓操作
	}
	// 设置杠杆
	if err := t.SetLeverage(symbol, leverage); err != nil {
		return nil, err
	}

	// 注意：仓位模式应该由调用方（AutoTrader）在开仓前通过 SetMarginMode 设置

	// 格式化数量到正确精度
	quantityStr, err := t.FormatQuantity(symbol, quantity)
	if err != nil {
		return nil, err
	}

	// ✅ 检查格式化后的数量是否为 0（防止四舍五入导致的错误）
	quantityFloat, parseErr := strconv.ParseFloat(quantityStr, 64)
	if parseErr != nil || quantityFloat <= 0 {
		return nil, fmt.Errorf("开仓数量过小，格式化后为 0 (原始: %.8f → 格式化: %s)。建议增加开仓金额或选择价格更低的币种", quantity, quantityStr)
	}

	// ✅ 检查最小名义价值（Binance 要求至少 10 USDT）
	if err := t.CheckMinNotional(symbol, quantityFloat); err != nil {
		return nil, err
	}

	// 创建市价卖出订单（使用br ID）
	order, err := t.client.NewCreateOrderService().
		Symbol(symbol).
		Side(futures.SideTypeSell).
		PositionSide(futures.PositionSideTypeShort).
		Type(futures.OrderTypeMarket).
		Quantity(quantityStr).
		NewClientOrderID(getBrOrderID()).
		Do(context.Background())

	if err != nil {
		return nil, fmt.Errorf("开空仓失败: %w", err)
	}

	log.Printf("✓ 开空仓成功: %s 数量: %s", symbol, quantityStr)
	log.Printf("  订单ID: %d", order.OrderID)

	result := make(map[string]interface{})
	result["orderId"] = order.OrderID
	result["symbol"] = order.Symbol
	result["status"] = order.Status

	// 开仓后获取最新市场价格作为实际成交价格
	// 这种方式可以避免依赖交易所返回的订单详情中的具体字段
	marketData, err := market.Get(symbol)
	if err == nil {
		result["price"] = marketData.CurrentPrice
		log.Printf("  实际成交价格: %.4f", marketData.CurrentPrice)
	}

	return result, nil
}

// CloseLong 平多仓
func (t *FuturesTrader) CloseLong(symbol string, quantity float64) (map[string]interface{}, error) {
	// 如果数量为0，获取当前持仓数量
	if quantity == 0 {
		positions, err := t.GetPositions()
		if err != nil {
			return nil, err
		}

		for _, pos := range positions {
			if pos["symbol"] == symbol && pos["side"] == "long" {
				quantity = pos["positionAmt"].(float64)
				break
			}
		}

		if quantity == 0 {
			return nil, fmt.Errorf("没有找到 %s 的多仓", symbol)
		}
	}

	// 格式化数量
	quantityStr, err := t.FormatQuantity(symbol, quantity)
	if err != nil {
		return nil, err
	}

	// 平仓前：取消该币种的所有止盈/止损订单（避免平仓过程中发生意外）
	if err := t.CancelStopOrders(symbol); err != nil {
		log.Printf("  ⚠ 平仓前取消止盈/止损订单失败: %v", err)
		// 继续执行，不中断平仓操作
	}

	// 创建市价卖出订单（平多，使用br ID）
	order, err := t.client.NewCreateOrderService().
		Symbol(symbol).
		Side(futures.SideTypeSell).
		PositionSide(futures.PositionSideTypeLong).
		Type(futures.OrderTypeMarket).
		Quantity(quantityStr).
		NewClientOrderID(getBrOrderID()).
		Do(context.Background())

	if err != nil {
		return nil, fmt.Errorf("平多仓失败: %w", err)
	}

	log.Printf("✓ 平多仓成功: %s 数量: %s", symbol, quantityStr)

	// 平仓后取消该币种的所有挂单（止损止盈单）
	if err := t.CancelAllOrders(symbol); err != nil {
		log.Printf("  ⚠ 取消挂单失败: %v", err)
	}

	// 获取当前价格
	price, _ := t.GetMarketPrice(symbol)

	// 发送Telegram通知
	tgMessage := fmt.Sprintf("🔄 **平多仓成功**\n"+
		"📋 币种: `%s`\n"+
		"📊 平仓价格: `%.4f`\n"+
		"📈 数量: `%.4f`\n"+
		"📝 订单ID: `%d`\n"+
		"⏰ 时间: `%s`",
		symbol,
		price,
		quantity,
		order.OrderID,
		time.Now().Format("2006-01-02 15:04:05"))
	logger.SendTelegramMessage(tgMessage)

	result := make(map[string]interface{})
	result["orderId"] = order.OrderID
	result["symbol"] = order.Symbol
	result["status"] = order.Status
	return result, nil
}

// CloseShort 平空仓
func (t *FuturesTrader) CloseShort(symbol string, quantity float64) (map[string]interface{}, error) {
	// 如果数量为0，获取当前持仓数量
	if quantity == 0 {
		positions, err := t.GetPositions()
		if err != nil {
			return nil, err
		}

		for _, pos := range positions {
			if pos["symbol"] == symbol && pos["side"] == "short" {
				quantity = -pos["positionAmt"].(float64) // 空仓数量是负的，取绝对值
				break
			}
		}

		if quantity == 0 {
			return nil, fmt.Errorf("没有找到 %s 的空仓", symbol)
		}
	}

	// 格式化数量
	quantityStr, err := t.FormatQuantity(symbol, quantity)
	if err != nil {
		return nil, err
	}

	// 平仓前：取消该币种的所有止盈/止损订单（避免平仓过程中发生意外）
	if err := t.CancelStopOrders(symbol); err != nil {
		log.Printf("  ⚠ 平仓前取消止盈/止损订单失败: %v", err)
		// 继续执行，不中断平仓操作
	}

	// 创建市价买入订单（平空，使用br ID）
	order, err := t.client.NewCreateOrderService().
		Symbol(symbol).
		Side(futures.SideTypeBuy).
		PositionSide(futures.PositionSideTypeShort).
		Type(futures.OrderTypeMarket).
		Quantity(quantityStr).
		NewClientOrderID(getBrOrderID()).
		Do(context.Background())

	if err != nil {
		return nil, fmt.Errorf("平空仓失败: %w", err)
	}

	log.Printf("✓ 平空仓成功: %s 数量: %s", symbol, quantityStr)

	// 平仓后取消该币种的所有挂单（止损止盈单）
	if err := t.CancelAllOrders(symbol); err != nil {
		log.Printf("  ⚠ 取消挂单失败: %v", err)
	}

	// 获取当前价格
	price, _ := t.GetMarketPrice(symbol)

	// 发送Telegram通知
	tgMessage := fmt.Sprintf("🔄 **平空仓成功**\n"+
		"📋 币种: `%s`\n"+
		"📊 平仓价格: `%.4f`\n"+
		"📉 数量: `%.4f`\n"+
		"📝 订单ID: `%d`\n"+
		"⏰ 时间: `%s`",
		symbol,
		price,
		quantity,
		order.OrderID,
		time.Now().Format("2006-01-02 15:04:05"))
	logger.SendTelegramMessage(tgMessage)

	result := make(map[string]interface{})
	result["orderId"] = order.OrderID
	result["symbol"] = order.Symbol
	result["status"] = order.Status
	return result, nil
}

// CancelStopLossOrders 仅取消止损单（不影响止盈单）
func (t *FuturesTrader) CancelStopLossOrders(symbol string) error {
	// 取消本地止损条件
	canceledCount := 0

	t.slTpMutex.Lock()
	for key, cond := range t.slTpConditions {
		if strings.HasPrefix(key, symbol+"_") {
			// 仅重置止损价格，不取消止盈
			if cond.StopLossPrice > 0 {
				cond.StopLossPrice = 0
				canceledCount++
				log.Printf("  ✓ 已取消 %s %s 的本地止损单", cond.Symbol, cond.PositionSide)
			}
		}
	}
	t.slTpMutex.Unlock()

	if canceledCount == 0 {
		log.Printf("  ℹ %s 没有止损单需要取消", symbol)
	} else {
		log.Printf("  ✓ 已取消 %s 的 %d 个本地止损单", symbol, canceledCount)
	}

	return nil
}

// CancelTakeProfitOrders 仅取消止盈单（不影响止损单）
func (t *FuturesTrader) CancelTakeProfitOrders(symbol string) error {
	// 取消本地止盈条件
	canceledCount := 0

	t.slTpMutex.Lock()
	for key, cond := range t.slTpConditions {
		if strings.HasPrefix(key, symbol+"_") {
			// 仅重置止盈价格，不取消止损
			if cond.TakeProfitPrice > 0 {
				cond.TakeProfitPrice = 0
				canceledCount++
				log.Printf("  ✓ 已取消 %s %s 的本地止盈单", cond.Symbol, cond.PositionSide)
			}
		}
	}
	t.slTpMutex.Unlock()

	if canceledCount == 0 {
		log.Printf("  ℹ %s 没有止盈单需要取消", symbol)
	} else {
		log.Printf("  ✓ 已取消 %s 的 %d 个本地止盈单", symbol, canceledCount)
	}

	return nil
}

// CancelAllOrders 取消该币种的所有挂单
func (t *FuturesTrader) CancelAllOrders(symbol string) error {
	err := t.client.NewCancelAllOpenOrdersService().
		Symbol(symbol).
		Do(context.Background())

	if err != nil {
		return fmt.Errorf("取消挂单失败: %w", err)
	}

	log.Printf("  ✓ 已取消 %s 的所有挂单", symbol)
	return nil
}

// CancelStopOrders 取消该币种的止盈/止损单（用于调整止盈止损位置）
func (t *FuturesTrader) CancelStopOrders(symbol string) error {
	// 取消本地止盈止损条件
	canceledCount := 0

	t.slTpMutex.Lock()
	for key, cond := range t.slTpConditions {
		if strings.HasPrefix(key, symbol+"_") {
			// 重置止盈止损价格
			cond.StopLossPrice = 0
			cond.TakeProfitPrice = 0
			cond.Active = false
			canceledCount++
			log.Printf("  ✓ 已取消 %s %s 的本地止盈止损单", cond.Symbol, cond.PositionSide)
		}
	}
	t.slTpMutex.Unlock()

	if canceledCount == 0 {
		log.Printf("  ℹ %s 没有止盈止损单需要取消", symbol)
	} else {
		log.Printf("  ✓ 已取消 %s 的 %d 个本地止盈止损单", symbol, canceledCount)
	}

	return nil
}

// GetMarketPrice 获取市场价格
func (t *FuturesTrader) GetMarketPrice(symbol string) (float64, error) {
	prices, err := t.client.NewListPricesService().Symbol(symbol).Do(context.Background())
	if err != nil {
		return 0, fmt.Errorf("获取价格失败: %w", err)
	}

	if len(prices) == 0 {
		return 0, fmt.Errorf("未找到价格")
	}

	price, err := strconv.ParseFloat(prices[0].Price, 64)
	if err != nil {
		return 0, err
	}

	return price, nil
}

// CalculatePositionSize 计算仓位大小
func (t *FuturesTrader) CalculatePositionSize(balance, riskPercent, price float64, leverage int) float64 {
	riskAmount := balance * (riskPercent / 100.0)
	positionValue := riskAmount * float64(leverage)
	quantity := positionValue / price
	return quantity
}

// SetStopLoss 设置止损单
// 使用本地维护的止盈止损条件来实现
func (t *FuturesTrader) SetStopLoss(symbol string, positionSide string, quantity, stopPrice float64) error {
	// 生成唯一键
	key := symbol + "_" + positionSide

	// 存储或更新本地止盈止损条件
	t.slTpMutex.Lock()
	cond, exists := t.slTpConditions[key]
	if !exists {
		// 创建新的条件
		cond = &StopLossTakeProfitCondition{
			Symbol:       symbol,
			PositionSide: positionSide,
			Quantity:     quantity,
			Active:       true,
		}
		t.slTpConditions[key] = cond
	}
	// 更新止损价格
	cond.StopLossPrice = stopPrice
	t.slTpMutex.Unlock()

	logger.Infof("✅ 本地止损已设置: %s %s, 数量: %.4f, 止损价格: %.4f",
		symbol, positionSide, quantity, stopPrice)
	return nil
}

// SetTakeProfit 设置止盈单
// 使用本地维护的止盈止损条件来实现
func (t *FuturesTrader) SetTakeProfit(symbol string, positionSide string, quantity, takeProfitPrice float64) error {
	// 生成唯一键
	key := symbol + "_" + positionSide

	// 存储或更新本地止盈止损条件
	t.slTpMutex.Lock()
	cond, exists := t.slTpConditions[key]
	if !exists {
		// 创建新的条件
		cond = &StopLossTakeProfitCondition{
			Symbol:       symbol,
			PositionSide: positionSide,
			Quantity:     quantity,
			Active:       true,
		}
		t.slTpConditions[key] = cond
	}
	// 更新止盈价格
	cond.TakeProfitPrice = takeProfitPrice
	t.slTpMutex.Unlock()

	logger.Infof("✅ 本地止盈已设置: %s %s, 数量: %.4f, 止盈价格: %.4f",
		symbol, positionSide, quantity, takeProfitPrice)
	return nil
}

// startStopLossTakeProfitChecker 启动止盈止损检查器
func (t *FuturesTrader) startStopLossTakeProfitChecker() {
	if t.slTpCheckerRunning {
		return // 检查器已经在运行
	}

	t.slTpCheckerRunning = true
	ticker := time.NewTicker(2 * time.Second) // 每两秒检查一次

	go func() {
		for {
			select {
			case <-ticker.C:
				t.checkStopLossTakeProfit()
			case <-t.slTpCheckerStopCh:
				ticker.Stop()
				t.slTpCheckerRunning = false
				return
			}
		}
	}()

	log.Println("✅ 本地止盈止损检查器已启动，每2秒检查一次")
}

// checkStopLossTakeProfit 检查止盈止损条件
func (t *FuturesTrader) checkStopLossTakeProfit() {
	// 获取所有活跃的止盈止损条件
	t.slTpMutex.RLock()
	conditions := make([]*StopLossTakeProfitCondition, 0, len(t.slTpConditions))
	for _, cond := range t.slTpConditions {
		if cond.Active {
			conditions = append(conditions, cond)
		}
	}
	t.slTpMutex.RUnlock()

	if len(conditions) == 0 {
		return // 没有活跃的条件，直接返回
	}

	// 获取所有持仓信息
	positions, err := t.GetPositions()
	if err != nil {
		log.Printf("⚠️ 获取持仓信息失败: %v", err)
		return
	}

	// 将持仓信息按 symbol_positionSide 分组
	positionMap := make(map[string]map[string]interface{})
	for _, pos := range positions {
		symbol := pos["symbol"].(string)
		side := pos["side"].(string)
		if side == "SELL" {
			side = "SHORT"
		} else if side == "BUY" {
			side = "LONG"
		}
		key := symbol + "_" + side
		positionMap[key] = pos
	}

	// 遍历检查每个条件
	for _, cond := range conditions {
		// 获取当前市场价格
		price, err := t.GetMarketPrice(cond.Symbol)
		if err != nil {
			log.Printf("⚠️ 获取 %s 价格失败: %v", cond.Symbol, err)
			continue
		}

		// 构建持仓key
		posKey := cond.Symbol + "_" + cond.PositionSide

		// 获取对应的持仓信息
		pos, hasPosition := positionMap[posKey]
		if !hasPosition {
			// 没有持仓，跳过检查
			continue
		}

		// 获取持仓信息中的相关字段
		var unrealizedPnl, marginUsed float64
		var leverage int

		// 从持仓信息中提取unrealizedPnl
		if upnl, ok := pos["unRealizedProfit"].(float64); ok {
			unrealizedPnl = upnl
		}

		// 从持仓信息中提取杠杆
		leverage = 10 // 默认值
		if lev, ok := pos["leverage"].(float64); ok {
			leverage = int(lev)
		} else if levStr, ok := pos["leverage"].(string); ok {
			if lev, err := strconv.Atoi(levStr); err == nil {
				leverage = lev
			}
		}

		// 计算占用保证金
		markPrice := 0.0
		if mp, ok := pos["markPrice"].(float64); ok {
			markPrice = mp
		} else {
			markPrice = price //  fallback to current price if mark price not available
		}

		quantity := 0.0
		if qty, ok := pos["positionAmt"].(float64); ok {
			quantity = qty
			if quantity < 0 {
				quantity = -quantity // 空仓数量为负，转为正数
			}
		}

		if quantity > 0 {
			marginUsed = (quantity * markPrice) / float64(leverage)
		}

		// 计算盈亏百分比
		pnlPct := calculatePnLPercentage(unrealizedPnl, marginUsed)

		// 检查是否满足杠杆对应的盈亏区间条件
		shouldTrigger := true
		if leverage < 50 {
			// 50倍以下杠杆：盈亏率在-10%~15%不触发
			if pnlPct >= -10 && pnlPct <= 15 {
				shouldTrigger = false
				log.Printf("📊 跳过止盈止损: %s %s, 杠杆: %d倍, 盈亏率: %.2f%% (在-10%%~15%%区间内)",
					cond.Symbol, cond.PositionSide, leverage, pnlPct)
			}
		} else {
			// 50倍以上杠杆：盈亏率在-20%~30%不触发
			if pnlPct >= -20 && pnlPct <= 30 {
				shouldTrigger = false
				log.Printf("📊 跳过止盈止损: %s %s, 杠杆: %d倍, 盈亏率: %.2f%% (在-20%%~30%%区间内)",
					cond.Symbol, cond.PositionSide, leverage, pnlPct)
			}
		}

		// 检查止损条件
		if cond.StopLossPrice > 0 && shouldTrigger {
			if (cond.PositionSide == "LONG" && price <= cond.StopLossPrice) ||
				(cond.PositionSide == "SHORT" && price >= cond.StopLossPrice) {
				// 触发止损
				log.Printf("🚨 触发止损: %s %s, 当前价格: %.4f, 止损价格: %.4f, 杠杆: %d倍, 盈亏率: %.2f%%",
					cond.Symbol, cond.PositionSide, price, cond.StopLossPrice, leverage, pnlPct)
				t.executeStopLoss(cond)
			}
		}

		// 检查止盈条件
		if cond.TakeProfitPrice > 0 && shouldTrigger {
			if (cond.PositionSide == "LONG" && price >= cond.TakeProfitPrice) ||
				(cond.PositionSide == "SHORT" && price <= cond.TakeProfitPrice) {
				// 触发止盈
				log.Printf("🚨 触发止盈: %s %s, 当前价格: %.4f, 止盈价格: %.4f, 杠杆: %d倍, 盈亏率: %.2f%%",
					cond.Symbol, cond.PositionSide, price, cond.TakeProfitPrice, leverage, pnlPct)
				t.executeTakeProfit(cond)
			}
		}
	}
}

// executeStopLoss 执行止损操作
func (t *FuturesTrader) executeStopLoss(cond *StopLossTakeProfitCondition) {
	// 执行平仓操作
	var err error
	if cond.PositionSide == "LONG" {
		_, err = t.CloseLong(cond.Symbol, cond.Quantity)
	} else {
		_, err = t.CloseShort(cond.Symbol, cond.Quantity)
	}

	if err != nil {
		log.Printf("❌ 执行止损失败: %v", err)
		return
	}

	// 停止止损止盈条件
	t.slTpMutex.Lock()
	cond.Active = false
	t.slTpMutex.Unlock()

	log.Printf("✅ 止损执行成功: %s %s, 平仓数量: %.4f", cond.Symbol, cond.PositionSide, cond.Quantity)
}

// executeTakeProfit 执行止盈操作
func (t *FuturesTrader) executeTakeProfit(cond *StopLossTakeProfitCondition) {
	// 执行平仓操作
	var err error
	if cond.PositionSide == "LONG" {
		_, err = t.CloseLong(cond.Symbol, cond.Quantity)
	} else {
		_, err = t.CloseShort(cond.Symbol, cond.Quantity)
	}

	if err != nil {
		log.Printf("❌ 执行止盈失败: %v", err)
		return
	}

	// 停止止损止盈条件
	t.slTpMutex.Lock()
	cond.Active = false
	t.slTpMutex.Unlock()

	log.Printf("✅ 止盈执行成功: %s %s, 平仓数量: %.4f", cond.Symbol, cond.PositionSide, cond.Quantity)
}

// GetMinNotional 获取最小名义价值（Binance要求）
func (t *FuturesTrader) GetMinNotional(symbol string) float64 {
	// 使用保守的默认值 10 USDT，确保订单能够通过交易所验证
	return 10.0
}

// CheckMinNotional 检查订单是否满足最小名义价值要求
func (t *FuturesTrader) CheckMinNotional(symbol string, quantity float64) error {
	price, err := t.GetMarketPrice(symbol)
	if err != nil {
		return fmt.Errorf("获取市价失败: %w", err)
	}

	notionalValue := quantity * price
	minNotional := t.GetMinNotional(symbol)

	if notionalValue < minNotional {
		return fmt.Errorf(
			"订单金额 %.2f USDT 低于最小要求 %.2f USDT (数量: %.4f, 价格: %.4f)",
			notionalValue, minNotional, quantity, price,
		)
	}

	return nil
}

// GetSymbolPrecision 获取交易对的数量精度
func (t *FuturesTrader) GetSymbolPrecision(symbol string) (int, error) {
	exchangeInfo, err := t.client.NewExchangeInfoService().Do(context.Background())
	if err != nil {
		return 0, fmt.Errorf("获取交易规则失败: %w", err)
	}

	for _, s := range exchangeInfo.Symbols {
		if s.Symbol == symbol {
			// 从LOT_SIZE filter获取精度
			for _, filter := range s.Filters {
				if filter["filterType"] == "LOT_SIZE" {
					stepSize := filter["stepSize"].(string)
					precision := calculatePrecision(stepSize)
					log.Printf("  %s 数量精度: %d (stepSize: %s)", symbol, precision, stepSize)
					return precision, nil
				}
			}
		}
	}

	log.Printf("  ⚠ %s 未找到精度信息，使用默认精度3", symbol)
	return 3, nil // 默认精度为3
}

// calculatePrecision 从stepSize计算精度
func calculatePrecision(stepSize string) int {
	// 去除尾部的0
	stepSize = trimTrailingZeros(stepSize)

	// 查找小数点
	dotIndex := -1
	for i := 0; i < len(stepSize); i++ {
		if stepSize[i] == '.' {
			dotIndex = i
			break
		}
	}

	// 如果没有小数点或小数点在最后，精度为0
	if dotIndex == -1 || dotIndex == len(stepSize)-1 {
		return 0
	}

	// 返回小数点后的位数
	return len(stepSize) - dotIndex - 1
}

// trimTrailingZeros 去除尾部的0
func trimTrailingZeros(s string) string {
	// 如果没有小数点，直接返回
	if !stringContains(s, ".") {
		return s
	}

	// 从后向前遍历，去除尾部的0
	for len(s) > 0 && s[len(s)-1] == '0' {
		s = s[:len(s)-1]
	}

	// 如果最后一位是小数点，也去掉
	if len(s) > 0 && s[len(s)-1] == '.' {
		s = s[:len(s)-1]
	}

	return s
}

// FormatQuantity 格式化数量到正确的精度
func (t *FuturesTrader) FormatQuantity(symbol string, quantity float64) (string, error) {
	precision, err := t.GetSymbolPrecision(symbol)
	if err != nil {
		// 如果获取失败，使用默认格式
		return fmt.Sprintf("%.3f", quantity), nil
	}

	format := fmt.Sprintf("%%.%df", precision)
	return fmt.Sprintf(format, quantity), nil
}

// GetTradeHistory 获取交易历史记录
func (t *FuturesTrader) GetTradeHistory(symbol string, limit int) ([]map[string]interface{}, error) {
	if limit <= 0 {
		limit = 500 // 默认获取500条
	}

	// TODO: 修复Binance API调用，当前库版本中NewGetAccountTradesService方法不存在
	// 暂时返回空列表，后续修复正确的API调用
	return []map[string]interface{}{}, nil
}

// 辅助函数
func contains(s, substr string) bool {
	return len(s) >= len(substr) && stringContains(s, substr)
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
