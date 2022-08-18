package testing

import "newTradingBot/models/trade"

const LiveTest = 2
const BackTest = 1
const Disable = 0

func GetProfitWinRateAndRoiForTrades(trades []*trade.Trade) (profit, winRate, roi float64) {
	profit = 0
	winRate = 0
	roi = 0
	winTradeCounter := 0

	if len(trades) == 0 {
		return
	}

	for _, tr := range trades {
		profit += tr.Profit
		if tr.Profit > 0 {
			winTradeCounter++
		}
	}

	if trades[0].IsFutures {
		roi = profit/(trades[0].USD/float64(*trades[0].Leverage))
	}
	winRate = float64(winTradeCounter)/ float64(len(trades))

	return
}