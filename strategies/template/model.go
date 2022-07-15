package template

import (
	"newTradingBot/models/block"
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

func (m *templateStrategy) HandlerFunc(marketData *block.Block)  {

}

func (m *templateStrategy) ProcessData(marketData *block.Block)  {
}
