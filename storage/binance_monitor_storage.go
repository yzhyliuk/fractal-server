package storage

import (
	"newTradingBot/models/monitoring"
	"newTradingBot/models/strategy"
)

var MonitorsBinance = make(map[string]*monitoring.BinanceMonitor)

var StrategiesStorage = make(map[int]strategy.Strategy)