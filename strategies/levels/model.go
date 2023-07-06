package levels

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

const atrLength = 14
const trendLine = 400

type Levels struct {
	common.Strategy
	config LevelsConfig

	lowPrice   []float64
	highPrice  []float64
	closePrice []float64
	volume     []float64

	trendBarsCounter int
	prevTrend        string

	highLevel float64
	lowLevel  float64
}

const StrategyName = "levels"
const UpTrend = "UP"
const DownTrend = "Down"
const NoTrend = "Flat"

// NewLevels - creates new Levels crossover strategy
func NewLevels(monitorChannel chan *block.Data, configRaw []byte, keys *users.Keys, historicalData []*block.Data, inst *instance.StrategyInstance) (strategy.Strategy, error) {

	var config LevelsConfig

	err := json.Unmarshal(configRaw, &config)
	if err != nil {
		return nil, err
	}

	acc, err := account.NewBinanceAccount(keys.ApiKey, keys.SecretKey, keys.ApiKey, keys.SecretKey)
	if err != nil {
		return nil, err
	}

	newStrategy := &Levels{
		config: config,
	}
	newStrategy.Account = acc
	newStrategy.StopSignal = make(chan bool)
	newStrategy.MonitorChannel = monitorChannel
	newStrategy.StrategyInstance = inst
	newStrategy.HandlerFunction = newStrategy.HandlerFunc
	newStrategy.DataProcessFunction = newStrategy.ProcessData
	newStrategy.DataLoadEndpoint = newStrategy.LoadData

	newStrategy.closePrice = make([]float64, trendLine)
	newStrategy.volume = make([]float64, trendLine)
	newStrategy.highPrice = make([]float64, trendLine)
	newStrategy.lowPrice = make([]float64, trendLine)

	return newStrategy, nil
}

func (l *Levels) HandlerFunc(marketData *block.Data) {
	if l.closePrice[0] != 0 {
		if l.StrategyInstance.Status == instance.StatusCreated && l.StrategyInstance.Testing == testing.Disable {
			l.StrategyInstance.Status = instance.StatusRunning
			db, _ := database.GetDataBaseConnection()
			_ = instance.UpdateStatus(db, l.StrategyInstance.ID, instance.StatusRunning)
		}

		if l.LastTrade != nil {
			l.TrailingStopLoss(marketData)
			return
		}

		if marketData.ClosePrice > l.highLevel {
			l.HandleBuy(marketData)

			ATR := indicators.AverageTrueRange(indicators.GetSlicedArray(l.highPrice, atrLength), indicators.GetSlicedArray(l.lowPrice, atrLength), indicators.GetSlicedArray(l.closePrice, atrLength), atrLength)
			l.StopLossPrice = marketData.ClosePrice - (2 * ATR[atrLength-1])
		}

		if marketData.ClosePrice < l.lowLevel {
			l.HandleSell(marketData)

			ATR := indicators.AverageTrueRange(indicators.GetSlicedArray(l.highPrice, atrLength), indicators.GetSlicedArray(l.lowPrice, atrLength), indicators.GetSlicedArray(l.closePrice, atrLength), atrLength)
			l.StopLossPrice = marketData.ClosePrice + (2 * ATR[atrLength-1])
		}

		l.GetLevels(marketData)
	}
}

func (l *Levels) GetLevels(marketData *block.Data) {
	low := 999999999999.
	high := 0.

	for i := range l.closePrice {
		if low > l.closePrice[i] {
			low = l.closePrice[i]
		}

		if high < l.closePrice[i] {
			high = l.closePrice[i]
		}
	}
}

func (l *Levels) TrailingStopLoss(marketData *block.Data) {
	index := len(l.closePrice) - 1
	closePrice := marketData.ClosePrice
	prevClosePrice := l.closePrice[index-1]
	if l.LastTrade.FuturesSide == futures.SideTypeBuy {
		if closePrice > prevClosePrice {
			delta := closePrice - prevClosePrice
			l.StopLossPrice = l.StopLossPrice + delta
		}
	} else if l.LastTrade.FuturesSide == futures.SideTypeSell {
		if closePrice < prevClosePrice {
			delta := prevClosePrice - closePrice
			l.StopLossPrice = l.StopLossPrice + delta
		}
	}
}

func (l *Levels) LoadData() error {
	if l.StrategyInstance.Testing == testing.Disable {
		client := futures.NewClient("", "")
		tf := l.config.TimeFrame / 60
		timeMark := "m"
		interval := fmt.Sprintf("%d%s", tf, timeMark)

		res, err := client.NewKlinesService().Symbol(l.config.Pair).Limit(400).Interval(interval).Do(context.Background())
		if err != nil {
			return err
		}

		for i := range res {
			closePrice, _ := strconv.ParseFloat(res[i].Close, 64)
			high, _ := strconv.ParseFloat(res[i].High, 64)
			low, _ := strconv.ParseFloat(res[i].Low, 64)
			md := &block.Data{
				Symbol:     l.config.Pair,
				ClosePrice: closePrice,
				Low:        low,
				High:       high,
			}

			l.ProcessData(md)
		}
	}

	return nil
}

func (l *Levels) GetCurrentTrend(marketData *block.Data) string {
	mean := indicators.Average(l.closePrice)

	if mean < marketData.ClosePrice {
		if l.prevTrend != UpTrend {
			l.prevTrend = UpTrend
			l.trendBarsCounter = 1
		} else {
			l.trendBarsCounter++
		}
	} else {
		if l.prevTrend != DownTrend {
			l.prevTrend = DownTrend
			l.trendBarsCounter = 1
		} else {
			l.trendBarsCounter++
		}
	}

	if l.trendBarsCounter > 6 {
		return l.prevTrend
	} else {
		return NoTrend
	}

}

func (l *Levels) ProcessData(marketData *block.Data) {
	l.volume = l.volume[1:]
	l.volume = append(l.volume, marketData.Volume)

	l.closePrice = l.closePrice[1:]
	l.closePrice = append(l.closePrice, marketData.ClosePrice)

	l.lowPrice = l.lowPrice[1:]
	l.lowPrice = append(l.lowPrice, marketData.Low)

	l.highPrice = l.highPrice[1:]
	l.highPrice = append(l.highPrice, marketData.High)
}

func (l *Levels) GetPotentialProfit(sell bool, targetPrice, currentPrice float64) float64 {
	if sell {
		return (currentPrice / targetPrice) - 1
	} else {
		return (targetPrice / currentPrice) - 1
	}
}
