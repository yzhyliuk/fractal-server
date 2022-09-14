package storage

import (
	"newTradingBot/models/monitoring"
	"sync"
)

var MonitorsLock sync.Mutex
var MonitorsBinance = make(map[string]*monitoring.BinanceMonitor)
