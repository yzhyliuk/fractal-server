package fibonacci_retrace

import (
	"encoding/json"
	"newTradingBot/api/database"
	"newTradingBot/indicators"
	"newTradingBot/models/account"
	"newTradingBot/models/block"
	"newTradingBot/models/strategy"
	"newTradingBot/models/strategy/instance"
	"newTradingBot/models/testing"
	"newTradingBot/models/users"
	"newTradingBot/strategies/common"
)

type FibonacciRetrace struct {
	common.Strategy
	config FibonacciRetraceConfig

	lowPrice   []float64
	highPrice  []float64
	closePrice []float64

	trendBarsCounter int
	prevTrend        string
}

const StrategyName = "fibonacci_retrace"
const UpTrend = "UP"
const DownTrend = "Down"
const NoTrend = "Flat"

const fibonacciLevel = 0.618

// NewBollingerBandsWithATR - creates new Moving Average crossover strategy
func NewBollingerBandsWithATR(monitorChannel chan *block.Data, configRaw []byte, keys *users.Keys, historicalData []*block.Data, inst *instance.StrategyInstance) (strategy.Strategy, error) {

	var config FibonacciRetraceConfig

	err := json.Unmarshal(configRaw, &config)
	if err != nil {
		return nil, err
	}

	acc, err := account.NewBinanceAccount(keys.ApiKey, keys.SecretKey, keys.ApiKey, keys.SecretKey)
	if err != nil {
		return nil, err
	}

	newStrategy := &FibonacciRetrace{
		config: config,
	}
	newStrategy.Account = acc
	newStrategy.StopSignal = make(chan bool)
	newStrategy.MonitorChannel = monitorChannel
	newStrategy.StrategyInstance = inst
	newStrategy.HandlerFunction = newStrategy.HandlerFunc
	newStrategy.DataProcessFunction = newStrategy.ProcessData

	newStrategy.closePrice = make([]float64, config.MALength)
	newStrategy.highPrice = make([]float64, 20)
	newStrategy.lowPrice = make([]float64, 20)

	return newStrategy, nil
}

func (f *FibonacciRetrace) HandlerFunc(marketData *block.Data) {
	if f.closePrice[0] != 0 {
		if f.StrategyInstance.Status == instance.StatusCreated && f.StrategyInstance.Testing == testing.Disable {
			f.StrategyInstance.Status = instance.StatusRunning
			db, _ := database.GetDataBaseConnection()
			_ = instance.UpdateStatus(db, f.StrategyInstance.ID, instance.StatusRunning)
		}

		trend := f.GetCurrentTrend(marketData)

		bbClosePrices := f.closePrice[f.config.MALength-f.config.BBLength:]
		bbMean := indicators.Average(bbClosePrices)
		bbStDev := indicators.StandardDeviation(bbClosePrices)
		bbUpper := bbMean + f.config.BBMultiplier*bbStDev
		bbLower := bbMean - f.config.BBMultiplier*bbStDev

		priceHigh := indicators.Max(f.highPrice)
		priceLow := indicators.Min(f.lowPrice)

		fibRetraceH := priceHigh - ((priceHigh - priceLow) * f.config.FibonacciLevel)
		fibRetraceL := priceLow + ((priceHigh - priceLow) * f.config.FibonacciLevel)

		if marketData.ClosePrice < bbLower && trend == UpTrend {
			f.HandleBuy(marketData)
			f.TakeProfitPrice = fibRetraceH
		}

		if marketData.ClosePrice > bbUpper && trend == DownTrend {
			f.HandleSell(marketData)
			f.TakeProfitPrice = fibRetraceL
		}
	}
}

func (f *FibonacciRetrace) GetCurrentTrend(marketData *block.Data) string {
	mean := indicators.Average(f.closePrice)

	if mean < marketData.ClosePrice {
		if f.prevTrend != UpTrend {
			f.prevTrend = UpTrend
			f.trendBarsCounter = 1
		} else {
			f.trendBarsCounter++
		}
	} else {
		if f.prevTrend != DownTrend {
			f.prevTrend = DownTrend
			f.trendBarsCounter = 1
		} else {
			f.trendBarsCounter++
		}
	}

	if f.trendBarsCounter > 6 {
		return f.prevTrend
	} else {
		return NoTrend
	}

}

func (f *FibonacciRetrace) ProcessData(marketData *block.Data) {
	f.lowPrice = f.lowPrice[1:]
	f.lowPrice = append(f.lowPrice, marketData.Low)

	f.highPrice = f.highPrice[1:]
	f.highPrice = append(f.highPrice, marketData.High)

	f.closePrice = f.closePrice[1:]
	f.closePrice = append(f.closePrice, marketData.ClosePrice)
}

func (f *FibonacciRetrace) GetPotentialProfit(sell bool, targetPrice, currentPrice float64) float64 {
	if sell {
		return (currentPrice / targetPrice) - 1
	} else {
		return (targetPrice / currentPrice) - 1
	}
}
