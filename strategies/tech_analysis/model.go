package tech_analysis

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/adshao/go-binance/v2/futures"
	"newTradingBot/api/database"
	"newTradingBot/indicators"
	"newTradingBot/models/account"
	"newTradingBot/models/block"
	"newTradingBot/models/strategy"
	"newTradingBot/models/strategy/instance"
	"newTradingBot/models/testing"
	"newTradingBot/models/users"
	"newTradingBot/strategies/common"
	"strconv"
)

const (
	StrategyName = "tech_analysis"
	UpTrend      = "UP"
	DownTrend    = "Down"
	NoTrend      = "Flat"

	levels = 3
)

type AdvancedTechAnalysis struct {
	common.Strategy
	config TechAnalysisConfig

	// base data
	lowPrice   []float64
	highPrice  []float64
	closePrice []float64
	openPrice  []float64
	volume     []float64

	// trend metrics
	trendBarsCounter int
	prevTrend        string

	// levels
	resistanceLevels []float64
	supportLevels    []float64
}

// NewAdvancedTechAnalysis - creates new Moving Average crossover strategy
func NewAdvancedTechAnalysis(monitorChannel chan *block.Data, configRaw []byte, keys *users.Keys, historicalData []*block.Data, inst *instance.StrategyInstance) (strategy.Strategy, error) {

	var config TechAnalysisConfig

	err := json.Unmarshal(configRaw, &config)
	if err != nil {
		return nil, err
	}

	acc, err := account.NewBinanceAccount(keys.ApiKey, keys.SecretKey, keys.ApiKey, keys.SecretKey)
	if err != nil {
		return nil, err
	}

	newStrategy := &AdvancedTechAnalysis{
		config: config,
	}
	newStrategy.Account = acc
	newStrategy.StopSignal = make(chan bool)
	newStrategy.MonitorChannel = monitorChannel
	newStrategy.StrategyInstance = inst
	newStrategy.HandlerFunction = newStrategy.HandlerFunc
	newStrategy.DataProcessFunction = newStrategy.ProcessData
	newStrategy.DataLoadEndpoint = newStrategy.LoadData

	newStrategy.closePrice = make([]float64, config.TrendLength)
	newStrategy.highPrice = make([]float64, config.TrendLength)
	newStrategy.lowPrice = make([]float64, config.TrendLength)
	newStrategy.openPrice = make([]float64, config.TrendLength)
	newStrategy.volume = make([]float64, config.TrendLength)

	newStrategy.resistanceLevels = make([]float64, levels)
	newStrategy.supportLevels = make([]float64, levels)

	return newStrategy, nil
}

func (f *AdvancedTechAnalysis) HandlerFunc(marketData *block.Data) {
	if f.closePrice[0] != 0 {
		if f.StrategyInstance.Status == instance.StatusCreated && f.StrategyInstance.Testing == testing.Disable {
			f.StrategyInstance.Status = instance.StatusRunning
			db, _ := database.GetDataBaseConnection()
			_ = instance.UpdateStatus(db, f.StrategyInstance.ID, instance.StatusRunning)
		}

		//trend := f.GetCurrentTrend(marketData)

		//mean := indicators.Average(indicators.GetSlicedArray(f.closePrice, f.config.BollingerLength))
		//sd := indicators.StandardDeviationWithMean(indicators.GetSlicedArray(f.closePrice, f.config.BollingerLength), mean)
		//upline := mean + f.config.BollingerMultiplier*sd

		lr8, _ := indicators.LinearRegressionForTimeSeries(indicators.GetSlicedArray(f.closePrice, 8))
		lr14, _ := indicators.LinearRegressionForTimeSeries(indicators.GetSlicedArray(f.closePrice, 14))

		if lr8 > lr14 {
			f.HandleBuy(marketData)
		} else if lr8 < lr14 {
			f.HandleSell(marketData)
		}

	}
}

func (f *AdvancedTechAnalysis) LoadData() error {
	if f.StrategyInstance.Testing == testing.Disable {
		client := futures.NewClient("", "")
		tf := f.config.TimeFrame / 60
		timeMark := "m"
		interval := fmt.Sprintf("%d%s", tf, timeMark)

		res, err := client.NewKlinesService().Symbol(f.config.Pair).Limit(f.config.TrendLength).Interval(interval).Do(context.Background())
		if err != nil {
			return err
		}

		for i := range res {
			closePrice, _ := strconv.ParseFloat(res[i].Close, 64)
			high, _ := strconv.ParseFloat(res[i].High, 64)
			low, _ := strconv.ParseFloat(res[i].Low, 64)
			open, _ := strconv.ParseFloat(res[i].Open, 64)
			volume, _ := strconv.ParseFloat(res[i].Volume, 64)
			md := &block.Data{
				Symbol:     f.config.Pair,
				ClosePrice: closePrice,
				Low:        low,
				High:       high,
				OpenPrice:  open,
				Volume:     volume,
			}

			f.ProcessData(md)
		}
	}

	return nil
}

func (f *AdvancedTechAnalysis) GetCurrentTrend(marketData *block.Data) string {
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

func (f *AdvancedTechAnalysis) ProcessData(marketData *block.Data) {
	f.lowPrice = f.lowPrice[1:]
	f.lowPrice = append(f.lowPrice, marketData.Low)

	f.highPrice = f.highPrice[1:]
	f.highPrice = append(f.highPrice, marketData.High)

	f.closePrice = f.closePrice[1:]
	f.closePrice = append(f.closePrice, marketData.ClosePrice)

	f.openPrice = f.openPrice[1:]
	f.openPrice = append(f.openPrice, marketData.OpenPrice)

	f.volume = f.volume[1:]
	f.volume = append(f.volume, marketData.Volume)
}

func (f *AdvancedTechAnalysis) GetPotentialProfit(sell bool, targetPrice, currentPrice float64) float64 {
	if sell {
		return (currentPrice / targetPrice) - 1
	} else {
		return (targetPrice / currentPrice) - 1
	}
}
