package rsi_crossover

import "newTradingBot/models/strategy/configs"

type RSICrossoverConfig struct {
	configs.BaseStrategyConfig
	LongTermPeriod int `json:"longTerm"`
	ShortTermPeriod int `json:"shortTerm"`
	RSIPeriod int `json:"rsiPeriod"`
}

