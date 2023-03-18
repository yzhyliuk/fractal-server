package bollinger_bands_with_atr

import (
	"encoding/json"
	"github.com/adshao/go-binance/v2/futures"
	"newTradingBot/api/database"
	"newTradingBot/indicators"
	"newTradingBot/logs"
	"newTradingBot/models/account"
	"newTradingBot/models/block"
	"newTradingBot/models/strategy"
	"newTradingBot/models/strategy/instance"
	"newTradingBot/models/testing"
	"newTradingBot/models/users"
	"newTradingBot/strategies/common"
)

type BollingerBandsWithATR struct {
	common.Strategy
	config BollingerBandsWithATRConfig

	candles    []*block.Data
	closePrice []float64

	upBreakOut  bool
	lowBreakOut bool

	lastUpperB float64
	lastLowerB float64

	reverseSignal bool

	waiting int
}

const StrategyName = "bollinger_bands_with_atr"

// NewBollingerBandsWithATR - creates new Moving Average crossover strategy
func NewBollingerBandsWithATR(monitorChannel chan *block.Data, configRaw []byte, keys *users.Keys, historicalData []*block.Data, inst *instance.StrategyInstance) (strategy.Strategy, error) {

	var config BollingerBandsWithATRConfig

	err := json.Unmarshal(configRaw, &config)
	if err != nil {
		return nil, err
	}

	acc, err := account.NewBinanceAccount(keys.ApiKey, keys.SecretKey, keys.ApiKey, keys.SecretKey)
	if err != nil {
		return nil, err
	}

	newStrategy := &BollingerBandsWithATR{
		config: config,
	}
	newStrategy.Account = acc
	newStrategy.StopSignal = make(chan bool)
	newStrategy.MonitorChannel = monitorChannel
	newStrategy.StrategyInstance = inst
	newStrategy.HandlerFunction = newStrategy.HandlerFunc
	newStrategy.DataProcessFunction = newStrategy.ProcessData

	newStrategy.candles = make([]*block.Data, config.MALength)
	newStrategy.closePrice = make([]float64, config.MALength)

	return newStrategy, nil
}

func (m *BollingerBandsWithATR) HandlerFunc(marketData *block.Data) {
	if m.candles[0] != nil {
		if m.StrategyInstance.Status == instance.StatusCreated && m.StrategyInstance.Testing == testing.Disable {
			m.StrategyInstance.Status = instance.StatusRunning
			db, _ := database.GetDataBaseConnection()
			_ = instance.UpdateStatus(db, m.StrategyInstance.ID, instance.StatusRunning)
		}

		sd := indicators.StandardDeviation(m.closePrice)
		mean := indicators.Average(m.closePrice)
		up := mean + m.config.BBMultiplier*sd
		low := mean - m.config.BBMultiplier*sd
		rsi := indicators.RSI(m.closePrice, 14)

		m.reverseSignal = false
		m.reverseSignal = ((marketData.ClosePrice < m.lastUpperB) && m.upBreakOut && rsi < 70) || ((marketData.ClosePrice > m.lastLowerB) && m.lowBreakOut && rsi > 30)

		if m.LastTrade != nil {
			if m.waiting > 5 {
				m.waiting = 0
				m.lowBreakOut = false
				m.upBreakOut = false
			} else if m.waiting != 0 {
				m.waiting++
			}

			exitBuy := m.LastTrade.FuturesSide == futures.SideTypeBuy && rsi > 55
			exitSell := m.LastTrade.FuturesSide == futures.SideTypeSell && rsi < 45

			if exitSell || exitBuy {
				m.CloseAllTrades()
			}
		} else {
			m.waiting = 0
		}

		if marketData.ClosePrice > up {
			m.upBreakOut = true
			m.lastUpperB = up
			m.waiting = 1
		} else if marketData.ClosePrice < low {
			m.lowBreakOut = true
			m.lastLowerB = low
			m.waiting = 1
		}

		if m.upBreakOut && m.reverseSignal {
			err := m.HandleSell(marketData)
			if err != nil {
				logs.LogError(err)
			}
			m.upBreakOut = false
			m.TakeProfitPrice = mean
		}

		if m.lowBreakOut && m.reverseSignal {
			err := m.HandleBuy(marketData)
			if err != nil {
				logs.LogError(err)
			}
			m.lowBreakOut = false
			m.TakeProfitPrice = mean
		}

	}
}

func (m *BollingerBandsWithATR) ProcessData(marketData *block.Data) {
	m.candles = m.candles[1:]
	m.candles = append(m.candles, marketData)

	m.closePrice = m.closePrice[1:]
	m.closePrice = append(m.closePrice, marketData.ClosePrice)
}

func (m *BollingerBandsWithATR) CustomTakeProfitAndStopLoss() {

}
