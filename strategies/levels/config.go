package levels

import "newTradingBot/models/strategy/configs"

type LevelsConfig struct {
	configs.BaseStrategyConfig
	MomentumLength int     `json:"momentumLength"`
	RSIOverbought  int     `json:"RSIOverbought"`
	RSIOversold    int     `json:"RSIOversold"`
	RSILength      int     `json:"RSILength"`
	EMA            int     `json:"EMA"`
	BandMultiplier float64 `json:"BandMultiplier"`
}
