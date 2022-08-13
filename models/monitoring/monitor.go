package monitoring

import (
	"fmt"
	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"log"
	"newTradingBot/logs"
	block2 "newTradingBot/models/block"
	"strconv"
	"sync"
	"time"
)

type BinanceMonitor struct {
	symbol string
	timeFrameDuration time.Duration
	client *binance.Client
	pause bool
	stopSignal chan bool
	subscribers map[int]chan *block2.Data
	isFutures bool
	mtx sync.Mutex
}

func NewBinanceMonitor(symbol string, timeFrameDuration time.Duration, isFutures bool) *BinanceMonitor  {
	var binMonitor BinanceMonitor
	binMonitor.timeFrameDuration = timeFrameDuration
	binMonitor.symbol = symbol

	binMonitor.client = binance.NewClient("", "")

	binMonitor.stopSignal = make(chan bool)
	binMonitor.subscribers = make(map[int]chan *block2.Data)
	binMonitor.isFutures = isFutures

	return &binMonitor
}

// Subscribe - add subscriber to the monitor
func (m *BinanceMonitor) Subscribe(id int) chan *block2.Data {
	m.subscribers[id] = make(chan *block2.Data)
	logs.LogDebug(fmt.Sprintf("Instance #%d is SUBSCRIBED to BINANCE Monitor", id), nil)
	return m.subscribers[id]
}

// UnSubscribe - remove subscriber with given id
func (m *BinanceMonitor) UnSubscribe(id int) {
	delete(m.subscribers, id)
	logs.LogDebug(fmt.Sprintf("Instance #%d is UNSUBSCRIBED to BINANCE Monitor", id), nil)
}

// IsEmptySubs - returns current number of subscribers for monitor
func (m *BinanceMonitor) IsEmptySubs() bool {
	return len(m.subscribers) == 0
}

// NotifyAll - send data to all existent subscribers
func (m *BinanceMonitor) NotifyAll(marketData block2.Data)  {
	logs.LogDebug("Start notify all instances for binance monitor", nil)
	m.mtx.Lock()
	for _, channel := range m.subscribers {
		channel <- &marketData
	}
	m.mtx.Unlock()
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
				block := new(block2.Data)
				block.Symbol = m.symbol
				block.TradesArray = make([]float64, 0)
				block.Low = defaultMinPrice
				block.Time = m.timeFrameDuration

				logs.LogDebug("BINANCE monitor loop started", nil)

				stopC := make(chan struct{})

				if !m.isFutures {
					wsAggTradeHandler := func(event *binance.WsAggTradeEvent) {

						price, _ := strconv.ParseFloat(event.Price, 64)
						m.mtx.Lock()
						block.TradesArray = append(block.TradesArray, price)
						block.TradesCount++

						quantity, _ := strconv.ParseFloat(event.Quantity, 64)
						block.Volume += quantity

						if price > block.High {
							block.High = price
						}
						if price < block.Low {
							block.Low = price
						}

						m.mtx.Unlock()
					}

					_, stopC, _ = binance.WsAggTradeServe(m.symbol, wsAggTradeHandler, errHandlerLog)
				} else {
					wsAggTradeHandler := func(event *futures.WsAggTradeEvent) {
						m.mtx.Lock()
						price, _ := strconv.ParseFloat(event.Price, 64)
						block.TradesArray = append(block.TradesArray, price)
						block.TradesCount++

						quantity, _ := strconv.ParseFloat(event.Quantity, 64)
						block.Volume += quantity

						if price > block.High {
							block.High = price
						}
						if price < block.Low {
							block.Low = price
						}
						m.mtx.Unlock()
					}

					_, stopC, _ = futures.WsAggTradeServe(m.symbol, wsAggTradeHandler, errHandlerLog)
				}

				time.Sleep(m.timeFrameDuration)
				stopC <- struct{}{}

				m.mtx.Lock()
				if block.TradesCount > 0 {
					block.OpenPrice = block.TradesArray[0]
					block.ClosePrice = block.TradesArray[block.TradesCount-1]
					sum := 0.

					for i := range block.TradesArray {
						sum += block.TradesArray[i]
					}
					block.AveragePrice = sum/float64(block.TradesCount)
					m.mtx.Unlock()

					go m.NotifyAll(*block)
				} else {
					m.mtx.Unlock()
				}
				logs.LogDebug("BINANCE monitor loop finished", nil)
			}
		}
	}()
}

// simple error logger
func errHandlerLog(err error) {
	log.Print(err)
}