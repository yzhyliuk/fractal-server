package experimental

import (
	"fmt"
	"github.com/adshao/go-binance/v2/futures"
	"newTradingBot/api/database"
	"newTradingBot/indicators"
	"newTradingBot/logs"
	"newTradingBot/models/account"
	"newTradingBot/models/strategy"
	"newTradingBot/models/strategy/instance"
	"newTradingBot/models/users"
	"newTradingBot/strategies/common"
	"strconv"
)

type experimentalContinuousStrategy struct {
	common.Strategy

	config continuousConfig
	stopChannel chan struct{}

	priceObservations []float64
}

func NewContinuousExperimentalStrategy(config continuousConfig, keys *users.Keys, inst *instance.StrategyInstance) (strategy.Strategy, error) {
	acc, err := account.NewBinanceAccount(keys.ApiKey,keys.SecretKey, keys.ApiKey, keys.SecretKey)
	if err != nil {
		return nil, err
	}


	newStrategy := &experimentalContinuousStrategy{
		config: config,
	}
	newStrategy.Account = acc
	newStrategy.StopSignal = make(chan bool)
	newStrategy.StrategyInstance = inst
	newStrategy.ExperimentalHandler = newStrategy.Start

	newStrategy.priceObservations = make([]float64, 400)

	return newStrategy, nil
}

func (e *experimentalContinuousStrategy) Start()  {
	if e.StrategyInstance.Status == instance.StatusCreated {
		e.StrategyInstance.Status = instance.StatusRunning
		db, _ := database.GetDataBaseConnection()
		_ = instance.UpdateStatus(db, e.StrategyInstance.ID, instance.StatusRunning)
	}

	wsAggTradeHandler := func(event *futures.WsAggTradeEvent) {
		price, _ := strconv.ParseFloat(event.Price, 64)
		//quantity, _ := strconv.ParseFloat(event.Quantity, 64)

		e.priceObservations = e.priceObservations[1:]
		e.priceObservations = append(e.priceObservations, price)

		if e.priceObservations[0] != 0 {
			slope, _ := indicators.LinearRegressionForTimeSeries(e.priceObservations)
			logs.LogDebug(fmt.Sprintf("SL: %f \t PRICE: %f", slope, price),nil)
			//sd := indicators.StandardDeviation(e.priceObservations)
			//mean := indicators.SimpleMA(e.priceObservations, len(e.priceObservations))
			//
			//if e.LastTrade != nil && e.LastTrade.FuturesSide == futures.SideTypeBuy{
			//	if price > mean {
			//		e.CloseAllTrades()
			//	}
			//}
			//
			//if e.LastTrade != nil && e.LastTrade.FuturesSide == futures.SideTypeSell{
			//	if price < mean {
			//		e.CloseAllTrades()
			//	}
			//}
			//
			//if price < mean-(sd*3) {
			//	err := e.HandleBuy(&block.Data{
			//		ClosePrice: price,
			//	})
			//	if err != nil {
			//		logs.LogDebug("", err)
			//	}
			//} else if price > mean+(sd*3) {
			//	err := e.HandleSell(&block.Data{
			//		ClosePrice: price,
			//	})
			//	if err != nil {
			//		logs.LogDebug("", err)
			//	}
			//}
	//		//logs.LogDebug(fmt.Sprintf("\nPRICE: %f \t QUANTITY: %f \n SD: %f \t MEAN: %f", price, quantity, sd, mean ),nil)
		}
	}

	_, stopC, _ := futures.WsAggTradeServe(e.config.Pair, wsAggTradeHandler, errHandlerLog)

	e.stopChannel = stopC
}

func errHandlerLog(err error)  {
	logs.LogDebug("", err)
}

func (e *experimentalContinuousStrategy) Stop()  {

	db, err := database.GetDataBaseConnection()
	if err != nil {
		logs.LogDebug("", err)
	}
	err = instance.UpdateStatus(db, e.StrategyInstance.ID, instance.StatusStopped)
	if err != nil {
		logs.LogDebug("", err)
	}

	e.stopChannel <- struct{}{}
}