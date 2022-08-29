package linear_regression

import (
	"fmt"
	"newTradingBot/api/database"
	"newTradingBot/indicators"
	"newTradingBot/logs"
	"newTradingBot/models/account"
	"newTradingBot/models/block"
	"newTradingBot/models/strategy"
	"newTradingBot/models/strategy/instance"
	"newTradingBot/models/testing"
	"newTradingBot/models/trade"
	"newTradingBot/models/users"
	"newTradingBot/strategies/common"
)

type linearRegression struct {
	common.Strategy

	config                 LinearRegressionConfig
	closePriceObservations []float64
	forecastTimeFrame int
	timeFrameCounter int

	prevSlopeDirectionUp bool
}

// NewLinearRegression - creates new Moving Average crossover strategy
func NewLinearRegression(monitorChannel chan *block.Data, config LinearRegressionConfig, keys *users.Keys, historicalData []*block.Data, inst *instance.StrategyInstance) (strategy.Strategy, error) {

	acc, err := account.NewBinanceAccount(keys.ApiKey,keys.SecretKey, keys.ApiKey, keys.SecretKey)
	if err != nil {
		return nil, err
	}


	newStrategy := &linearRegression{
		config: config,
	}
	newStrategy.Account = acc
	newStrategy.StopSignal = make(chan bool)
	newStrategy.MonitorChannel = monitorChannel
	newStrategy.StrategyInstance = inst
	newStrategy.HandlerFunction = newStrategy.HandlerFunc
	newStrategy.DataProcessFunction = newStrategy.ProcessData
	newStrategy.forecastTimeFrame = 10
	newStrategy.timeFrameCounter = 0

	newStrategy.closePriceObservations = make([]float64, newStrategy.config.Period)

	return newStrategy, nil
}

func (l *linearRegression) HandlerFunc(marketData *block.Data)  {
	if l.closePriceObservations[0] != 0 {
		if l.StrategyInstance.Status == instance.StatusCreated && l.StrategyInstance.Testing == testing.Disable {
			l.StrategyInstance.Status = instance.StatusRunning
			db, _ := database.GetDataBaseConnection()
			_ = instance.UpdateStatus(db, l.StrategyInstance.ID, instance.StatusRunning)
		}

		if l.timeFrameCounter != 0  && l.timeFrameCounter <= l.forecastTimeFrame{
			l.timeFrameCounter++
			return
		} else if l.timeFrameCounter > l.forecastTimeFrame {
			l.CloseAllTrades()
			l.timeFrameCounter = 0
		}

		// Strategy here

		slope, intercept := indicators.LinearRegressionForTimeSeries(l.closePriceObservations)

		predictedPrice := intercept + (slope*(float64(l.config.Period+l.forecastTimeFrame)))

		quantity := l.config.BidSize/marketData.ClosePrice
		bid := l.config.BidSize/float64(*l.config.Leverage)
		fee := l.config.BidSize*trade.BinanceFuturesTakerFeeRate

		tradeBuyProfit := ((predictedPrice*quantity)-(marketData.ClosePrice*quantity))-2*fee
		tradeSellProfit := ((marketData.ClosePrice*quantity)-(predictedPrice*quantity))-2*fee

		tradeBuyROI := tradeBuyProfit/bid
		tradeSellROI := tradeSellProfit/bid



		logs.LogDebug(fmt.Sprintf("PREDICTED PRICE: %f SLOPE: %f",predictedPrice, slope), nil)

		l.Evaluate(marketData, slope, tradeBuyROI, tradeSellROI)
	}
}

func (l *linearRegression) Evaluate(marketData *block.Data, slope, roiBUY,roiSELL float64)  {
	currentTrend := slope > 0
	//trendChange := l.prevSlopeDirectionUp != currentTrend
	l.prevSlopeDirectionUp = currentTrend

		if roiBUY > l.config.TradeTakeProfit {
			err := l.HandleBuy(marketData)
			logs.LogError(err)
			if err == nil {
				l.timeFrameCounter = 1
			}
		} else if roiSELL > l.config.TradeTakeProfit {
			err := l.HandleSell(marketData)
			logs.LogError(err)
			if err == nil {
				l.timeFrameCounter = 1
			}
		}

	//if slope < 0 && trendChange{
	//	err := l.HandleSell(marketData)
	//	logs.LogError(err)
	//} else if slope > 0 && trendChange{
	//	err := l.HandleBuy(marketData)
	//	logs.LogError(err)
	//}

}

func (l *linearRegression) ProcessData(marketData *block.Data)  {
	l.closePriceObservations = l.closePriceObservations[1:]
	l.closePriceObservations = append(l.closePriceObservations, marketData.ClosePrice)
}


