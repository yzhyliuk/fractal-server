package monitoring

import (
	"fmt"
	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"log"
	"newTradingBot/logs"
	block2 "newTradingBot/models/block"
	"strconv"
	"time"
)

type BinanceMonitor struct {
	symbol string
	timeFrameDuration time.Duration
	client *binance.Client
	pause bool
	stopSignal chan bool
	subscribers map[int]chan *block2.Block
	isFutures bool
}

func NewBinanceMonitor(symbol string, timeFrameDuration time.Duration, isFutures bool) *BinanceMonitor  {
	var binMonitor BinanceMonitor
	binMonitor.timeFrameDuration = timeFrameDuration
	binMonitor.symbol = symbol

	binMonitor.client = binance.NewClient("", "")

	binMonitor.stopSignal = make(chan bool)
	binMonitor.subscribers = make(map[int]chan *block2.Block)
	binMonitor.isFutures = isFutures

	return &binMonitor
}

// Subscribe - add subscriber to the monitor
func (m *BinanceMonitor) Subscribe(id int) chan *block2.Block {
	m.subscribers[id] = make(chan *block2.Block)
	logs.LogDebug(fmt.Sprintf("Instance #%d is SUBSCRIBED to BINANCE Monitor", id), nil)
	return m.subscribers[id]
}

// UnSubscribe - remove subscriber with given id
func (m *BinanceMonitor) UnSubscribe(id int) {
	delete(m.subscribers, id)
	logs.LogDebug(fmt.Sprintf("Instance #%d is UNSUBSCRIBED to BINANCE Monitor", id), nil)
}


// NotifyAll - send data to all existent subscribers
func (m *BinanceMonitor) NotifyAll(marketData *block2.Block)  {
	logs.LogDebug("Start notify all instances for binance monitor", nil)
	for _, channel := range m.subscribers {
		channel <- marketData
	}
}

func (m *BinanceMonitor) IsFutures() bool  {
	return m.isFutures
}

// Stop - stops binance market monitor
func (m *BinanceMonitor) Stop()  {
	m.stopSignal <- true
	logs.LogDebug(fmt.Sprintf("Binance monitor is STOPPED for %s with %d min. time frame", m.symbol, int(m.timeFrameDuration/time.Minute)), nil)
}

// RunMonitor - starts monitoring loop
func (m *BinanceMonitor) RunMonitor()  {
	logs.LogDebug(fmt.Sprintf("Binance monitor is RUNNING for %s with %d min. time frame", m.symbol, int(m.timeFrameDuration/time.Minute)), nil)
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

				logs.LogDebug("BINANCE monitor loop started", nil)

				stopC := make(chan struct{})

				if !m.isFutures {
					wsAggTradeHandler := func(event *binance.WsAggTradeEvent) {
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

					_, stopC, _ = binance.WsAggTradeServe(m.symbol, wsAggTradeHandler, errHandlerLog)
				} else {
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

					_, stopC, _ = futures.WsAggTradeServe(m.symbol, wsAggTradeHandler, errHandlerLog)
				}

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

					go m.NotifyAll(block)

					logs.LogDebug("BINANCE monitor loop finished", nil)
				}
			}
		}
	}()
}

// simple error logger
func errHandlerLog(err error) {
	log.Print(err)
}