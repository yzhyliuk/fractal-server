package mac

import "newTradingBot/models/strategy/configs"

type MovingAverageCrossoverConfig struct {
	configs.BaseStrategyConfig
	LongTermPeriod int `json:"longTerm"`
	ShortTermPeriod int `json:"shortTerm"`
	MovingAverageType string `json:"maType"`
}
