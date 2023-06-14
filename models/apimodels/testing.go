package apimodels

import "newTradingBot/models/trade"

type BackTestModel struct {
	Config     interface{} `json:"config"`
	StrategyID int         `json:"strategyId"`
	CaptureID  int         `json:"sessionId"`
}

type MassBackTesting struct {
	Config     interface{} `json:"config"`
	Pair       string      `json:"pair"`
	StrategyID int         `json:"strategyId"`
	TimeFrame  int         `json:"timeframe"`
}

type BackTestingResult struct {
	Trades                []*trade.Trade `json:"trades"`
	Roi                   float64        `json:"roi"`
	WinRate               float64        `json:"winRate"`
	AverageTradeLength    float64        `json:"averageTradeLength"`
	TradesCount           int            `json:"tradesCount"`
	AverageProfitPerTrade float64        `json:"averageProfitPerTrade"`
	TimeFrame             int            `json:"timeFrame"`
	Pair                  string         `json:"pair"`
	Profit                float64        `json:"profit"`
	MaxCumulativeProfit   float64        `json:"maxCumulativeProfit"`
	MaxCumulativeLoss     float64        `json:"maxCumulativeLoss"`
}

type MassTestingResult struct {
	Results      []*BackTestingResult `json:"results"`
	TotalProfit  float64              `json:"totalProfit"`
	TotalWinRate float64              `json:"totalWinRate"`
	TotalTrades  int                  `json:"totalTrades"`
	TotalRoi     float64              `json:"totalRoi"`
}
