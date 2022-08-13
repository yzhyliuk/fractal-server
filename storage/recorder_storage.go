package storage

import (
	"newTradingBot/models/monitoring"
	"newTradingBot/models/recording"
)

var BinanceCaptureMonitors = make(map[string]*monitoring.BinanceMonitor)
var Recorders = make(map[int]*recording.Recorder)
