package actions

import (
	"newTradingBot/models/block"
	"newTradingBot/models/monitoring"
	"newTradingBot/models/recording"
	"newTradingBot/storage"
	"time"
)

func StartCapture(recordSession *recording.CapturedSession) error {
	newCaptureMonitor := monitoring.NewBinanceMonitor(recordSession.Symbol, time.Second*time.Duration(recordSession.TimeFrame), recordSession.IsFutures)
	monitorName := recordSession.GetMonitorName()
	var inputChan chan *block.Data

	if storage.BinanceCaptureMonitors[monitorName] != nil {
		inputChan = storage.BinanceCaptureMonitors[monitorName].Subscribe(recordSession.ID)
	} else {
		storage.BinanceCaptureMonitors[monitorName] = newCaptureMonitor
		inputChan = storage.BinanceCaptureMonitors[monitorName].Subscribe(recordSession.ID)
	}

	recorder := recording.NewRecorder(inputChan, recordSession)

	storage.Recorders[recordSession.ID] = recorder

	storage.BinanceCaptureMonitors[monitorName].RunMonitor()
	return storage.Recorders[recordSession.ID].Start()
}
