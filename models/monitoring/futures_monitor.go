package monitoring

import (
	"github.com/adshao/go-binance/v2/futures"
	block2 "newTradingBot/models/block"
	"newTradingBot/models/strategy"
	"strconv"
	"time"
)

type BinanceFuturesMonitor struct {
	symbol string
	timeFrameDuration time.Duration
	client *futures.Client

	pause bool

	stopSignal chan bool
}

func NewBinanceFuturesMonitor(symbol, apiKey, secretKey string, timeFrameDuration time.Duration) *BinanceFuturesMonitor  {
	var binMonitor BinanceFuturesMonitor
	binMonitor.timeFrameDuration = timeFrameDuration
	binMonitor.symbol = symbol

	binMonitor.client = futures.NewClient(apiKey,secretKey)

	binMonitor.stopSignal = make(chan bool)
	binMonitor.pause = false

	return &binMonitor
}

func (m *BinanceFuturesMonitor) RunWithStrategy(strategy strategy.Strategy)  {
	go func() {
		for {
			select {
			case <-m.stopSignal:
				return
			default:
				block := new(block2.Block)
				block.Symbol = m.symbol
				block.Trades = make([]float64, 0)
				block.MinPrice = defaultMinPrice
				block.Time = m.timeFrameDuration

				wsAggTradeHandler := func(event *futures.WsAggTradeEvent) {
					price, _ := strconv.ParseFloat(event.Price, 64)
					block.Trades = append(block.Trades, price)
					block.TradesCount++

					quantity, _ := strconv.ParseFloat(event.Quantity, 64)
					block.Volume += quantity

					if price > block.MaxPrice {
						block.MaxPrice = price
					}
					if price < block.MinPrice {
						block.MinPrice = price
					}
				}

				_, stopC, _ := futures.WsAggTradeServe(m.symbol, wsAggTradeHandler, errHandlerLog)

				time.Sleep(m.timeFrameDuration)
				stopC <- struct{}{}

				if block.TradesCount > 0 {
					block.EntryPrice = block.Trades[0]
					block.ClosePrice = block.Trades[block.TradesCount-1]
					sum := 0.

					for i := range block.Trades {
						sum += block.Trades[i]
					}
					block.AveragePrice = sum/float64(block.TradesCount)

					go strategy.Execute()
				}
			}
		}
	}()
}


// Stop - stops binance market monitor
func (m *BinanceFuturesMonitor) Stop() {
	m.stopSignal <- true
}

// Pause - stops binance market monitor
func (m *BinanceFuturesMonitor) Pause() {
	m.stopSignal <- true
}

// Resume - stops binance market monitor
func (m *BinanceFuturesMonitor) Resume() {
	m.stopSignal <- true
}
