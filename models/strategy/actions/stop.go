package actions

import (
	"fmt"
	"gorm.io/gorm"
	"newTradingBot/models/strategy/instance"
	"newTradingBot/storage"
)

// StopStrategy - stops strategy instance
func StopStrategy(db *gorm.DB, inst *instance.StrategyInstance) error {
	monitorName := fmt.Sprintf("%s:%d:%t",inst.Pair, inst.TimeFrame, inst.IsFutures)

	storage.MonitorsBinance[monitorName].UnSubscribe(inst.ID)

	go storage.StrategiesStorage[inst.ID].Stop()
	delete(storage.StrategiesStorage, inst.ID)

	if storage.MonitorsBinance[monitorName].IsEmptySubs() {
		delete(storage.MonitorsBinance, monitorName)
	}

	return nil
}

