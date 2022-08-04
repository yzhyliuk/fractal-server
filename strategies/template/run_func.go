package template

import (
	"encoding/json"
	"newTradingBot/models/strategy/actions"
	"newTradingBot/storage"
)

func RunTemplateStrategy(userID int, rawConfig []byte) error{
	var config StrategyTemplateConfig

	err := json.Unmarshal(rawConfig, &config)
	if err != nil {
		return err
	}

	inst, monitorChannel, keys, err := actions.PrepareStrategy(config.BaseStrategyConfig, userID)
	if err != nil {
		return err
	}

	strat, err := NewTemplateStrategy(monitorChannel, config, keys, nil, inst)
	if err != nil {
		return err
	}


	storage.StrategiesStorage[inst.ID] = strat
	storage.StrategiesStorage[inst.ID].Execute()

	return nil
}
