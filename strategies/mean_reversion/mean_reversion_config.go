package mean_reversion

import "newTradingBot/models/strategy/configs"

type MeanReversionConfig struct {
	configs.BaseStrategyConfig
	MeanPeriod int `json:"mean"`
	SDMultiplier int `json:"sdMultiplier"`
}

