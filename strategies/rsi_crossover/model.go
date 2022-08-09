package rsi_crossover

import (
	"fmt"
	"newTradingBot/api/database"
	"newTradingBot/indicators"
	"newTradingBot/logs"
	"newTradingBot/models/account"
	"newTradingBot/models/block"
	"newTradingBot/models/strategy"
	"newTradingBot/models/strategy/instance"
	"newTradingBot/models/users"
	"newTradingBot/strategies/common"
)

type rsiCrossover struct {
	common.Strategy

	closePriceObservations []float64
	rsiObservations []float64

	config RSICrossoverConfig
}

// NewRSICrossoverStrategy - creates new RSI crossover strategy
func NewRSICrossoverStrategy(monitorChannel chan *block.Block, config RSICrossoverConfig, keys *users.Keys, historicalData []*block.Block, inst *instance.StrategyInstance) (strategy.Strategy, error) {

	acc, err := account.NewBinanceAccount(keys.ApiKey,keys.SecretKey, keys.ApiKey, keys.SecretKey)
	if err != nil {
		return nil, err
	}


	newStrategy := &rsiCrossover{
		config: config,
	}
	newStrategy.Account = acc
	newStrategy.StopSignal = make(chan bool)
	newStrategy.MonitorChannel = monitorChannel
	newStrategy.StrategyInstance = inst
	newStrategy.HandlerFunction = newStrategy.HandlerFunc
	newStrategy.DataProcessFunction = newStrategy.ProcessData

	newStrategy.closePriceObservations = make([]float64, config.RSIPeriod)
	newStrategy.rsiObservations = make([]float64, config.LongTermPeriod)

	return newStrategy, nil
}

func (m *rsiCrossover) HandlerFunc(marketData *block.Block)  {
	if m.rsiObservations[0] == 0 {
		return
	}
	if m.StrategyInstance.Status == instance.StatusCreated {
		m.StrategyInstance.Status = instance.StatusRunning
		db, _ := database.GetDataBaseConnection()
		_ = instance.UpdateStatus(db, m.StrategyInstance.ID, instance.StatusRunning)
	}



	longTerm := indicators.SimpleMA(m.rsiObservations,m.config.LongTermPeriod)
	currentRSI := m.rsiObservations[len(m.rsiObservations)-1]

	min := indicators.Min(m.closePriceObservations)
	max := indicators.Max(m.closePriceObservations)
	currentVolatility := ((max/min)-1)*100

	logs.LogDebug(fmt.Sprintf("RSI: (CURRENT): %f \t (LONG): %f \n (VOL): %f%", currentRSI, longTerm, currentVolatility),nil)

	if m.config.Volatility != 0 && m.config.Volatility > currentVolatility {
		m.CloseAllTrades()
		return
	}

	if currentRSI > longTerm {
		err := m.HandleBuy(marketData)
		if err != nil {
			logs.LogDebug("", err)
			return
		}
	} else if currentRSI < longTerm {
		err := m.HandleSell(marketData)
		if err != nil {
			logs.LogDebug("", err)
			return
		}
	}
}

func (m *rsiCrossover) ProcessData(marketData *block.Block)  {
	m.closePriceObservations = m.closePriceObservations[1:]
	m.closePriceObservations = append(m.closePriceObservations, marketData.ClosePrice)

	if m.closePriceObservations[0] != 0 {
		rsi := indicators.RSI(m.closePriceObservations,m.config.RSIPeriod)
		m.rsiObservations = m.rsiObservations[1:]
		m.rsiObservations = append(m.rsiObservations, rsi)
	}
}

