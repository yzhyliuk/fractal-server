package template

import "newTradingBot/models/strategy/configs"

type StrategyTemplateConfig struct {
	configs.BaseStrategyConfig
	LongTermPeriod int `json:"longTerm"`
	ShortTermPeriod int `json:"shortTerm"`
	MovingAverageType string `json:"maType"`
}
