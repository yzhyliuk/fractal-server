package common

import (
	"github.com/adshao/go-binance/v2/futures"
	"newTradingBot/logs"
	"newTradingBot/models/account"
	"newTradingBot/models/block"
	"newTradingBot/models/testing"
)

// CloseAllTrades - terminates all opened trades for current strategy
func (s *Strategy) CloseAllTrades() {
	if s.StrategyInstance.Testing == testing.BackTest {
		s.TestingCloseTrade(s.currentMarketData)
	}
	if s.LastTrade != nil {
		tradeClosed, err := s.Account.CloseFuturesPosition(s.LastTrade)
		if err != nil {
			logs.LogDebug("", err)
			return
		}
		s.TotalProfit += tradeClosed.Profit
		s.LastTrade = nil
	}
}

// HandleSell - opens new sell trade, closing all prev trades
func (s *Strategy) HandleSell(marketData *block.Data) error {
	return s.handleOpenTrade(marketData, futures.SideTypeSell)
}

// HandleBuy - opens new buy trade, closing all prev trades
func (s *Strategy) HandleBuy(marketData *block.Data) error {
	return s.handleOpenTrade(marketData, futures.SideTypeBuy)
}

func (s *Strategy) handleOpenTrade(marketData *block.Data, side futures.SideType) error {
	// there is no need to open a new trade with the same side
	if s.LastTrade != nil && s.LastTrade.FuturesSide == side {
		return nil
	}
	// handle back testing
	quantity := account.QuantityFromPrice(s.StrategyInstance.Bid, marketData.ClosePrice)
	if s.StrategyInstance.Testing == testing.BackTest {
		if side == futures.SideTypeSell {
			s.TestingSell(marketData, quantity)
		} else {
			s.TestingBuy(marketData, quantity)
		}
		return nil
	}

	// close previous trade if it's exists
	if s.LastTrade != nil {
		s.CloseAllTrades()
		if marketData == nil {
			return nil
		}
	}

	// Open new trade
	futuresTrade, err := s.Account.OpenFuturesPosition(quantity, s.StrategyInstance.Pair, side, s.StrategyInstance)
	if err != nil {
		return err
	}

	s.StopLossPrice = s.CalculateStopLossPrice(futuresTrade.PriceOpen, side == futures.SideTypeSell)

	s.LastTrade = futuresTrade
	return nil
}

// CalculateTradeData - calculates max and min price values for given trade
func (s *Strategy) CalculateTradeData(marketData *block.Data) {
	if s.LastTrade != nil {
		s.LastTrade.LengthCounter++
		profit := 0.
		if s.LastTrade.FuturesSide == futures.SideTypeBuy {
			profit = (s.LastTrade.Quantity * marketData.Low) - s.LastTrade.USD
		} else {
			profit = (s.LastTrade.Quantity * s.LastTrade.PriceOpen) - (s.LastTrade.Quantity * marketData.High)

		}

		roi := profit / (s.LastTrade.USD / float64(*s.LastTrade.Leverage))
		if s.LastTrade.MaxDropDown == 0 || s.LastTrade.MaxDropDown > roi {
			s.LastTrade.MaxDropDown = roi
		}

		if s.LastTrade.MaxHeadUp == 0 || s.LastTrade.MaxHeadUp < roi {
			s.LastTrade.MaxHeadUp = roi
		}
	}
}
