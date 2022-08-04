package mac

import (
	"encoding/json"
	"newTradingBot/models/strategy/actions"
	"newTradingBot/storage"
)

func RunMovingAverageCrossover(userID int, rawConfig []byte) error{
	var config MovingAverageCrossoverConfig

	err := json.Unmarshal(rawConfig, &config)
	if err != nil {
		return err
	}

	inst, monitorChannel, keys, err := actions.PrepareStrategy(config.BaseStrategyConfig, userID, 1)
	if err != nil {
		return err
	}

	strat, err := NewMacStrategy(monitorChannel, config, keys, nil, inst)
	if err != nil {
		return err
	}


	storage.StrategiesStorage[inst.ID] = strat
	storage.StrategiesStorage[inst.ID].Execute()

	return nil
}
