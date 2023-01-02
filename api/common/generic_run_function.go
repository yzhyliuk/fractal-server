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
	"newTradingBot/strategies/glide_on_price"
	"newTradingBot/strategies/linear_regression"
	"newTradingBot/strategies/mac"
	"newTradingBot/strategies/mean_reversion"
	qqe "newTradingBot/strategies/qqe_strategy"
	"newTradingBot/strategies/regression_channels"
	"newTradingBot/strategies/rsi_crossover"
	"newTradingBot/strategies/simple_rsi"
	"newTradingBot/strategies/trend_with_rsi"
)

var NewStrategy = map[string]func(monitorChannel chan *block.Data, configRaw []byte, keys *users.Keys, historicalData []*block.Data, inst *instance.StrategyInstance) (strategy.Strategy, error) {
	simple_rsi.StrategyName: simple_rsi.NewSimpleRSI,
	glide_on_price.StrategyName: glide_on_price.NewGlideOnPriceStrategy,
	linear_regression.StrategyName: linear_regression.NewLinearRegression,
	mac.StrategyName: mac.NewMacStrategy,
	mean_reversion.StrategyName: mean_reversion.NewMeanReversion,
	qqe.StrategyName: qqe.NewQQEStrategy,
	regression_channels.StrategyName: regression_channels.NewLinearRegression,
	rsi_crossover.StrategyName: rsi_crossover.NewRSICrossoverStrategy,
	trend_with_rsi.StrategyName: trend_with_rsi.NewTrendFollowWithRSIStrategy,
}


func RunFunction(userID int, rawConfig []byte, test, strategyID int, strategyName string, sessionID *int) ([]*trade.Trade, *int, error) {
	var genericConfig configs.GenericConfig

	err := json.Unmarshal(rawConfig, &genericConfig)
	if err != nil {
		return nil, nil,err
	}

	var inst *instance.StrategyInstance
	var monitorChannel chan *block.Data
	var keys *users.Keys

	var replayMonitor *replay.MonitorReplay

	if test == testing.Disable {
		inst, monitorChannel, keys, err = actions.PrepareStrategy(genericConfig.BaseStrategyConfig, userID, strategyID)
	} else if test == testing.BackTest{
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
	} else  if test == testing.LiveTest {

	}

	return nil, nil, nil
}

