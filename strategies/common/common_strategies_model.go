package common

import (
	"fmt"
	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"newTradingBot/api/database"
	"newTradingBot/indicators"
	"newTradingBot/logs"
	"newTradingBot/models/account"
	"newTradingBot/models/block"
	"newTradingBot/models/notifications"
	"newTradingBot/models/strategy/actions"
	"newTradingBot/models/strategy/instance"
	"newTradingBot/models/testing"
	"newTradingBot/models/trade"
	"strconv"
	"time"
)

const maxPriceDefault = 0.
const lowPriceDefault = 9999999999.

type Strategy struct {
	StrategyInstance *instance.StrategyInstance
	Account          account.Account
	MonitorChannel   chan *block.Data
	StopSignal       chan bool
	LastTrade        *trade.Trade

	trades      []*trade.Trade
	TotalProfit float64

	prevMarketData *block.Data

	currentMarketData   *block.Data
	Stopped             bool
	HandlerFunction     func(marketData *block.Data)
	DataProcessFunction func(marketData *block.Data)
	ExperimentalHandler func()

	TakeProfitPrice float64
	StopLossPrice   float64
}

func (m *Strategy) Execute() {
	m.TotalProfit = 0
	m.trades = make([]*trade.Trade, 0)

	if m.StrategyInstance.Testing != testing.BackTest {
		m.LivePriceMonitoring()
	}

	go func() {
		for {
			select {
			case <-m.StopSignal:
				return
			default:
				marketData := <-m.MonitorChannel
				if m.Stopped {
					return
				}

				m.currentMarketData = marketData

				m.CalculateTradeData(marketData)
				m.HandleStrategyDefinedStopLoss(marketData)

				//marketData = m.ToHeikinAshi(marketData)

				if m.LastTrade != nil {
					m.LastTrade.LengthCounter++
				}

				if m.StopLossCondition() {
					continue
				}
				m.HandleTPansSL(marketData)

				m.DataProcessFunction(marketData)
				m.HandlerFunction(marketData)

				logs.LogDebug(fmt.Sprintf("Data received by instance #%d", m.StrategyInstance.ID), nil)

			}
		}
	}()
}

func (q *Strategy) ChangeBid(bid float64) error {
	q.StrategyInstance.Bid = bid
	db, err := database.GetDataBaseConnection()
	if err != nil {
		return err
	}

	return db.Save(q.StrategyInstance).Error
}

func (g *Strategy) ToHeikinAshi(marketData *block.Data) *block.Data {
	closeP := (marketData.ClosePrice + marketData.OpenPrice + marketData.High + marketData.Low) / float64(4)
	open := marketData.OpenPrice
	if g.prevMarketData != nil {
		open = (g.prevMarketData.OpenPrice + g.prevMarketData.ClosePrice) / float64(2)
	}
	high := indicators.Max([]float64{marketData.High, marketData.OpenPrice, marketData.ClosePrice})
	low := indicators.Min([]float64{marketData.Low, marketData.OpenPrice, marketData.ClosePrice})

	marketData.ClosePrice = closeP
	marketData.OpenPrice = open
	marketData.High = high
	marketData.Low = low

	g.prevMarketData = marketData

	return marketData
}

func (m *Strategy) HandleTPansSL(marketData *block.Data) {
	// TODO : TP and SL for spot trading
	if m.LastTrade == nil || !m.StrategyInstance.IsFutures {
		return
	}

	price := marketData.ClosePrice

	profit := 0.
	roi := 0.

	switch m.LastTrade.FuturesSide {
	case futures.SideTypeBuy:
		profit = (m.LastTrade.Quantity * price) - m.LastTrade.USD
	case futures.SideTypeSell:
		profit = (m.LastTrade.Quantity * m.LastTrade.PriceOpen) - (m.LastTrade.Quantity * price)
	}

	fee := m.LastTrade.USD * trade.BinanceFuturesTakerFeeRate
	profit -= 2 * fee

	roi = profit / (m.LastTrade.USD / float64(*m.LastTrade.Leverage))

	// Trade stop loss
	if m.StrategyInstance.TradeStopLoss != 0 {
		if roi < m.StrategyInstance.TradeStopLoss*-1 {
			if m.StrategyInstance.Testing == testing.BackTest {
				m.TestingCloseTrade(marketData)
				return
			}
			m.CloseAllTrades()
		}
	}

	// trade take profit
	if m.StrategyInstance.TradeTakeProfit != 0 {
		if roi > m.StrategyInstance.TradeTakeProfit {
			if m.StrategyInstance.Testing == testing.BackTest {
				m.TestingCloseTrade(marketData)
				return
			}
			m.CloseAllTrades()
		}
	}
}

