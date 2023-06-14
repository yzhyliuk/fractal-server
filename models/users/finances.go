package users

import (
	"gorm.io/gorm"
	"newTradingBot/models/account"
	"newTradingBot/models/trade"
	"time"
)

type DayProfit struct {
	Date   string  `json:"label"`
	Profit float64 `json:"value"`
}

type UserFinances struct {
	UserBalance *account.UserBalances `json:"balance"`
	DayProfit   float64               `json:"dayProfit"`
	WeekProfit  []*DayProfit          `json:"weekProfit"`
}

func GetUserFinances(db *gorm.DB, userID int) (*UserFinances, error) {
	keys, err := GetUserKeys(db, userID)
	if err != nil {
		return nil, err
	}

	uFinances := &UserFinances{}

	uFinances.UserBalance, err = getUserBalance(keys)
	if err != nil {
		return nil, err
	}

	lastDayTime := time.Now().Add(-time.Duration(time.Now().Hour()) * time.Hour).Add(-time.Duration(time.Now().Minute()) * time.Minute)
	dailyProfit, err := getDailyProfit(db, userID, &lastDayTime, nil)
	if err != nil {
		return nil, err
	}

	uFinances.DayProfit = dailyProfit

	weekDailyProfits, err := getProfit(db, userID, 90)
	uFinances.WeekProfit = weekDailyProfits

	return uFinances, nil
}

func getUserBalance(keys Keys) (*account.UserBalances, error) {
	acc, err := account.NewBinanceAccount(keys.ApiKey, keys.SecretKey, keys.ApiKey, keys.SecretKey)
	if err != nil {
		return nil, err
	}

	return acc.GetStableBalance()
}

func getDailyProfit(db *gorm.DB, userID int, from *time.Time, to *time.Time) (profit float64, err error) {

	var trades []*trade.Trade

	db = db.Where("user_id = ?", userID)
	if from != nil {
		db = db.Where("time_stamp > ?", from)
	}

	if to != nil {
		db = db.Where("time_stamp < ?", to)
	}

	err = db.Find(&trades).Error
	if err != nil {
		return 0, err
	}

	profit = 0

	for i := range trades {
		profit += trades[i].Profit
	}

	return
}

func getProfit(db *gorm.DB, userID, days int) ([]*DayProfit, error) {

	from := time.Now().Add(-time.Duration(time.Now().Hour()) * time.Hour).Add(-time.Duration(time.Now().Minute()) * time.Minute)
	to := from.Add(24 * time.Hour)

	dailyProfits := make([]*DayProfit, 0)

	var trades []*trade.Trade
	err := db.Where("user_id = ? and time_stamp > ?", userID, from.Add(-time.Duration(days)*24*time.Hour)).Find(&trades).Error
	if err != nil {
		return nil, err
	}

	for i := 1; i < days; i++ {
		from = from.Add(-24 * time.Hour)
		to = to.Add(-24 * time.Hour)

		profit := 0.

		for j := range trades {
			if trades[j].TimeStamp.After(from) && trades[j].TimeStamp.Before(to) {
				profit += trades[j].Profit
			}
		}

		dailyProfits = append(dailyProfits, &DayProfit{
			Profit: profit,
			Date:   from.Format("02-01-2006"),
		})
	}

	return dailyProfits, nil
}
