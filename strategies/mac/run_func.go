package mac

import (
	"encoding/json"
	"newTradingBot/models/block"
	"newTradingBot/models/monitoring/replay"
	"newTradingBot/models/strategy/actions"
	"newTradingBot/models/strategy/instance"
	"newTradingBot/models/testing"
	"newTradingBot/models/trade"
	"newTradingBot/models/users"
	"newTradingBot/storage"
)

func RunMovingAverageCrossover(userID int, rawConfig []byte, test int, sessionID *int) ([]*trade.Trade, error){
	var config MovingAverageCrossoverConfig

	err := json.Unmarshal(rawConfig, &config)
	if err != nil {
		return nil, err
	}

	var inst *instance.StrategyInstance
	var monitorChannel chan *block.Data
	var keys *users.Keys

	var replayMonitor *replay.MonitorReplay

	if test == testing.Disable {
		inst, monitorChannel, keys, err = actions.PrepareStrategy(config.BaseStrategyConfig, userID, 1)
	} else if test == testing.BackTest{
		inst, replayMonitor, keys, err = actions.PrepareBackTesting(config.BaseStrategyConfig, *sessionID, userID, 1)
		monitorChannel = replayMonitor.OutputChannel
	} else if test == testing.LiveTest {

	}

	if err != nil {
		return nil, err
	}

	strat, err := NewMacStrategy(monitorChannel, config, keys, nil, inst)
	if err != nil {
		return nil, err
	}

	if test == testing.Disable {
		storage.StrategiesStorage[inst.ID] = strat
		storage.StrategiesStorage[inst.ID].Execute()
	} else if test == testing.BackTest {
		strat.Execute()
		err := replayMonitor.Start()
		if err != nil {
			return nil, err
		}
		trades := strat.GetTestingTrades()
		strat.Stop()

		return trades, nil
	} else  if test == testing.LiveTest {

	}

	return nil, nil
}

