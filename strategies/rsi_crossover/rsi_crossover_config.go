package rsi_crossover

import "newTradingBot/models/strategy/configs"

type RSICrossoverConfig struct {
	configs.BaseStrategyConfig
	LongTermPeriod int `json:"longTerm"`
	RSIPeriod int `json:"rsiPeriod"`
	Volatility float64 `json:"volatility"`
}