func (m *Strategy) ExecuteExperimental() {
	m.ExperimentalHandler()
}

func (m *Strategy) GetInstance() *instance.StrategyInstance {
	return m.StrategyInstance
}

func (m *Strategy) GetTestingTrades() []*trade.Trade {
	return m.trades
}

func (m *Strategy) StopLossCondition() bool {
	if m.StrategyInstance.StopLoss == 0 {
		return false
	}
	if m.TotalProfit*-1 > m.StrategyInstance.StopLoss {
		db, _ := database.GetDataBaseConnection()
		go func() {
			err := actions.StopStrategy(db, m.StrategyInstance)
			if err != nil {
				logs.LogDebug("", err)
			}
		}()

		err := notifications.CreateUserNotification(db, m.StrategyInstance.UserID, notifications.Warning, notifications.StrategyStopLoss(m.StrategyInstance.Pair, m.TotalProfit))
		if err != nil {
			logs.LogError(err)
		}

		return true
	}

	return false
}

func (m *Strategy) HandleSell(marketData *block.Data) error {
	if m.StrategyInstance.IsFutures {
		if m.LastTrade == nil {
			return m.sell(marketData)
		}
		if m.LastTrade.FuturesSide != futures.SideTypeSell {
			return m.sell(marketData)
		}
		return nil
	} else {
		return m.sell(marketData)
	}
}

func (m *Strategy) HandleBuy(marketData *block.Data) error {
	if m.StrategyInstance.IsFutures {
		if m.LastTrade == nil {
			return m.buy(marketData)
		}
		if m.LastTrade.FuturesSide != futures.SideTypeBuy {
			return m.buy(marketData)
		}
		return nil
	} else {
		return m.buy(marketData)
	}
}

// buy - preforms buy
func (m *Strategy) buy(marketData *block.Data) error {
	quantity := account.QuantityFromPrice(m.StrategyInstance.Bid, marketData.ClosePrice)

	if m.StrategyInstance.Testing == testing.BackTest {
		m.TestingBuy(marketData, quantity)
		return nil
	}

	if m.StrategyInstance.IsFutures {
		if m.LastTrade != nil {
			m.closePreviousTrade()
		}
		futuresTrade, err := m.Account.OpenFuturesPosition(quantity, m.StrategyInstance.Pair, futures.SideTypeBuy, m.StrategyInstance)
		if err != nil {
			return err
		}

		m.StopLossPrice = m.CalculateStopLossPrice(futuresTrade.PriceOpen, true)

		m.LastTrade = futuresTrade
	} else {
		spotTrade, err := m.Account.PlaceMarketOrder(quantity, m.StrategyInstance.Pair, binance.SideTypeBuy, m.StrategyInstance, m.LastTrade)
		if err != nil {
			return err
		}

		m.LastTrade = spotTrade
	}

	return nil
}

func (m *Strategy) sell(marketData *block.Data) error {
	quantity := account.QuantityFromPrice(m.StrategyInstance.Bid, marketData.ClosePrice)
	if m.StrategyInstance.Testing == testing.BackTest {
		m.TestingSell(marketData, quantity)
		return nil
	}

	if m.StrategyInstance.IsFutures {
		if m.LastTrade != nil {
			m.closePreviousTrade()
			if marketData == nil {
				return nil
			}
		}
		futuresTrade, err := m.Account.OpenFuturesPosition(quantity, m.StrategyInstance.Pair, futures.SideTypeSell, m.StrategyInstance)
		if err != nil {
			return err
		}

		m.StopLossPrice = m.CalculateStopLossPrice(futuresTrade.PriceOpen, true)

		m.LastTrade = futuresTrade
	} else {
		if m.LastTrade == nil {
			return nil
		}
		_, err := m.Account.PlaceMarketOrder(m.LastTrade.Quantity, m.StrategyInstance.Pair, binance.SideTypeSell, m.StrategyInstance, m.LastTrade)
		if err != nil {
			return err
		}

		m.LastTrade = nil
	}
	return nil
}

