package experimental

import (
	"encoding/json"
	"newTradingBot/models/strategy/actions"
	"newTradingBot/storage"
)

func RunExperimentalContinuousStrategy(userID int, rawConfig []byte) error{
	var config continuousConfig

	err := json.Unmarshal(rawConfig, &config)
	if err != nil {
		return err
	}

	inst, keys, err := actions.PrepareExperimentalStrategy(config.BaseStrategyConfig, userID,5)
	if err != nil {
		return err
	}

	strat, err := NewContinuousExperimentalStrategy(config, keys, inst)
	if err != nil {
		return err
	}


	storage.StrategiesStorage[inst.ID] = strat
	storage.StrategiesStorage[inst.ID].ExecuteExperimental()

	return nil
}
