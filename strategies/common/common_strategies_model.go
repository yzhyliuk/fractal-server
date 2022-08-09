package common

import (
	"fmt"
	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"newTradingBot/api/database"
	"newTradingBot/logs"
	"newTradingBot/models/account"
	"newTradingBot/models/block"
	"newTradingBot/models/strategy/actions"
	"newTradingBot/models/strategy/instance"
	"newTradingBot/models/trade"
)

type Strategy struct {
	StrategyInstance *instance.StrategyInstance
	Account account.Account
	MonitorChannel chan *block.Block
	StopSignal chan bool
	LastTrade *trade.Trade

	TotalProfit float64

	Stopped bool
	HandlerFunction func(marketData *block.Block)
	DataProcessFunction func(marketData *block.Block)
	ExperimentalHandler func()
}

func (m *Strategy) Execute()  {
	m.TotalProfit = 0
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

				if m.StopLossCondition() {
					continue
				}

				m.DataProcessFunction(marketData)
				m.HandlerFunction(marketData)

				logs.LogDebug(fmt.Sprintf("Data received by instance #%d", m.StrategyInstance.ID), nil)

			}
		}
	}()
}

func (m *Strategy) ExecuteExperimental()  {
	m.ExperimentalHandler()
}

func (m *Strategy) GetInstance() *instance.StrategyInstance {
	return m.StrategyInstance
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

func (m *Strategy) HandleSell(marketData *block.Block) error {
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



func (m *Strategy) HandleBuy(marketData *block.Block) error {
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
func (m *Strategy) buy(marketData *block.Block) error {
	quantity := account.QuantityFromPrice(m.StrategyInstance.Bid, marketData.ClosePrice)

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

func (m *Strategy) sell(marketData *block.Block) error {
	if m.StrategyInstance.IsFutures {
		if m.LastTrade != nil {
			m.closePreviousTrade()
			if marketData == nil {
				return nil
			}
		}
		quantity := account.QuantityFromPrice(m.StrategyInstance.Bid, marketData.ClosePrice)
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

