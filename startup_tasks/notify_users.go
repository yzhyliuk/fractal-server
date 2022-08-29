package startup_tasks

import (
	"newTradingBot/api/database"
	"newTradingBot/configuration"
	"newTradingBot/models/notifications"
)

func NotifyUsers() error {

	if configuration.IsDebugProd() {
		return nil
	}

	db, err := database.GetDataBaseConnection()
	if err != nil {
		return err
	}

	return notifications.CreateGeneralNotification(db, notifications.Info, notifications.ServerRestartedMessage())
}