func (m *Strategy) closePreviousTrade() {
	if m.StrategyInstance.Testing == testing.BackTest {
		return
	}
	tradeClosed, err := m.Account.CloseFuturesPosition(m.LastTrade)
	if err != nil {
		logs.LogDebug("", err)
		return
	}
	m.TotalProfit += tradeClosed.Profit
	m.LastTrade = nil
}

func (m *Strategy) Stop() {
	db, err := database.GetDataBaseConnection()
	if err != nil {
		logs.LogDebug("", err)
	}
	err = instance.UpdateStatus(db, m.StrategyInstance.ID, instance.StatusStopped)
	if err != nil {
		logs.LogDebug("", err)
	}

	m.Stopped = true
	go func() { m.StopSignal <- true }()

	m.CloseAllTrades()
}

func (m *Strategy) CloseAllTrades() {
	if m.StrategyInstance.Testing == testing.BackTest {
		m.TestingCloseTrade(m.currentMarketData)
	}
	if m.LastTrade != nil && !m.StrategyInstance.IsFutures {
		err := m.HandleSell(nil)
		if err != nil {
			logs.LogDebug("", err)
		}
	} else if m.LastTrade != nil && m.StrategyInstance.IsFutures {
		m.closePreviousTrade()
	}
}

func (m *Strategy) TestingBuy(marketData *block.Data, quantity float64) {
	newTrade := &trade.Trade{
		Quantity:      quantity,
		IsFutures:     m.StrategyInstance.IsFutures,
		USD:           m.StrategyInstance.Bid,
		PriceOpen:     marketData.ClosePrice,
		FuturesSide:   futures.SideTypeBuy,
		LengthCounter: 0,
	}

	if m.StrategyInstance.IsFutures {
		if m.LastTrade != nil && m.LastTrade.FuturesSide == futures.SideTypeBuy {
			return
		}

		if m.LastTrade != nil && m.LastTrade.FuturesSide == futures.SideTypeSell {
			m.TestingCloseTrade(marketData)

			newTrade.Leverage = m.StrategyInstance.Leverage
			m.LastTrade = newTrade
		}
		if m.LastTrade == nil {
			newTrade.Leverage = m.StrategyInstance.Leverage
			m.LastTrade = newTrade
		}
	} else {
		if m.LastTrade == nil {
			m.LastTrade = newTrade
		}
	}
}

func (m *Strategy) TestingSell(marketData *block.Data, quantity float64) {
	newTrade := &trade.Trade{
		Quantity:      quantity,
		USD:           m.StrategyInstance.Bid,
		PriceOpen:     marketData.ClosePrice,
		IsFutures:     m.StrategyInstance.IsFutures,
		FuturesSide:   futures.SideTypeSell,
		LengthCounter: 0,
	}

	if m.StrategyInstance.IsFutures {
		if m.LastTrade != nil && m.LastTrade.FuturesSide == futures.SideTypeSell {
			return
		}

		if m.LastTrade != nil && m.LastTrade.FuturesSide == futures.SideTypeBuy {
			m.TestingCloseTrade(marketData)
		}

		if m.LastTrade == nil {
			newTrade.Leverage = m.StrategyInstance.Leverage
			m.LastTrade = newTrade
		}
	} else {
		m.TestingCloseTrade(marketData)
	}
}

func (m *Strategy) TestingCloseTrade(marketData *block.Data) {
	if m.LastTrade != nil {
		closedTrade := m.LastTrade
		closedTrade.PriceClose = marketData.ClosePrice
		closedTrade.CalculateProfitRoi()
		m.trades = append(m.trades, closedTrade)
		m.LastTrade = nil
	}
}

