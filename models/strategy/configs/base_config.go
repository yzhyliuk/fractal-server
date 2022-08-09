package configs

type BaseStrategyConfig struct {
	StrategyID int `json:"strategyId"`
	Pair string `json:"pair"`
	BidSize float64 `json:"bid"`
	TimeFrame int `json:"timeFrame"`
	IsFutures bool `json:"isFutures"`
	Leverage *int `json:"leverage"`
	StopLoss float64 `json:"stopLoss"`
	TradeTakeProfit float64 `json:"tradeTakeProfit"`
	TradeStopLoss float64 `json:"tradeStopLoss"`
}

