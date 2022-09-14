package actions

import (
	"fmt"
	"newTradingBot/api/database"
	"newTradingBot/models/block"
	"newTradingBot/models/monitoring"
	"newTradingBot/models/monitoring/replay"
	"newTradingBot/models/recording"
	"newTradingBot/models/strategy/configs"
	"newTradingBot/models/strategy/instance"
	"newTradingBot/models/testing"
	"newTradingBot/models/users"
	"newTradingBot/storage"
	"time"
)

func PrepareStrategy(conf configs.BaseStrategyConfig, userID int, strategyID int) (*instance.StrategyInstance, chan *block.Data, *users.Keys, error) {
	inst := instance.GetInstanceFromConfig(conf, userID, strategyID)

	db, err := database.GetDataBaseConnection()
	if err != nil {
		return nil, nil, nil, err
	}

	inst, err = instance.CreateStrategyInstance(db, inst)
	if err != nil {
		return nil, nil, nil, err
	}

	monitorName := fmt.Sprintf("%s:%d:%t",conf.Pair, conf.TimeFrame, inst.IsFutures)

	var monitorChannel chan *block.Data
	storage.MonitorsLock.Lock()
	if storage.MonitorsBinance[monitorName] != nil{
		monitorChannel = storage.MonitorsBinance[monitorName].Subscribe(inst.ID)
	} else {
		monitor := monitoring.NewBinanceMonitor(conf.Pair, time.Duration(conf.TimeFrame*int(time.Second)),inst.IsFutures)
		storage.MonitorsBinance[monitorName] = monitor
		storage.MonitorsBinance[monitorName].RunMonitor()
		monitorChannel = storage.MonitorsBinance[monitorName].Subscribe(inst.ID)
	}
	storage.MonitorsLock.Unlock()

	keys, err := users.GetUserKeys(db, userID)
	if err != nil {
		return nil, nil, nil, err
	}

	return inst, monitorChannel, &keys, nil
}

func PrepareExperimentalStrategy(conf configs.BaseStrategyConfig, userID int, strategyID int) (*instance.StrategyInstance, *users.Keys, error) {
	inst := instance.GetInstanceFromConfig(conf, userID, strategyID)

	db, err := database.GetDataBaseConnection()
	if err != nil {
		return nil, nil, err
	}

	inst, err = instance.CreateStrategyInstance(db, inst)
	if err != nil {
		return nil, nil, err
	}

	keys, err := users.GetUserKeys(db, userID)
	if err != nil {
		return nil, nil, err
	}

	return inst, &keys, nil
}

func PrepareBackTesting(conf configs.BaseStrategyConfig, captureSession, userID, strategyID int) (*instance.StrategyInstance, *replay.MonitorReplay, *users.Keys, error) {
	db, err := database.GetDataBaseConnection()
	if err != nil {
		return nil, nil, nil, err
	}

	var session recording.CapturedSession
	err = db.Where("id = ?", captureSession).Find(&session).Error
	if err != nil {
		return nil, nil, nil ,err
	}

	conf.Pair = session.Symbol
	conf.IsFutures = session.IsFutures

	inst := instance.GetInstanceFromConfig(conf, userID, strategyID)

	inst.TimeFrame = session.TimeFrame
	inst.Testing = testing.BackTest

	keys, err := users.GetUserKeys(db, userID)
	if err != nil {
		return nil, nil, nil, err
	}

	replayMonitor := replay.NewMonitorReplay(captureSession)

	return inst, replayMonitor, &keys, nil

}