package rsi_crossover


import (
	"encoding/json"
	"newTradingBot/models/strategy/actions"
	"newTradingBot/storage"
)

func RunRSICrossoverStrategy(userID int, rawConfig []byte) error{
	var config RSICrossoverConfig

	err := json.Unmarshal(rawConfig, &config)
	if err != nil {
		return err
	}

	inst, monitorChannel, keys, err := actions.PrepareStrategy(config.BaseStrategyConfig, userID, 4)
	if err != nil {
		return err
	}

	strat, err := NewRSICrossoverStrategy(monitorChannel, config, keys, nil, inst)
	if err != nil {
		return err
	}


	storage.StrategiesStorage[inst.ID] = strat
	storage.StrategiesStorage[inst.ID].Execute()

	return nil
}

