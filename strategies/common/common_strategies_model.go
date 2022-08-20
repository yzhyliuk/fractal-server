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
	"newTradingBot/models/strategy/actions"
	"newTradingBot/models/strategy/instance"
	"newTradingBot/models/testing"
	"newTradingBot/models/trade"
)

type Strategy struct {
	StrategyInstance *instance.StrategyInstance
	Account account.Account
	MonitorChannel chan *block.Data
	StopSignal chan bool
	LastTrade *trade.Trade

	trades []*trade.Trade
	TotalProfit float64

	prevMarketData *block.Data
	Stopped bool
	HandlerFunction func(marketData *block.Data)
	DataProcessFunction func(marketData *block.Data)
	ExperimentalHandler func()
}

func (m *Strategy) Execute()  {
	m.TotalProfit = 0
	m.trades = make([]*trade.Trade, 0)
	go func() {
		for  {
			select {
			case <-m.StopSignal:
				return
			default:
				marketData := <- m.MonitorChannel
				if m.Stopped {
					return
				}

				//marketData = m.ToHeikinAshi(marketData)

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

func (g *Strategy) ToHeikinAshi(marketData *block.Data) *block.Data{
	closeP := (marketData.ClosePrice + marketData.OpenPrice + marketData.High + marketData.Low) / float64(4)
	open := marketData.OpenPrice
	if g.prevMarketData != nil {
		open = (g.prevMarketData.OpenPrice+g.prevMarketData.ClosePrice)/float64(2)
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


func (m *Strategy) HandleTPansSL(marketData *block.Data)  {
	// TODO : TP and SL for spot trading
	if  m.LastTrade == nil || !m.StrategyInstance.IsFutures{
		return
	}

	price := marketData.ClosePrice

	profit := 0.
	roi := 0.

	switch m.LastTrade.FuturesSide {
	case futures.SideTypeBuy:
		profit = (m.LastTrade.Quantity*price)-m.LastTrade.USD
	case futures.SideTypeSell:
		profit = (m.LastTrade.Quantity*m.LastTrade.PriceOpen)-(m.LastTrade.Quantity*price)
	}

	fee := m.LastTrade.USD*trade.BinanceFuturesTakerFeeRate
	profit -= 2*fee

	roi = profit / (m.LastTrade.USD / float64(*m.LastTrade.Leverage))

	if m.StrategyInstance.TradeStopLoss != 0 {
		if roi < m.StrategyInstance.TradeStopLoss*-1 {
			if m.StrategyInstance.Testing == testing.BackTest {
				m.TestingCloseTrade(marketData)
				return
			}
			m.CloseAllTrades()
		}
	}

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

func (m *Strategy) ExecuteExperimental()  {
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

		return true
	}

	return false
}

func (m *Strategy) HandleSell(marketData *block.Data) error {
	if m.StrategyInstance.IsFutures {
		if m.LastTrade == nil {
			return m.sell(marketData)
		}
		if m.LastTrade.FuturesSide != futures.SideTypeSell{
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

		m.LastTrade = futuresTrade
	} else {
		spotTrade, err := m.Account.PlaceMarketOrder(quantity, m.StrategyInstance.Pair,binance.SideTypeBuy, m.StrategyInstance, m.LastTrade)
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

func (m *Strategy) closePreviousTrade()  {
	if m.StrategyInstance.Testing != 0 {
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

func (m *Strategy) Stop()  {
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
	if m.LastTrade != nil && !m.StrategyInstance.IsFutures{
		err := m.HandleSell(nil)
		if err != nil {
			logs.LogDebug("",err)
		}
	} else if m.LastTrade != nil && m.StrategyInstance.IsFutures{
		m.closePreviousTrade()
	}
}

func (m *Strategy) TestingBuy(marketData *block.Data, quantity float64)  {
	newTrade := &trade.Trade{
		Quantity: quantity,
		IsFutures: m.StrategyInstance.IsFutures,
		USD: m.StrategyInstance.Bid,
		PriceOpen: marketData.ClosePrice,
		FuturesSide: futures.SideTypeBuy,
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

func (m *Strategy) TestingSell(marketData *block.Data, quantity float64)  {
	newTrade := &trade.Trade{
		Quantity: quantity,
		USD: m.StrategyInstance.Bid,
		PriceOpen: marketData.ClosePrice,
		IsFutures: m.StrategyInstance.IsFutures,
		FuturesSide: futures.SideTypeSell,
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

func (m *Strategy) TestingCloseTrade(marketData *block.Data)  {
	if m.LastTrade != nil {
		closedTrade := m.LastTrade
		closedTrade.PriceClose = marketData.ClosePrice
		closedTrade.CalculateProfitRoi()
		m.trades = append(m.trades, closedTrade)
		m.LastTrade = nil
	}
}