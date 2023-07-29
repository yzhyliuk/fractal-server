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

type Levels struct {
	common.Strategy
	config LevelsConfig

	lowPrice   []float64
	highPrice  []float64
	closePrice []float64
	openPrice  []float64
	volume     []float64

	rsi []*float64

	deltaPrice []float64

	trendBarsCounter int
	prevTrend        string

	prevEMA float64

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

	newStrategy.closePrice = make([]float64, config.EMA)
	newStrategy.volume = make([]float64, config.EMA)
	newStrategy.highPrice = make([]float64, config.EMA)
	newStrategy.lowPrice = make([]float64, config.EMA)
	newStrategy.openPrice = make([]float64, config.EMA)
	newStrategy.deltaPrice = make([]float64, config.EMA)
	newStrategy.rsi = make([]*float64, 3)

	return newStrategy, nil
}

func (l *Levels) HandlerFunc(marketData *block.Data) {
	if l.closePrice[0] != 0 {
		if l.StrategyInstance.Status == instance.StatusCreated && l.StrategyInstance.Testing == testing.Disable {
			l.StrategyInstance.Status = instance.StatusRunning
			db, _ := database.GetDataBaseConnection()
			_ = instance.UpdateStatus(db, l.StrategyInstance.ID, instance.StatusRunning)
		}

		//if l.prevEMA == 0 {
		//	l.prevEMA = indicators.Average(l.closePrice)
		//	return
		//}
		//
		//ema := indicators.ExponentialMA(l.config.EMA, l.prevEMA, indicators.Average(l.closePrice))
		//l.prevEMA = ema

		ema := indicators.Average(l.closePrice)

		mom := marketData.ClosePrice - l.closePrice[len(l.closePrice)-1-l.config.MomentumLength]

		// Pay attention to this line while debug
		crossUp := mom > 0
		crossDown := mom < 0

		// calculate RMA
		u0 := marketData.ClosePrice
		u1 := l.closePrice[len(l.closePrice)-2]
		if u1 > u0 {
			u0 = 0
			u1 = 0
		}
		up := indicators.RMA(u0, u1, l.config.RSILength)

		u0 = marketData.ClosePrice
		u1 = l.closePrice[len(l.closePrice)-2]
		if u1 < u0 {
			u0 = 0
			u1 = 0
		} else {
			u1 = marketData.ClosePrice
			u0 = l.closePrice[len(l.closePrice)-2]
		}
		// Pay attention here maybe need absolute values
		down := indicators.RMA(u0, u1, l.config.RSILength)
		rsi := getRSI(up, down)
		l.SaveRSI(rsi)
		if l.rsi[0] == nil {
			return
		}

		oversoldAgo := CheckOversoldAgo(l.rsi, float64(l.config.RSIOversold))
		overboughtAgo := CheckOverboughtAgo(l.rsi, float64(l.config.RSIOverbought))

		bullishDivergenceCondition := CheckBullishDivergence(l.rsi)
		bearishDivergenceCondition := CheckBearishDivergence(l.rsi)

		// Entry Conditions
		longEntryCondition := crossUp && oversoldAgo && bullishDivergenceCondition
		shortEntryCondition := crossDown && overboughtAgo && bearishDivergenceCondition

		longCondition := longEntryCondition && marketData.ClosePrice <= ema
		// Sell if short condition is met and price has pulled back to or above the 100 EMA
		shortCondition := shortEntryCondition && marketData.ClosePrice >= ema

		if longCondition {
			l.HandleBuy(marketData)
			l.SetStopLossPrice(marketData)
			l.SetTakeProfit(marketData)
		}
		if shortCondition {
			l.HandleSell(marketData)
			l.SetStopLossPrice(marketData)
			l.SetTakeProfit(marketData)
		}

	}
}

func (l *Levels) SaveRSI(rsi float64) {
	l.rsi = l.rsi[1:]
	l.rsi = append(l.rsi, &rsi)
}

// CheckOversoldAgo checks if any of the last four RSI values are below or equal to the oversold threshold.
func CheckOversoldAgo(rsi []*float64, rsiOversold float64) bool {
	for i := range rsi {
		if *rsi[i] <= rsiOversold {
			return true
		}
	}
	return false
}

// CheckOverboughtAgo checks if any of the last four RSI values are above or equal to the overbought threshold.
func CheckOverboughtAgo(rsi []*float64, rsiOverbought float64) bool {
	for i := range rsi {
		if *rsi[i] >= rsiOverbought {
			return true
		}
	}
	return false
}

// CheckBullishDivergence checks for bullish divergence conditions.
func CheckBullishDivergence(rsi []*float64) bool {
	return *rsi[2] > *rsi[1] && *rsi[1] < *rsi[0]
}

// CheckBearishDivergence checks for bearish divergence conditions.
func CheckBearishDivergence(rsi []*float64) bool {
	return *rsi[2] < *rsi[1] && *rsi[1] > *rsi[0]
}

func getRSI(up, down float64) float64 {
	if down == 0 {
		return 100.0
	} else if up == 0 {
		return 0.0
	} else {
		rsi := 100.0 - 100.0/(1.0+up/down)
		return rsi
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

func (l *Levels) ProcessData(marketData *block.Data) {
	l.volume = l.volume[1:]
	l.volume = append(l.volume, marketData.Volume)

	l.closePrice = l.closePrice[1:]
	l.closePrice = append(l.closePrice, marketData.ClosePrice)

	l.lowPrice = l.lowPrice[1:]
	l.lowPrice = append(l.lowPrice, marketData.Low)

	l.highPrice = l.highPrice[1:]
	l.highPrice = append(l.highPrice, marketData.High)

	l.openPrice = l.openPrice[1:]
	l.openPrice = append(l.openPrice, marketData.OpenPrice)

	volumePriceChange := marketData.ClosePrice - marketData.OpenPrice
	l.deltaPrice = l.deltaPrice[1:]
	l.deltaPrice = append(l.deltaPrice, volumePriceChange)
}

func (l *Levels) GetPotentialProfit(sell bool, targetPrice, currentPrice float64) float64 {
	if sell {
		return (currentPrice / targetPrice) - 1
	} else {
		return (targetPrice / currentPrice) - 1
	}
}
