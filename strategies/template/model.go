package template

import (
	"newTradingBot/models/account"
	"newTradingBot/models/block"
	"newTradingBot/models/strategy"
	"newTradingBot/models/strategy/instance"
	"newTradingBot/models/users"
	"newTradingBot/strategies/common"
	"time"
)

type templateStrategy struct {
	common.Strategy

	config StrategyTemplateConfig
}

// NewTemplateStrategy - creates new Moving Average crossover strategy
func NewTemplateStrategy(monitorChannel chan *block.Block, config StrategyTemplateConfig, keys *users.Keys, historicalData []*block.Block, inst *instance.StrategyInstance) (strategy.Strategy, error) {

	acc, err := account.NewBinanceAccount(keys.ApiKey,keys.SecretKey, keys.ApiKey, keys.SecretKey)
	if err != nil {
		return nil, err
	}


	newStrategy := &templateStrategy{
		config: config,
	}
	newStrategy.Account = acc
	newStrategy.StopSignal = make(chan bool)
	newStrategy.MonitorChannel = monitorChannel
	newStrategy.StrategyInstance = inst
	newStrategy.HandlerFunction = newStrategy.HandlerFunc
	newStrategy.DataProcessFunction = newStrategy.ProcessData

	return newStrategy, nil
}


type FrameData struct {
	Time time.Time `json:"time"`
}

func (m *templateStrategy) HandlerFunc(marketData *block.Block)  {

}

func (m *templateStrategy) ProcessData(marketData *block.Block)  {
}
