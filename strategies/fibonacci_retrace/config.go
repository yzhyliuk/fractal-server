package fibonacci_retrace

import "newTradingBot/models/strategy/configs"

type FibonacciRetraceConfig struct {
	configs.BaseStrategyConfig
	MALength       int     `json:"maLength"`
	BBLength       int     `json:"bbLength"`
	BBMultiplier   float64 `json:"bbMultiplier"`
	FibonacciLevel float64 `json:"fibonacciLevel"`
}
