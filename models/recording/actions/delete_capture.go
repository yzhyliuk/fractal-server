package actions

import (
	"gorm.io/gorm"
	"newTradingBot/models/recording"
)

func DeleteCaptureSession(db *gorm.DB, userID, sessionID int) error {
	return db.Where("id = ? AND userid = ?", sessionID, userID).Delete(&recording.CapturedSession{}).Error
}
