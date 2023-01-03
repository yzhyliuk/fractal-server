package bollinger_bands_with_atr

import "newTradingBot/models/strategy/configs"

type BollingerBandsWithATRConfig struct {
	configs.BaseStrategyConfig
	MALength int `json:"maLength"`
	BBLength int `json:"bbLength"`
	BBMultiplier float64 `json:"bbMultiplier"`
	ATRLength int `json:"atrLength"`
}
