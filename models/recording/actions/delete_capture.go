package actions

import (
	"gorm.io/gorm"
	"newTradingBot/models/recording"
)

func DeleteCaptureSession(db *gorm.DB, sessionID int) error {
	return db.Where("id = ?", sessionID).Delete(&recording.CapturedSession{}).Error
}
