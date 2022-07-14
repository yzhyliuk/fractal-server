package configs

type BaseStrategyConfig struct {
	StrategyID int `json:"strategyId"`
	Pair string `json:"pair"`
	BidSize float64 `json:"bid"`
	TimeFrame int `json:"timeFrame"`
	IsFutures bool `json:"isFutures"`
	Leverage *int `json:"leverage"`
}

