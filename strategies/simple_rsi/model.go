package simple_rsi

import (
	"encoding/json"
	"fmt"
	"newTradingBot/api/database"
	"newTradingBot/indicators"
	"newTradingBot/logs"
	"newTradingBot/models/account"
	"newTradingBot/models/block"
	"newTradingBot/models/strategy"
	"newTradingBot/models/strategy/instance"
	"newTradingBot/models/users"
	"newTradingBot/strategies/common"
)

const StrategyName = "simple_rsi"

type SimpleRSI struct {
	common.Strategy

	config                 SimpleRSIConfig
	closePriceObservations []float64

	orderUptrend bool
}

// NewSimpleRSI - creates new Moving Average crossover strategy
func NewSimpleRSI(monitorChannel chan *block.Data, configRaw []byte, keys *users.Keys, historicalData []*block.Data, inst *instance.StrategyInstance) (strategy.Strategy, error) {

	var config SimpleRSIConfig

	err := json.Unmarshal(configRaw, &config)
	if err != nil {
		return nil,err
	}

	acc, err := account.NewBinanceAccount(keys.ApiKey,keys.SecretKey, keys.ApiKey, keys.SecretKey)
	if err != nil {
		return nil, err
	}


	newStrategy := &SimpleRSI{
		config: config,
	}
	newStrategy.Account = acc
	newStrategy.StopSignal = make(chan bool)
	newStrategy.MonitorChannel = monitorChannel
	newStrategy.StrategyInstance = inst
	newStrategy.HandlerFunction = newStrategy.HandlerFunc
	newStrategy.DataProcessFunction = newStrategy.ProcessData

	newStrategy.closePriceObservations = make([]float64, newStrategy.config.RSIPeriod)

	return newStrategy, nil
}

func (t *SimpleRSI) HandlerFunc(marketData *block.Data)  {
	if t.closePriceObservations[0] != 0{
		if t.StrategyInstance.Status == instance.StatusCreated {
			t.StrategyInstance.Status = instance.StatusRunning
			db, _ := database.GetDataBaseConnection()
			_ = instance.UpdateStatus(db, t.StrategyInstance.ID, instance.StatusRunning)
		}

		rsi := indicators.RSI(t.closePriceObservations, t.config.RSIPeriod)


		rsiOverbought := rsi > float64(t.config.RSIOverboughtLevel)
		rsiOversold := rsi < float64(t.config.RSIOversoldLevel)

		logs.LogDebug(fmt.Sprintf("RSI: %f ", rsi), nil)

		t.Evaluate(marketData,rsi,rsiOverbought,rsiOversold)
	}
}

func (t *SimpleRSI) Evaluate(marketData *block.Data, rsi float64, rsiOverbought, rsiOversold bool)  {
	if rsiOversold {
		err := t.HandleBuy(marketData)
		if err != nil {
			logs.LogDebug("", err)
			return
		}
	}

	if rsiOverbought {
		err := t.HandleSell(marketData)
		if err != nil {
			logs.LogDebug("", err)
			return
		}
	}
}

func (t *SimpleRSI) ProcessData(marketData *block.Data)  {
	t.closePriceObservations = t.closePriceObservations[1:]
	t.closePriceObservations = append(t.closePriceObservations, marketData.ClosePrice)
}
