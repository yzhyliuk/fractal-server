package trend_with_rsi

import (
	"encoding/json"
	"newTradingBot/models/strategy/actions"
	"newTradingBot/storage"
)

func RunTrendWithRSI(userID int, rawConfig []byte) error{
	var config TrendWithRSIConfig

	err := json.Unmarshal(rawConfig, &config)
	if err != nil {
		return err
	}

	inst, monitorChannel, keys, err := actions.PrepareStrategy(config.BaseStrategyConfig, userID,3)
	if err != nil {
		return err
	}

	strat, err := NewTrendFollowWithRSIStrategy(monitorChannel, config, keys, nil, inst)
	if err != nil {
		return err
	}


	storage.StrategiesStorage[inst.ID] = strat
	storage.StrategiesStorage[inst.ID].Execute()

	return nil
}

