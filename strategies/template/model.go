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

type FrameData struct {
	Time time.Time `json:"time"`
}


// NewTemplateStrategy - creates new Moving Average crossover strategy
func NewTemplateStrategy(monitorChannel chan *block.Block, config StrategyTemplateConfig, keys *users.Keys, historicalData []*block.Block, inst *instance.StrategyInstance) (strategy.Strategy, error) {

	acc, err := account.NewBinanceAccount(keys.ApiKey,keys.SecretKey, keys.ApiKey, keys.SecretKey)
	if err != nil {
		return nil, err
	}


	strategy := &templateStrategy{
		config: config,
	}
	strategy.Account = acc
	strategy.StopSignal = make(chan bool)
	strategy.MonitorChannel = monitorChannel
	strategy.StrategyInstance = inst
	strategy.HandlerFunction = strategy.HandlerFunc
	strategy.DataProcessFunction = strategy.ProcessData

	return strategy, nil
}

func (m *templateStrategy) HandlerFunc(marketData *block.Block)  {

}

func (m *templateStrategy) ProcessData(marketData *block.Block)  {
}
