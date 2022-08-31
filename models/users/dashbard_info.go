package users

import (
	"context"
	"github.com/adshao/go-binance/v2/futures"
	"gorm.io/gorm"
	"newTradingBot/models/trade"
	"strconv"
	"time"
)

type UserDashBoard struct {
	ActiveTrades  []*TradeInfo `json:"activeTrades"`
	TodayProfit   float64        `json:"todayProfit"`
}

type TradeInfo struct {
	*trade.Trade
	UnrealizedPNL float64 `json:"unrealizedPNL"`
}

func GetUserDashboardInfo(db *gorm.DB, userID int) (*UserDashBoard, error) {
	uDashboard := &UserDashBoard{}

	activeTrades := make([]*trade.Trade,0)
	activeTradesInfo := make([]*TradeInfo, 0)

	err := db.Where("status = 'active' AND user_id = ?",userID).Find(&activeTrades).Error
	if err != nil {
		return nil, err
	}

	userKeys, err := GetUserKeys(db, userID)
	if err != nil {
		return nil, err
	}

	client := futures.NewClient(userKeys.ApiKey, userKeys.SecretKey)

	for _, t := range activeTrades {
		resp, err := client.NewGetPositionRiskService().Symbol(t.Pair).Do(context.Background())
		if err != nil {
			return nil, err
		}

		pnl, _ := strconv.ParseFloat(resp[0].UnRealizedProfit, 64)

		activeTradesInfo = append(activeTradesInfo, &TradeInfo{
			Trade: t,
			UnrealizedPNL: pnl,
		})
	}

	uDashboard.ActiveTrades = activeTradesInfo

	from := time.Now().Add(-time.Duration(time.Now().Hour())*time.Hour).Add(-time.Duration(time.Now().Minute())*time.Minute)
	to := from.Add(24*time.Hour)

	profit, err := getDailyProfit(db,userID,&from,&to)

	uDashboard.TodayProfit = profit

	return uDashboard, err
}