package actions

import (
	"fmt"
	"gorm.io/gorm"
	"newTradingBot/models/apimodels"
	"newTradingBot/models/strategy/instance"
	"newTradingBot/storage"
)

// StopStrategy - stops strategy instance
func StopStrategy(db *gorm.DB, inst *instance.StrategyInstance) error {
	info := apimodels.StrategyInfo{}
	db.Where("id = ?", inst.StrategyID).Find(&info)

	if !info.IsContinuous {
		monitorName := fmt.Sprintf("%s:%d:%t",inst.Pair, inst.TimeFrame, inst.IsFutures)

		storage.MonitorsBinance[monitorName].UnSubscribe(inst.ID)

		if storage.MonitorsBinance[monitorName].IsEmptySubs() {
			delete(storage.MonitorsBinance, monitorName)
		}
	}

	go storage.StrategiesStorage[inst.ID].Stop()
	delete(storage.StrategiesStorage, inst.ID)

	return nil
}

