package common

import (
	"github.com/adshao/go-binance/v2/futures"
	"newTradingBot/models/block"
	"newTradingBot/models/trade"
)

// TestingBuy - back testing buy function
func (s *Strategy) TestingBuy(marketData *block.Data, quantity float64) {
	newTrade := &trade.Trade{
		Quantity:      quantity,
		USD:           s.StrategyInstance.Bid,
		PriceOpen:     marketData.ClosePrice,
		FuturesSide:   futures.SideTypeBuy,
		LengthCounter: 0,
	}

	if s.LastTrade != nil && s.LastTrade.FuturesSide == futures.SideTypeBuy {
		return
	}

	if s.LastTrade != nil && s.LastTrade.FuturesSide == futures.SideTypeSell {
		s.TestingCloseTrade(marketData)

		newTrade.Leverage = s.StrategyInstance.Leverage
		s.LastTrade = newTrade
	}
	if s.LastTrade == nil {
		newTrade.Leverage = s.StrategyInstance.Leverage
		s.LastTrade = newTrade
	}

}

// TestingSell - back testing sell function
func (s *Strategy) TestingSell(marketData *block.Data, quantity float64) {
	newTrade := &trade.Trade{
		Quantity:      quantity,
		USD:           s.StrategyInstance.Bid,
		PriceOpen:     marketData.ClosePrice,
		FuturesSide:   futures.SideTypeSell,
		LengthCounter: 0,
	}
	if s.LastTrade != nil && s.LastTrade.FuturesSide == futures.SideTypeSell {
		return
	}

	if s.LastTrade != nil && s.LastTrade.FuturesSide == futures.SideTypeBuy {
		s.TestingCloseTrade(marketData)
	}

	if s.LastTrade == nil {
		newTrade.Leverage = s.StrategyInstance.Leverage
		s.LastTrade = newTrade
	}
}

// TestingCloseTrade - closes trade for backtesting
func (s *Strategy) TestingCloseTrade(marketData *block.Data) {
	if s.LastTrade != nil {
		closedTrade := s.LastTrade
		closedTrade.PriceClose = marketData.ClosePrice
		closedTrade.CalculateProfitRoi()
		s.trades = append(s.trades, closedTrade)
		s.LastTrade = nil
	}
}
