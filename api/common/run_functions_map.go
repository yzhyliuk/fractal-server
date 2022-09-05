package common

import (
	"newTradingBot/models/trade"
	"newTradingBot/strategies/glide_on_price"
	"newTradingBot/strategies/linear_regression"
	"newTradingBot/strategies/mac"
	"newTradingBot/strategies/mean_reversion"
	qqe "newTradingBot/strategies/qqe_strategy"
	"newTradingBot/strategies/regression_channels"
	"newTradingBot/strategies/rsi_crossover"
	"newTradingBot/strategies/simple_rsi"
	"newTradingBot/strategies/trend_with_rsi"
)

var RunStrategy = map[int]func(userID int, rawConfig []byte, testing int, session *int) ([]*trade.Trade, error){
	1: mac.RunMovingAverageCrossover,
	2: glide_on_price.RunGlideOnPrice,
	3: trend_with_rsi.RunTrendWithRSI,
	4: rsi_crossover.RunRSICrossoverStrategy,
	//5: experimental.RunExperimentalContinuousStrategy,
	6: simple_rsi.RunSimpleRSI,
	7: qqe.RunQQEStrategy,
	8: mean_reversion.RunMeanReversion,
	9: linear_regression.RunLinearRegression,
	11: regression_channels.RunLinearRegression,
}