package glide_on_price

import "newTradingBot/models/strategy/configs"

type GlideOnPriceConfig struct {
	configs.BaseStrategyConfig
	VolatilityLimit float64 `json:"volatilityLimit"`
	VolatilityObservationsTimeFrame int `json:"volatilityTF"`
	SlopeTimeFrameDifference int `json:"slopeTF"`
}
