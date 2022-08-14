package common

import (
	"newTradingBot/models/trade"
	"newTradingBot/strategies/glide_on_price"
	"newTradingBot/strategies/mac"
	"newTradingBot/strategies/rsi_crossover"
	"newTradingBot/strategies/trend_with_rsi"
)

var RunStrategy = map[int]func(userID int, rawConfig []byte, testing int, session *int) ([]*trade.Trade, error){
	1: mac.RunMovingAverageCrossover,
	2: glide_on_price.RunGlideOnPrice,
	3: trend_with_rsi.RunTrendWithRSI,
	4: rsi_crossover.RunRSICrossoverStrategy,
	//5: experimental.RunExperimentalContinuousStrategy,
}