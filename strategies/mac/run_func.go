package mac

import (
	"encoding/json"
	"fmt"
	"newTradingBot/api/database"
	"newTradingBot/api/helpers"
	"newTradingBot/models/block"
	"newTradingBot/models/monitoring"
	"newTradingBot/models/strategy/instance"
	"newTradingBot/models/users"
	"newTradingBot/storage"
	"time"
)

func RunMovingAverageCrossover(userID int, rawConfig []byte) error{
	var config MovingAverageCrossoverConfig

	err := json.Unmarshal(rawConfig, &config)
	if err != nil {
		return err
	}

	inst := &instance.StrategyInstance{
		Pair:       config.Pair,
		Bid:        config.BidSize,
		UserID:     userID,
		StrategyID: 1,
		TimeFrame:  config.TimeFrame,
		Status:     helpers.Created,
	}

	db, err := database.GetDataBaseConnection()
	if err != nil {
		return err
	}

	inst, err = instance.CreateStrategyInstance(db, inst)
	if err != nil {
		return err
	}

	monitorName := fmt.Sprintf("%s:%d",config.Pair, config.TimeFrame)

	var monitorChannel chan *block.Block

	var observationsData []*block.Block

	if storage.SpotMonitorsBinance[monitorName] != nil {
		monitorChannel = storage.SpotMonitorsBinance[monitorName].Subscribe(inst.ID)
		observationsData, err = storage.SpotMonitorsBinance[monitorName].GetHistoricalData(config.LongTermPeriod)
	} else {
		monitor := monitoring.NewBinanceMonitor(config.Pair, time.Duration(config.TimeFrame*int(time.Minute)))
		storage.SpotMonitorsBinance[monitorName] = monitor
		storage.SpotMonitorsBinance[monitorName].RunMonitor()
		monitorChannel = storage.SpotMonitorsBinance[monitorName].Subscribe(inst.ID)
		observationsData = make([]*block.Block, config.LongTermPeriod)
	}

	keys, err := users.GetUserKeys(db, userID)
	if err != nil {
		return err
	}

	strat, err := NewMacStrategy(monitorChannel, config, &keys, observationsData, inst)
	if err != nil {
		return err
	}


	storage.StrategiesStorage[inst.ID] = strat
	storage.StrategiesStorage[inst.ID].Execute()

	return nil
}
