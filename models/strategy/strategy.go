package strategy

import (
	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"newTradingBot/models/block"
	"newTradingBot/models/strategy/instance"
)

type Strategy interface {
	Execute()
	GetInstance() *instance.StrategyInstance
	Stop()
}

type HandlerFunc func(block *block.Block, orderSpot *binance.CreateOrderResponse, orderFutures *futures.CreateOrderResponse, err error)

type Settings struct {
	BaseBid float64
	Symbol string
}