package nadaraya_watsons

import (
	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"newTradingBot/logs"
	"newTradingBot/models/account"
	"newTradingBot/models/block"
)

// buy - preforms buy
func (m *mac) buy(marketData *block.Block) error {
	quantity := account.QuantityFromPrice(m.config.BidSize, marketData.ClosePrice)

	if m.strategyInstance.IsFutures {
		if m.lastTrade != nil {
			m.closePreviousTrade()
		}
		futuresTrade, err := m.account.OpenFuturesPosition(quantity,m.config.Pair, futures.SideTypeBuy, m.GetInstance())
		if err != nil {
			return err
		}

		m.lastTrade = futuresTrade
	} else {
		spotTrade, err := m.account.PlaceMarketOrder(quantity,m.config.Pair,binance.SideTypeBuy, m.GetInstance(), m.lastTrade)
		if err != nil {
			return err
		}

		m.lastTrade = spotTrade
	}

	return nil
}

func (m *mac) sell(marketData *block.Block) error {
	if m.strategyInstance.IsFutures {
		if m.lastTrade != nil {
			m.closePreviousTrade()
			if marketData == nil {
				return nil
			}
		}
		quantity := account.QuantityFromPrice(m.config.BidSize, marketData.ClosePrice)
		futuresTrade, err := m.account.OpenFuturesPosition(quantity,m.config.Pair, futures.SideTypeSell, m.GetInstance())
		if err != nil {
			return err
		}

		m.lastTrade = futuresTrade
	} else {
		if m.lastTrade == nil {
			return nil
		}
		_, err := m.account.PlaceMarketOrder(m.lastTrade.Quantity, m.config.Pair, binance.SideTypeSell, m.GetInstance(), m.lastTrade)
		if err != nil {
			return err
		}

		m.lastTrade = nil
	}
	return nil
}

func (m *mac) closePreviousTrade()  {
	_, err := m.account.CloseFuturesPosition(m.lastTrade)
	if err != nil {
		logs.LogDebug("", err)
		return
	}
	m.lastTrade = nil
}

