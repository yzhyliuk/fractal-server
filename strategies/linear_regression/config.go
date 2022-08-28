package linear_regression

import "newTradingBot/models/strategy/configs"

type LinearRegressionConfig struct {
	configs.BaseStrategyConfig
	Period int `json:"period"`
}


