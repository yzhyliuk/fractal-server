package price_channel_breakout

import "newTradingBot/models/strategy/configs"

type PriceChannelBreakoutConfig struct {
	configs.BaseStrategyConfig
	LongTermPeriod    int    `json:"longTerm"`
	ShortTermPeriod   int    `json:"shortTerm"`
	MovingAverageType string `json:"maType"`
}
