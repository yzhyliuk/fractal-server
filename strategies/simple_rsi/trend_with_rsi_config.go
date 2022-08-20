package simple_rsi

import "newTradingBot/models/strategy/configs"

type SimpleRSIConfig struct {
	configs.BaseStrategyConfig
	RSIPeriod int `json:"rsiPeriod"`
	RSIOversoldLevel int `json:"oversold"`
	RSIOverboughtLevel int `json:"overbought"`
}

