package glide_on_price

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

func RunGlideOnPrice(userID int, rawConfig []byte) error{
	var config GlideOnPriceConfig

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
		IsFutures: config.IsFutures,
		Status:     helpers.Created,
	}

	if inst.IsFutures {
		inst.Leverage = config.Leverage
	}

	db, err := database.GetDataBaseConnection()
	if err != nil {
		return err
	}

	inst, err = instance.CreateStrategyInstance(db, inst)
	if err != nil {
		return err
	}

	monitorName := fmt.Sprintf("%s:%d:%t",config.Pair, config.TimeFrame, inst.IsFutures)

	var monitorChannel chan *block.Block

	var observationsData []*block.Block

	if storage.MonitorsBinance[monitorName] != nil{
		monitorChannel = storage.MonitorsBinance[monitorName].Subscribe(inst.ID)
	} else {
		monitor := monitoring.NewBinanceMonitor(config.Pair, time.Duration(config.TimeFrame*int(time.Second)),inst.IsFutures)
		storage.MonitorsBinance[monitorName] = monitor
		storage.MonitorsBinance[monitorName].RunMonitor()
		monitorChannel = storage.MonitorsBinance[monitorName].Subscribe(inst.ID)
		observationsData = make([]*block.Block, config.VolatilityObservationsTimeFrame)
	}

	keys, err := users.GetUserKeys(db, userID)
	if err != nil {
		return err
	}

	strat, err := NewGlideOnPriceStrategy(monitorChannel, config, &keys, observationsData, inst)
	if err != nil {
		return err
	}


	storage.StrategiesStorage[inst.ID] = strat
	storage.StrategiesStorage[inst.ID].Execute()

	return nil
}
