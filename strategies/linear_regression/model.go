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
	"newTradingBot/models/users"
	"newTradingBot/strategies/common"
)

type linearRegression struct {
	common.Strategy

	config                 LinearRegressionConfig
	closePriceObservations []float64
	forecastTimeFrame int
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

		// Strategy here

		slope, _ := indicators.LinearRegressionForTimeSeries(l.closePriceObservations)

		logs.LogDebug(fmt.Sprintf("SLOPE: %f", slope), nil)

		l.Evaluate(marketData, slope)
	}
}

func (l *linearRegression) Evaluate(marketData *block.Data, slope float64)  {
	if slope < 0 {
		err := l.HandleSell(marketData)
		logs.LogError(err)
	} else if slope > 0 {
		err := l.HandleBuy(marketData)
		logs.LogError(err)
	}
}

func (l *linearRegression) ProcessData(marketData *block.Data)  {
	l.closePriceObservations = l.closePriceObservations[1:]
	l.closePriceObservations = append(l.closePriceObservations, marketData.ClosePrice)
}


