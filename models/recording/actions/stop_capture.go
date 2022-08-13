package actions

import (
	"gorm.io/gorm"
	"newTradingBot/models/recording"
	"newTradingBot/storage"
)

func StopRecording(db *gorm.DB, sessionID int) error {
	var recordSession *recording.CapturedSession
	err := db.Where("id = ?", sessionID).Find(&recordSession).Error
	if err != nil {
		return err
	}

	monitorName := recordSession.GetMonitorName()

	storage.BinanceCaptureMonitors[monitorName].UnSubscribe(sessionID)
	if storage.BinanceCaptureMonitors[monitorName].IsEmptySubs() {
		delete(storage.BinanceCaptureMonitors, monitorName)
	}

	err = storage.Recorders[sessionID].Stop()
	if err != nil {
		return err
	}

	delete(storage.Recorders, sessionID)

	return nil
}
