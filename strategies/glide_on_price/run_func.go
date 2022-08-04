package glide_on_price

import (
	"encoding/json"
	"newTradingBot/models/strategy/actions"
	"newTradingBot/storage"
)

func RunGlideOnPrice(userID int, rawConfig []byte) error{
	var config GlideOnPriceConfig

	err := json.Unmarshal(rawConfig, &config)
	if err != nil {
		return err
	}

	inst, monitorChannel, keys, err := actions.PrepareStrategy(config.BaseStrategyConfig, userID, 2)
	if err != nil {
		return err
	}
	strat, err := NewGlideOnPriceStrategy(monitorChannel, config, keys, nil, inst)
	if err != nil {
		return err
	}


	storage.StrategiesStorage[inst.ID] = strat
	storage.StrategiesStorage[inst.ID].Execute()

	return nil
}
