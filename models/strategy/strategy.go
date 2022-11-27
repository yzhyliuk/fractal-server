package strategy

import (
	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"newTradingBot/models/block"
	"newTradingBot/models/strategy/instance"
	"newTradingBot/models/trade"
)

type Strategy interface {
	Execute()
	GetInstance() *instance.StrategyInstance
	Stop()
	ExecuteExperimental()
	GetTestingTrades() []*trade.Trade
	ChangeBid(bid float64) error
}

type HandlerFunc func(block *block.Data, orderSpot *binance.CreateOrderResponse, orderFutures *futures.CreateOrderResponse, err error)

type Settings struct {
	BaseBid float64
	Symbol string
}