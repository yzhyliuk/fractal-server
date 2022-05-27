package block

import "time"

// Block - represents market info block
type Block struct {
	Symbol string

	TradesCount int
	Time   time.Duration

	Volume float64

	MaxPrice float64
	MinPrice float64

	EntryPrice float64
	ClosePrice float64

	AveragePrice float64

	// slice of all trades price for given time frame
	Trades  []float64
}
