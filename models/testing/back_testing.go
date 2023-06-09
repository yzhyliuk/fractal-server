package testing

import (
	"newTradingBot/models/apimodels"
	"newTradingBot/models/trade"
)

const LiveTest = 2
const BackTest = 1
const Disable = 0

func GetProfitWinRateAndRoiForTrades(trades []*trade.Trade) (profit, winRate, roi, averageTradeLength float64) {
	profit = 0
	winRate = 0
	roi = 0
	winTradeCounter := 0
	totalTradeLength := 0
	averageTradeLength = 0.

	if len(trades) == 0 {
		return
	}

	for _, tr := range trades {
		profit += tr.Profit
		totalTradeLength += tr.LengthCounter
		if tr.Profit > 0 {
			winTradeCounter++
		}
	}

	roi = profit / (trades[0].USD / float64(*trades[0].Leverage))

	winRate = float64(winTradeCounter) / float64(len(trades))

	averageTradeLength = float64(totalTradeLength) / float64(len(trades))

	return
}

func GetMetricsForTrades(results []*apimodels.BackTestingResult) (*apimodels.MassTestingResult, error) {
	totalProfit := 0.
	totalWinRate := 0.
	totalRoi := 0.
	totalTrades := 0
	for _, result := range results {
		profit, winRate, roi, averageTradeLength := GetProfitWinRateAndRoiForTrades(result.Trades)
		result.AverageTradeLength = averageTradeLength
		result.TradesCount = len(result.Trades)
		result.WinRate = winRate
		result.Roi = roi
		result.Profit = profit

		totalProfit += profit
		totalWinRate += winRate
		totalRoi += roi
		totalTrades += result.TradesCount
	}

	return &apimodels.MassTestingResult{
		Results:      results,
		TotalProfit:  totalProfit,
		TotalTrades:  totalTrades,
		TotalWinRate: totalWinRate / float64(len(results)),
		TotalRoi:     totalRoi,
	}, nil
}
