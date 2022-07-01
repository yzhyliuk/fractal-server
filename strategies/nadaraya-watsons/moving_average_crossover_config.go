package nadaraya_watsons

import "newTradingBot/models/strategy/configs"

type MovingAverageCrossoverConfig struct {
	configs.BaseStrategyConfig
	Multiplier int `json:"mul"`
}
