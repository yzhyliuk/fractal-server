package regression_channels
import "newTradingBot/models/strategy/configs"

type LinearRegressionConfig struct {
	configs.BaseStrategyConfig
	Period int `json:"period"`
	TargetParameter string `json:"target"`
}



