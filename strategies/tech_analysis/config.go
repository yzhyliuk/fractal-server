package tech_analysis

import "newTradingBot/models/strategy/configs"

type TechAnalysisConfig struct {
	configs.BaseStrategyConfig
	TrendLength          int     `json:"trendLength"`
	BollingerLength      int     `json:"bollingerLength"`
	BollingerMultiplier  float64 `json:"bollingerMultiplier"`
	PriceChannelLookBack float64 `json:"priceChannelLookBack"`
}
