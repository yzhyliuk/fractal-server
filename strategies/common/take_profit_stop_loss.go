package common

import (
	"github.com/adshao/go-binance/v2/futures"
	"newTradingBot/api/database"
	"newTradingBot/logs"
	"newTradingBot/models/block"
	"newTradingBot/models/notifications"
	"newTradingBot/models/strategy/actions"
	"newTradingBot/models/testing"
	"newTradingBot/models/trade"
	"strconv"
	"time"
)

const maxPriceDefault = 0.
const lowPriceDefault = 9999999999.

// HandleTPansSL - handles user defined TP and SL for strategy
func (s *Strategy) HandleTPansSL(marketData *block.Data) {
	if s.LastTrade == nil {
		return
	}

	price := marketData.ClosePrice

	profit := 0.
	roi := 0.

	switch s.LastTrade.FuturesSide {
	case futures.SideTypeBuy:
		profit = (s.LastTrade.Quantity * price) - s.LastTrade.USD
	case futures.SideTypeSell:
		profit = (s.LastTrade.Quantity * s.LastTrade.PriceOpen) - (s.LastTrade.Quantity * price)
	}

	fee := s.LastTrade.USD * trade.BinanceFuturesTakerFeeRate
	profit -= 2 * fee

	roi = profit / (s.LastTrade.USD / float64(*s.LastTrade.Leverage))

	// Trade stop loss
	if s.StrategyInstance.TradeStopLoss != 0 {
		if roi < s.StrategyInstance.TradeStopLoss*-1 {
			if s.StrategyInstance.Testing == testing.BackTest {
				s.TestingCloseTrade(marketData)
				return
			}
			s.CloseAllTrades()
		}
	}

	// trade take profit
	if s.StrategyInstance.TradeTakeProfit != 0 {
		if roi > s.StrategyInstance.TradeTakeProfit {
			if s.StrategyInstance.Testing == testing.BackTest {
				s.TestingCloseTrade(marketData)
				return
			}
			s.CloseAllTrades()
		}
	}
}

func (s *Strategy) MaxLossPerStrategyCondition() bool {
	if s.StrategyInstance.StopLoss != 0 && s.TotalProfit*-1 > s.StrategyInstance.StopLoss {
		db, _ := database.GetDataBaseConnection()
		go func() {
			err := actions.StopStrategy(db, s.StrategyInstance)
			if err != nil {
				logs.LogDebug("", err)
			}
		}()

		err := notifications.CreateUserNotification(db, s.StrategyInstance.UserID, notifications.Warning, notifications.StrategyStopLoss(s.StrategyInstance.Pair, s.TotalProfit))
		if err != nil {
			logs.LogError(err)
		}

		return true
	}

	return false
}

// HandleStrategyDefinedStopLoss - handles strategy defined stop loss prices
func (s *Strategy) HandleStrategyDefinedStopLoss(data *block.Data) {
	if s.LastTrade != nil {
		exitTargetUp := data.High
		exitTargetDown := data.Low

		takeProfitSell := false
		takeProfitBuy := false

		stopLossSell := false
		stopLossBuy := false

		if s.TakeProfitPrice != 0 {
			takeProfitSell = s.LastTrade.FuturesSide == futures.SideTypeSell && exitTargetDown < s.TakeProfitPrice
			takeProfitBuy = s.LastTrade.FuturesSide == futures.SideTypeBuy && exitTargetUp > s.TakeProfitPrice
		}

		if s.StopLossPrice != 0 {
			stopLossSell = s.LastTrade.FuturesSide == futures.SideTypeSell && exitTargetUp > s.StopLossPrice
			stopLossBuy = s.LastTrade.FuturesSide == futures.SideTypeBuy && exitTargetDown < s.StopLossPrice
		}

		if takeProfitSell || takeProfitBuy || stopLossSell || stopLossBuy {
			s.CloseAllTrades()
			return
		}
	}
}

// CalculateStopLossPrice - calculates user defined stop loss price
func (s *Strategy) CalculateStopLossPrice(priceCurrent float64, sell bool) float64 {
	sl := s.StrategyInstance.TradeStopLoss
	if sl == 0 {
		return 0
	}

	priceDelta := priceCurrent * (sl / float64(*s.StrategyInstance.Leverage))

	if sell {
		return priceCurrent + priceDelta
	} else {
		return priceCurrent - priceDelta
	}
}

// LivePriceMonitoring - monitors price in real time to handle take profit and stop loss
func (s *Strategy) LivePriceMonitoring() {
	go func() {
		for {
			if s.Stopped {
				break
			} else if s.LastTrade == nil {
				time.Sleep(10 * time.Second)
				continue
			}

			lowPrice := lowPriceDefault
			highPrice := maxPriceDefault

			wsAggTradeHandler := func(event *futures.WsAggTradeEvent) {
				price, _ := strconv.ParseFloat(event.Price, 64)
				if price < lowPrice {
					lowPrice = price
				}
				if price > highPrice {
					highPrice = price
				}
			}

			_, stopC, _ := futures.WsAggTradeServe(s.StrategyInstance.Pair, wsAggTradeHandler, nil)

			time.Sleep(time.Second * 5)

			stopC <- struct{}{}

			shouldClose := false

			if s.LastTrade != nil && (lowPrice != lowPriceDefault && highPrice != maxPriceDefault) {
				if s.StopLossPrice != 0 {
					switch s.LastTrade.FuturesSide {
					case futures.SideTypeSell:
						if highPrice > s.StopLossPrice {
							shouldClose = true
						}
					case futures.SideTypeBuy:
						if lowPrice < s.StopLossPrice {
							shouldClose = true
						}
					}
				}
				if s.TakeProfitPrice != 0 {
					switch s.LastTrade.FuturesSide {
					case futures.SideTypeSell:
						if lowPrice < s.TakeProfitPrice {
							shouldClose = true
						}
					case futures.SideTypeBuy:
						if highPrice > s.TakeProfitPrice {
							shouldClose = true
						}
					}
				}
			}

			if shouldClose {
				s.CloseAllTrades()
			}
		}
	}()
}
