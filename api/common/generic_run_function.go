package common

import (
	"encoding/json"
	"newTradingBot/models/block"
	"newTradingBot/models/monitoring/replay"
	"newTradingBot/models/strategy"
	"newTradingBot/models/strategy/actions"
	"newTradingBot/models/strategy/configs"
	"newTradingBot/models/strategy/instance"
	"newTradingBot/models/testing"
	"newTradingBot/models/trade"
	"newTradingBot/models/users"
	"newTradingBot/storage"
	"newTradingBot/strategies/bollinger_bands_with_atr"
	"newTradingBot/strategies/price_channel_breakout"
	"newTradingBot/strategies/regression_channels"
)

var NewStrategy = map[string]func(monitorChannel chan *block.Data, configRaw []byte, keys *users.Keys, historicalData []*block.Data, inst *instance.StrategyInstance) (strategy.Strategy, error){
	regression_channels.StrategyName:      regression_channels.NewLinearRegression,
	price_channel_breakout.StrategyName:   price_channel_breakout.NewPriceChannelBreakoutStrategy,
	bollinger_bands_with_atr.StrategyName: bollinger_bands_with_atr.NewBollingerBandsWithATR,
}

func RunFunction(userID int, rawConfig []byte, test, strategyID int, strategyName string, sessionID *int) ([]*trade.Trade, *int, error) {
	var genericConfig configs.GenericConfig

	err := json.Unmarshal(rawConfig, &genericConfig)
	if err != nil {
		return nil, nil, err
	}

	var inst *instance.StrategyInstance
	var monitorChannel chan *block.Data
	var keys *users.Keys

	var replayMonitor *replay.MonitorReplay

	if test == testing.Disable {
		inst, monitorChannel, keys, err = actions.PrepareStrategy(genericConfig.BaseStrategyConfig, userID, strategyID)
	} else if test == testing.BackTest {
		inst, replayMonitor, keys, err = actions.PrepareBackTesting(genericConfig.BaseStrategyConfig, *sessionID, userID, strategyID)
		monitorChannel = replayMonitor.OutputChannel
	} else if test == testing.LiveTest {

	}

	if err != nil {
		return nil, nil, err
	}

	strat, err := NewStrategy[strategyName](monitorChannel, rawConfig, keys, nil, inst)
	if err != nil {
		return nil, nil, err
	}
	if test == testing.Disable {
		storage.StrategiesStorage[inst.ID] = strat
		storage.StrategiesStorage[inst.ID].Execute()
		return nil, &inst.ID, nil
	} else if test == testing.BackTest {
		strat.Execute()
		err := replayMonitor.Start()
		if err != nil {
			return nil, nil, err
		}
		trades := strat.GetTestingTrades()
		strat.Stop()

		return trades, nil, nil
	} else if test == testing.LiveTest {

	}

	return nil, nil, nil
}