func (m *Strategy) HandleStrategyDefinedStopLoss(data *block.Data) {
	if m.LastTrade != nil {
		exitTargetUp := data.High
		exitTargetDown := data.Low

		takeProfitSell := false
		takeProfitBuy := false

		stopLossSell := false
		stopLossBuy := false

		if m.TakeProfitPrice != 0 {
			takeProfitSell = m.LastTrade.FuturesSide == futures.SideTypeSell && exitTargetDown < m.TakeProfitPrice
			takeProfitBuy = m.LastTrade.FuturesSide == futures.SideTypeBuy && exitTargetUp > m.TakeProfitPrice
		}

		if m.StopLossPrice != 0 {
			stopLossSell = m.LastTrade.FuturesSide == futures.SideTypeSell && exitTargetUp > m.StopLossPrice
			stopLossBuy = m.LastTrade.FuturesSide == futures.SideTypeBuy && exitTargetDown < m.StopLossPrice
		}

		if takeProfitSell || takeProfitBuy || stopLossSell || stopLossBuy {
			m.CloseAllTrades()
			return
		}
	}
}

func (m *Strategy) CalculateStopLossPrice(priceCurrent float64, sell bool) float64 {
	sl := m.StrategyInstance.TradeStopLoss
	if sl == 0 {
		return 0
	}

	priceDelta := priceCurrent * (sl / float64(*m.StrategyInstance.Leverage))

	if sell {
		return priceCurrent + priceDelta
	} else {
		return priceCurrent - priceDelta
	}
}

func (m *Strategy) CalculateTradeData(marketData *block.Data) {
	if m.LastTrade != nil {
		profit := 0.
		if m.LastTrade.FuturesSide == futures.SideTypeBuy {
			profit = (m.LastTrade.Quantity * marketData.Low) - m.LastTrade.USD
		} else {
			profit = (m.LastTrade.Quantity * m.LastTrade.PriceOpen) - (m.LastTrade.Quantity * marketData.High)

		}

		roi := profit / (m.LastTrade.USD / float64(*m.LastTrade.Leverage))
		if m.LastTrade.MaxDropDown == 0 || m.LastTrade.MaxDropDown > roi {
			m.LastTrade.MaxDropDown = roi
		}

		if m.LastTrade.MaxHeadUp == 0 || m.LastTrade.MaxHeadUp < roi {
			m.LastTrade.MaxHeadUp = roi
		}
	}
}

func (m *Strategy) LivePriceMonitoring() {
	go func() {
		for {
			if m.Stopped {
				break
			} else if m.LastTrade == nil {
				time.Sleep(10 * time.Second)
				continue
			}

			lowPrice := lowPriceDefault
			highPrice := maxPriceDefault

			wsAggTradeHandler := func(event *futures.WsAggTradeEvent) {
				price, _ := strconv.ParseFloat(event.Price, 64)
				if price < lowPrice {
					lowPrice = price
				}
				if price > highPrice {
					highPrice = price
				}
			}

			_, stopC, _ := futures.WsAggTradeServe(m.StrategyInstance.Pair, wsAggTradeHandler, nil)

			time.Sleep(time.Second * 5)

			stopC <- struct{}{}

			shouldClose := false

			if m.LastTrade != nil && (lowPrice != lowPriceDefault || highPrice != maxPriceDefault) {
				if m.StopLossPrice != 0 {
					switch m.LastTrade.FuturesSide {
					case futures.SideTypeSell:
						if highPrice > m.StopLossPrice {
							shouldClose = true
						}
					case futures.SideTypeBuy:
						if lowPrice < m.StopLossPrice {
							shouldClose = true
						}
					}
				}
				if m.TakeProfitPrice != 0 {
					switch m.LastTrade.FuturesSide {
					case futures.SideTypeSell:
						if lowPrice < m.TakeProfitPrice {
							shouldClose = true
						}
					case futures.SideTypeBuy:
						if highPrice > m.TakeProfitPrice {
							shouldClose = true
						}
					}
				}
			}

			if shouldClose {
				m.CloseAllTrades()
			}
		}
	}()
}
