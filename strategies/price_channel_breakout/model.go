package price_channel_breakout

import (
	"encoding/json"
	"fmt"
	"newTradingBot/api/database"
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
	trend                   string

	config PriceChannelBreakoutConfig
}

const (
	UP      = "UP"
	DOWN    = "DOWN"
	NOTREND = "FLAT"
)

const windowLength = 100

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

		// EXECUTION

		volumeBuy := 0.
		volumeSell := 0.

		for idx := range m.volumeObservations {
			if m.priceChangeObservations[idx] > 1 {
				volumeBuy += m.volumeObservations[idx]
			} else {
				volumeSell += m.volumeObservations[idx]
			}
		}

		volumeRatio := volumeBuy / volumeSell

		if volumeRatio > 1.2 {
			//buy
			if m.trend != UP {
				m.trend = UP
				m.PrintData(marketData.ClosePrice, volumeRatio)

				m.CloseAllTrades()
				err := m.HandleBuy(marketData)
				if err != nil {
					logs.LogError(err)
				}
			}
		} else if volumeRatio < 0.8 {
			//sell
			if m.trend != DOWN {
				m.trend = DOWN
				m.PrintData(marketData.ClosePrice, volumeRatio)

				m.CloseAllTrades()
				err := m.HandleSell(marketData)
				if err != nil {
					logs.LogError(err)
				}
			}
		} else {
			conditionOne := m.trend == UP && volumeRatio < 1
			conditionTwo := m.trend == DOWN && volumeRatio > 1
			if conditionOne || conditionTwo {
				m.CloseAllTrades()
				m.trend = NOTREND
				m.PrintData(marketData.ClosePrice, volumeRatio)
			}
		}

		// Print debug data

		fmt.Println("Volume Ratio: ", volumeRatio)
		fmt.Println("Trades Count: ", marketData.TradesCount)
		fmt.Println("Volume: ", marketData.Volume)
		fmt.Println("______________________")

	}
}

func (m *PriceChannelBreakout) ProcessData(marketData *block.Data) {
	m.volumeObservations = m.volumeObservations[1:]
	m.volumeObservations = append(m.volumeObservations, marketData.Volume)

	m.priceChangeObservations = m.priceChangeObservations[1:]
	m.priceChangeObservations = append(m.priceChangeObservations, marketData.ClosePrice/marketData.OpenPrice)
}
func (m *PriceChannelBreakout) PrintData(price, volume float64) {
	fmt.Printf("================== \n")
	fmt.Printf("PRICE: %f TREND: %s \n Volume Ratio: %f", price, m.trend, volume)
}
