package common

import (
	"newTradingBot/strategies/experimental"
	"newTradingBot/strategies/glide_on_price"
	"newTradingBot/strategies/mac"
	"newTradingBot/strategies/rsi_crossover"
	"newTradingBot/strategies/trend_with_rsi"
)

var RunStrategy = map[int]func(userID int, rawConfig []byte) error{
	1: mac.RunMovingAverageCrossover,
	2: glide_on_price.RunGlideOnPrice,
	3: trend_with_rsi.RunTrendWithRSI,
	4: rsi_crossover.RunRSICrossoverStrategy,
	5: experimental.RunExperimentalContinuousStrategy,
}
