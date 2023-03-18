package price_channel_breakout

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

const StrategyName = "price_channel_breakout"

type PriceChannelBreakout struct {
	common.Strategy

	volumeObservations      []float64
	priceChangeObservations []float64

	config PriceChannelBreakoutConfig
}

const windowLength = 60

// NewPriceChannelBreakoutStrategy - creates new Moving Average crossover strategy
func NewPriceChannelBreakoutStrategy(monitorChannel chan *block.Data, configRaw []byte, keys *users.Keys, historicalData []*block.Data, inst *instance.StrategyInstance) (strategy.Strategy, error) {
	var config PriceChannelBreakoutConfig

	err := json.Unmarshal(configRaw, &config)
	if err != nil {
		return nil, err
	}

	var acc account.Account

	if inst.Testing != testing.BackTest {
		acc, err = account.NewBinanceAccount(keys.ApiKey, keys.SecretKey, keys.ApiKey, keys.SecretKey)
		if err != nil {
			return nil, err
		}
	}

	newStrategy := &PriceChannelBreakout{
		config: config,
	}
	newStrategy.Account = acc
	newStrategy.StopSignal = make(chan bool)
	newStrategy.MonitorChannel = monitorChannel
	newStrategy.StrategyInstance = inst
	newStrategy.HandlerFunction = newStrategy.HandlerFunc
	newStrategy.DataProcessFunction = newStrategy.ProcessData

	newStrategy.volumeObservations = make([]float64, windowLength)
	newStrategy.priceChangeObservations = make([]float64, windowLength)

	return newStrategy, nil
}

func (m *PriceChannelBreakout) HandlerFunc(marketData *block.Data) {
	if m.volumeObservations[0] != 0 {
		if m.StrategyInstance.Status == instance.StatusCreated && m.StrategyInstance.Testing == testing.Disable {
			m.StrategyInstance.Status = instance.StatusRunning
			db, _ := database.GetDataBaseConnection()
			_ = instance.UpdateStatus(db, m.StrategyInstance.ID, instance.StatusRunning)
		}

		if m.HandleTrade(marketData) {
			return
		}

		// identify volume extreme values
		volumeSD := indicators.StandardDeviation(m.volumeObservations)
		priceChangeSD := indicators.StandardDeviation(m.priceChangeObservations)

		volumeMean := indicators.Average(m.priceChangeObservations)
		priceChangeMean := indicators.Average(m.priceChangeObservations)

		volumeUpLine := volumeMean + volumeSD*4
		volumeDownLine := volumeMean - volumeSD*4

		priceChangeUpLine := priceChangeMean + priceChangeSD*4
		priceChangeDownLine := priceChangeMean - priceChangeSD*4

		lastPriceChange := m.priceChangeObservations[windowLength-1]

		// Enter Buy
		if marketData.Volume > volumeUpLine && lastPriceChange > priceChangeUpLine {
			err := m.HandleBuy(marketData)
			if err != nil {
				logs.LogError(err)
			}
		}

		// Enter Sell
		if marketData.Volume < volumeDownLine && lastPriceChange < priceChangeDownLine {
			err := m.HandleSell(marketData)
			if err != nil {
				logs.LogError(err)
			}
		}
	}
}

func (m *PriceChannelBreakout) HandleTrade(data *block.Data) bool {
	if m.LastTrade != nil {
		//вольюмДикріс := m.volumeObservations[windowLength-2] < m.volumeObservations[windowLength-1]
		if m.LastTrade.FuturesSide == futures.SideTypeBuy && m.priceChangeObservations[windowLength-1] < 1 {
			m.CloseAllTrades()
		} else if m.LastTrade.FuturesSide == futures.SideTypeSell && m.priceChangeObservations[windowLength-1] > 1 {
			m.CloseAllTrades()
		}
		return true
	}
	return false
}

func (m *PriceChannelBreakout) ProcessData(marketData *block.Data) {
	m.volumeObservations = m.volumeObservations[1:]
	m.volumeObservations = append(m.volumeObservations, marketData.Volume)

	m.priceChangeObservations = m.priceChangeObservations[1:]
	m.priceChangeObservations = append(m.priceChangeObservations, marketData.ClosePrice/marketData.OpenPrice)
}
func (l *PriceChannelBreakout) GetPotentialProfit(sell bool, targetPrice, currentPrice float64) float64 {
	if sell {
		return (currentPrice / targetPrice) - 1
	} else {
		return (targetPrice / currentPrice) - 1
	}
}
