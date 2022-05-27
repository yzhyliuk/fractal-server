package account

import "github.com/adshao/go-binance/v2/futures"

type InternalFuturesOrder struct {
	Side futures.SideType
	StopLoss float64
	TakeProfit float64
}
