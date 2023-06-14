package common

import (
	"newTradingBot/indicators"
	"newTradingBot/models/block"
)

// ToHeikinAshi - turns market data to HeikinAshi candlesticks data
func (s *Strategy) ToHeikinAshi(marketData *block.Data) *block.Data {
	closeP := (marketData.ClosePrice + marketData.OpenPrice + marketData.High + marketData.Low) / float64(4)
	open := marketData.OpenPrice
	if s.prevMarketData != nil {
		open = (s.prevMarketData.OpenPrice + s.prevMarketData.ClosePrice) / float64(2)
	}
	high := indicators.Max([]float64{marketData.High, marketData.OpenPrice, marketData.ClosePrice})
	low := indicators.Min([]float64{marketData.Low, marketData.OpenPrice, marketData.ClosePrice})

	marketData.ClosePrice = closeP
	marketData.OpenPrice = open
	marketData.High = high
	marketData.Low = low

	s.prevMarketData = marketData

	return marketData
}
