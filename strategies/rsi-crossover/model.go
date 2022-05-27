package rsi_crossover

import (
	"fmt"
	"github.com/adshao/go-binance/v2"
	"newTradingBot/indicators"
	"newTradingBot/models/account"
	"newTradingBot/models/block"
	"newTradingBot/models/strategy"
	"time"
)

type rsiCrossOver struct {

	observations []float64
	observationsLength int

	timeFrame time.Duration

	handler strategy.HandlerFunc
	account account.Account
	settings *strategy.Settings

	lastOrder *account.InternalFuturesOrder

	rsiObservations []float64


	lastEMA float64
	lastOrderAmount float64

	onTrade bool
}

func NewRSICrossover(handler strategy.HandlerFunc, account account.Account,
	settings *strategy.Settings, timeFrame time.Duration, monitorTimeFrame time.Duration) strategy.Strategy {

	rsiCrossover := &rsiCrossOver{
		handler: handler,
		account: account,
		settings: settings,
		observationsLength: int((timeFrame*15)/monitorTimeFrame),
		lastEMA: 0,
		onTrade: false,
	}

	rsiCrossover.timeFrame = timeFrame

	rsiCrossover.observations = make([]float64, rsiCrossover.observationsLength)
	rsiCrossover.rsiObservations = make([]float64,rsiCrossover.observationsLength)

	return rsiCrossover
}

func (m *rsiCrossOver) Stop()  {

}

func (m *rsiCrossOver) Execute()  {
	//m.processData(marketData)

	// TO REMOVE
	fmt.Println(time.Now(),"\n")

	if m.observations[0] == 0 {
		return
	}

	rsi := indicators.RSI(m.observations, m.observationsLength)

	m.rsiObservations = m.rsiObservations[1:]
	m.rsiObservations = append(m.rsiObservations, rsi)


	ema := indicators.ExponentialMA(14, m.lastEMA, rsi)

	m.lastEMA = ema

	fmt.Printf("RSI: %f \t EMA: %f \n", rsi, ema)

	//m.evaluate(rsi,ema, marketData)
}

func (m *rsiCrossOver) evaluate(rsi, ema float64, marketData *block.Block) {

	// Some sort of smoothing for current rsi
	prevRsiIndex := m.observationsLength - 3
	rsi = (m.rsiObservations[prevRsiIndex] + rsi) /2

	if m.onTrade {
		m.glideOnRsi(rsi, marketData)
		return
	}

	if rsi > ema && !m.onTrade {
		m.lastOrderAmount = account.QuantityFromPrice(m.settings.BaseBid, m.observations[m.observationsLength-1])
		order, err := m.account.PlaceMarketOrder(m.lastOrderAmount, m.settings.Symbol, binance.SideTypeBuy)
		if err != nil {
			m.handler(marketData,order, nil, err)
			return
		}
		m.onTrade = true
	}
}

func (m *rsiCrossOver) glideOnRsi(rsi float64, marketData *block.Block)  {
	if rsi < m.lastEMA || rsi < m.rsiObservations[m.observationsLength-2]{
		order, err := m.account.PlaceMarketOrder(m.lastOrderAmount, m.settings.Symbol, binance.SideTypeSell)
		if err != nil {
			m.handler(marketData,order, nil, err)
			return
		}
		m.onTrade = false
	}
}

func (m *rsiCrossOver) processData(marketData *block.Block)  {
	m.observations = m.observations[1:]
	m.observations = append(m.observations, marketData.AveragePrice)
}
