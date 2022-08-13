package storage

import (
	"newTradingBot/models/strategy"
)

var StrategiesStorage = make(map[int]strategy.Strategy)
