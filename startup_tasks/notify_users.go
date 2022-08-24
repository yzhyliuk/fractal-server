package startup_tasks

import (
	"newTradingBot/api/database"
	"newTradingBot/models/notifications"
)

func NotifyUsers() error {
	db, err := database.GetDataBaseConnection()
	if err != nil {
		return err
	}

	return notifications.CreateGeneralNotification(db, notifications.Info, notifications.ServerRestartedMessage())
}
