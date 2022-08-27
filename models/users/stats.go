package users

import (
	"gorm.io/gorm"
	"newTradingBot/models/trade"
)

type UserStats struct {
	User
	TradesClosed int `json:"trades_closed"`
	TotalVolumeMargin float64 `json:"total_volume_margin"`
	TotalVolume           float64 `json:"total_volume"`
	TotalProfit           float64 `json:"total_profit"`
	AverageProfitPerTrade float64 `json:"average_profit_per_trade"`
}

func GetUserStats(db *gorm.DB, userID int) (*UserStats, error) {
	var user User
	err := db.Where("id = ?", userID).Find(&user).Error
	if err != nil {
		return nil, err
	}

	var trades []*trade.Trade

	err = db.Where("user_id = ?", userID).Find(&trades).Error
	if err != nil {
		return nil, err
	}

	stats := &UserStats{
		TradesClosed: len(trades),
	}

	totalProfit, volume, volumeMargin := TotalProfitAndVolume(trades)

	stats.User = user
	stats.TotalProfit = totalProfit
	stats.TotalVolume = volume
	stats.TotalVolumeMargin = volumeMargin
	stats.AverageProfitPerTrade = totalProfit/float64(stats.TradesClosed)

	return stats, nil

}

func TotalProfitAndVolume(trades []*trade.Trade) (profit, volume, volumeMargin float64) {
	profit = 0.
	volume = 0.
	volumeMargin = 0.
	for _, t := range trades {
		profit += t.Profit
		if t.IsFutures && t.Leverage != nil {
			volume += t.USD / float64(*(t.Leverage))
			volumeMargin += t.USD
		} else {
			volume += t.USD
		}
	}

	return
}
