package fibonacci_retrace

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	"strconv"
)

type FibonacciRetrace struct {
	common.Strategy
	config FibonacciRetraceConfig

	lowPrice   []float64
	highPrice  []float64
	closePrice []float64
	volume     []float64

	trendBarsCounter int
	prevTrend        string
}

const StrategyName = "fibonacci_retrace"
const UpTrend = "UP"
const DownTrend = "Down"
const NoTrend = "Flat"

// NewFibonacciRetrace - creates new FibonacciRetrace crossover strategy
func NewFibonacciRetrace(monitorChannel chan *block.Data, configRaw []byte, keys *users.Keys, historicalData []*block.Data, inst *instance.StrategyInstance) (strategy.Strategy, error) {

	var config FibonacciRetraceConfig

	err := json.Unmarshal(configRaw, &config)
	if err != nil {
		return nil, err
	}

	// Validation
	if config.MALength <= config.BBLength {
		return nil, errors.New("length of MA parameter should be greater than BB length")
	} else if config.MALength > 500 {
		return nil, errors.New("maximum length of MA parameter is 500")
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
	newStrategy.DataLoadEndpoint = newStrategy.LoadData

	newStrategy.closePrice = make([]float64, config.MALength)
	newStrategy.highPrice = make([]float64, 20)
	newStrategy.lowPrice = make([]float64, 20)
	newStrategy.volume = make([]float64, config.BBLength)

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
			err := f.HandleBuy(marketData)
			if err != nil {
				logs.LogError(err)
			}
			f.TakeProfitPrice = fibRetraceH
		}

		if marketData.ClosePrice > bbUpper && trend == DownTrend {
			err := f.HandleSell(marketData)
			if err != nil {
				logs.LogError(err)
			}
			f.TakeProfitPrice = fibRetraceL
		}
	}
}

func (f *FibonacciRetrace) LoadData() error {
	if f.StrategyInstance.Testing == testing.Disable {
		client := futures.NewClient("", "")
		tf := f.config.TimeFrame / 60
		timeMark := "m"
		interval := fmt.Sprintf("%d%s", tf, timeMark)

		res, err := client.NewKlinesService().Symbol(f.config.Pair).Limit(f.config.MALength).Interval(interval).Do(context.Background())
		if err != nil {
			return err
		}

		for i := range res {
			closePrice, _ := strconv.ParseFloat(res[i].Close, 64)
			high, _ := strconv.ParseFloat(res[i].High, 64)
			low, _ := strconv.ParseFloat(res[i].Low, 64)
			md := &block.Data{
				Symbol:     f.config.Pair,
				ClosePrice: closePrice,
				Low:        low,
				High:       high,
			}

			f.ProcessData(md)
		}
	}

	return nil
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
