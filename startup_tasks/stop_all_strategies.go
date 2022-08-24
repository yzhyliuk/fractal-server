package startup_tasks

import (
	"gorm.io/gorm"
	"newTradingBot/api/helpers"
	"newTradingBot/logs"
	"newTradingBot/models/account"
	"newTradingBot/models/strategy/instance"
	"newTradingBot/models/trade"
	"newTradingBot/models/users"
	"strconv"
)

func StopAllStrategies(db *gorm.DB) error {
	return db.Table(instance.StrategyInstancesTableName).Where("status != ?", helpers.Stopped).Update("status", helpers.Stopped).Error
}

func CloseAllTrades(db *gorm.DB) error {
	trades := make([]trade.Trade, 0)

	err := db.Where("status = ?", trade.StatusActive).Find(&trades).Error
	if err != nil {
		logs.LogError(err)
		return err
	}

	for _, t := range trades {
		userKeys, err := users.GetUserKeys(db, t.UserID)
		if err != nil {
			logs.LogError(err)
			continue
		}

		acc, err := account.NewBinanceAccount(userKeys.ApiKey, userKeys.SecretKey,userKeys.ApiKey, userKeys.SecretKey)

		if t.IsFutures {
			res, err := acc.GetOpenedFuturesTrades(t.Pair, t.TimeStamp)
			if err != nil {
				logs.LogError(err)
				continue
			}

			for _, order := range res {
				entryPrice, err := strconv.ParseFloat(order.EntryPrice, 64)
				if err != nil {
					continue
				}

				if t.PriceOpen == entryPrice {
					_, err := acc.CloseFuturesPosition(&t)
					if err != nil {
						// TODO : notify user about close
						logs.LogError(err)
						continue
					}
				}
			}
		}

	}

	return nil
}