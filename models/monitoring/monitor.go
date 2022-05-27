package monitoring

import (
	"errors"
	"fmt"
	"github.com/adshao/go-binance/v2"
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

	historicalData []*block2.Block
}

func NewBinanceMonitor(symbol string, timeFrameDuration time.Duration) *BinanceMonitor  {
	var binMonitor BinanceMonitor
	binMonitor.timeFrameDuration = timeFrameDuration
	binMonitor.symbol = symbol

	binMonitor.client = binance.NewClient("", "")

	binMonitor.stopSignal = make(chan bool)
	binMonitor.subscribers = make(map[int]chan *block2.Block)


	// Save data for last 24 Hours
	binMonitor.historicalData = make([]*block2.Block, int(24*time.Hour)/int(timeFrameDuration))

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

	m.historicalData = m.historicalData[1:]
	m.historicalData = append(m.historicalData, marketData)
}

// Stop - stops binance market monitor
func (m *BinanceMonitor) Stop()  {
	m.stopSignal <- true
	logs.LogDebug(fmt.Sprintf("Binance monitor is STOPPED for %s with %d min. time frame", m.symbol, int(m.timeFrameDuration/time.Minute)), nil)
}

// GetHistoricalData - returns historical data for given timeframes count
func (m *BinanceMonitor) GetHistoricalData(timeFrames int) ([]*block2.Block, error) {
	length := len(m.historicalData)
	if timeFrames > length {
		return  nil, errors.New("monitor doesn't contains enough timeframes")
	}
	return m.historicalData[length-timeFrames-1:], nil
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

				_, stopC, _ := binance.WsAggTradeServe(m.symbol, wsAggTradeHandler, errHandlerLog)

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