package fibonacci_with_levels

import "newTradingBot/models/strategy/configs"

type FibonacciRetraceWithLevelsConfig struct {
	configs.BaseStrategyConfig
	MALength       int     `json:"maLength"`
	BBLength       int     `json:"bbLength"`
	BBMultiplier   float64 `json:"bbMultiplier"`
	FibonacciLevel float64 `json:"fibonacciLevel"`
}
