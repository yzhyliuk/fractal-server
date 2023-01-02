package regression_channels

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

const StrategyName = "regression_channels"

type linearRegression struct {
	common.Strategy

	config                 LinearRegressionConfig
	closePriceObservations []float64
	highPriceObservations []float64
	lowPriceObservations []float64

	breakUp bool
	breakDown bool

	sdMultiplier float64
}

// NewLinearRegression - creates new Moving Average crossover strategy
func NewLinearRegression(monitorChannel chan *block.Data, configRaw []byte, keys *users.Keys, historicalData []*block.Data, inst *instance.StrategyInstance) (strategy.Strategy, error) {
	var config LinearRegressionConfig

	err := json.Unmarshal(configRaw, &config)
	if err != nil {
		return nil,err
	}

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
	newStrategy.sdMultiplier = 2

	newStrategy.closePriceObservations = make([]float64, newStrategy.config.Period)
	newStrategy.highPriceObservations = make([]float64, newStrategy.config.Period)
	newStrategy.lowPriceObservations = make([]float64, newStrategy.config.Period)

	return newStrategy, nil
}

func (l *linearRegression) HandlerFunc(marketData *block.Data)  {
	if l.closePriceObservations[0] != 0 {
		if l.StrategyInstance.Status == instance.StatusCreated && l.StrategyInstance.Testing == testing.Disable {
			l.StrategyInstance.Status = instance.StatusRunning
			db, _ := database.GetDataBaseConnection()
			_ = instance.UpdateStatus(db, l.StrategyInstance.ID, instance.StatusRunning)
		}

		slope, intercept := indicators.LinearRegressionForTimeSeries(l.closePriceObservations)

		linearMean := intercept + slope*float64(l.config.Period)
		regularMean := indicators.Average(l.closePriceObservations)

		sdLinear := indicators.StandardDeviationWithMean(l.closePriceObservations, linearMean)
		sdRegular := indicators.StandardDeviation(l.closePriceObservations)

		upperLinearLine := linearMean+(sdLinear*l.sdMultiplier)
		lowerLinearLine := linearMean-(sdLinear*l.sdMultiplier)

		upperRegularLine := regularMean+(sdRegular*l.sdMultiplier)
		lowerRegularLine := regularMean-(sdRegular*l.sdMultiplier)


		l.Evaluate(marketData, upperLinearLine, lowerLinearLine, upperRegularLine, lowerRegularLine)
	}
}

func (l *linearRegression) Evaluate(marketData *block.Data, upLinear, lowLinear, upRegular, lowRegular float64)  {

	targetUp := marketData.ClosePrice
	targetDown := marketData.ClosePrice

	if l.config.TargetParameter == "high" {
		targetUp = marketData.High
		targetDown = marketData.Low
	}

	// handle exit
	if l.LastTrade != nil {
		exitTargetUp := marketData.High
		exitTargetDown := marketData.Low
		conditionOne := l.LastTrade.FuturesSide == futures.SideTypeSell && exitTargetDown < lowLinear
		conditionTwo := l.LastTrade.FuturesSide == futures.SideTypeBuy && exitTargetUp > upLinear

		if conditionOne || conditionTwo {
			l.CloseAllTrades()
			return
		}
	}

	if lowLinear > targetDown && lowRegular > targetDown{
		l.breakDown = true
		return
	} else if upLinear < targetUp && upRegular < targetUp {
		l.breakUp = true
		return
	}

	if l.breakUp && upLinear > targetUp {
		l.breakUp = false
		err := l.HandleSell(marketData)
		if err != nil {
			logs.LogError(err)
		}
		return
	} else if l.breakDown && lowLinear < targetDown{
		l.breakDown = false
		err := l.HandleBuy(marketData)
		if err != nil {
			logs.LogError(err)
		}
		return
	}

}

func (l *linearRegression) ProcessData(marketData *block.Data)  {
	l.closePriceObservations = l.closePriceObservations[1:]
	l.closePriceObservations = append(l.closePriceObservations, marketData.ClosePrice)

	l.highPriceObservations = l.highPriceObservations[1:]
	l.highPriceObservations = append(l.highPriceObservations, marketData.High)

	l.lowPriceObservations = l.lowPriceObservations[1:]
	l.lowPriceObservations = append(l.lowPriceObservations, marketData.Low)
}

func (l *linearRegression) LastLow() float64 {
	return l.lowPriceObservations[l.config.Period-1]
}

func (l *linearRegression) LastHigh() float64 {
	return l.highPriceObservations[l.config.Period-1]
}


