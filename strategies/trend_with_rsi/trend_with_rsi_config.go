package trend_with_rsi

import "newTradingBot/models/strategy/configs"

type TrendWithRSIConfig struct {
	configs.BaseStrategyConfig
	RSIPeriod int `json:"rsiPeriod"`
	TrendMAPeriod int `json:"trendMA"`
	RSIOversoldLevel int `json:"rsiOversoldLevel"`
	RSIOverboughtLevel int `json:"rsiOverboughtLevel"`
}

