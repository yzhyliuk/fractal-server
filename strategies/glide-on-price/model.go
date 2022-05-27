package glide_on_price

import (
	"fmt"
	"github.com/adshao/go-binance/v2"
	"newTradingBot/indicators"
	"newTradingBot/models/account"
	"newTradingBot/models/block"
	"newTradingBot/models/strategy"
	"time"
)

type glideOnPrice struct {

	observations []*block.Block
	observationsLength int

	priceChangeObservations []float64

	timeFrame time.Duration

	handler strategy.HandlerFunc
	account account.Account
	settings *strategy.Settings

	lastOrder *account.InternalFuturesOrder
	lastOrderAmount float64
	lastPriceChangeAverage float64

	onTrade bool
}

func NewGlideOnPrice(handler strategy.HandlerFunc, account account.Account,
	settings *strategy.Settings, timeFrame time.Duration, monitorTimeFrame time.Duration) strategy.Strategy {

	glideOnPriceStrategy := &glideOnPrice{
		handler: handler,
		account: account,
		settings: settings,
		observationsLength: int((timeFrame*3)/monitorTimeFrame),
		onTrade: false,
		lastPriceChangeAverage: 1,
	}

	glideOnPriceStrategy.timeFrame = timeFrame

	glideOnPriceStrategy.priceChangeObservations = make([]float64, 7)
	glideOnPriceStrategy.observations = make([]*block.Block, glideOnPriceStrategy.observationsLength)

	return glideOnPriceStrategy
}

func (m *glideOnPrice) Execute(marketData *block.Block)  {

	fmt.Println(time.Now().Format("15:04:05 2006-01-02 "))

	m.processData(marketData)

	if m.priceChangeObservations[0] == 0 {
		if m.observations[0] == nil {
			return
		}
		m.calculatePriceChange()
		return
	}

	m.calculatePriceChange()

	delta := m.observations[2].ClosePrice/m.observations[0].ClosePrice

	priceChangeAverage := indicators.Average(m.priceChangeObservations)
	priceChangeRatio := priceChangeAverage / m.lastPriceChangeAverage
	m.lastPriceChangeAverage = priceChangeAverage



	fmt.Printf("%f \t %f \t %f \n", m.observations[0].ClosePrice, m.observations[1].ClosePrice, m.observations[2].ClosePrice)
	fmt.Printf("Delta: %f \n", delta)
	fmt.Printf("Price Change Ratio: %f \n \n", priceChangeRatio)

	if delta > 1 && !m.onTrade {
		m.lastOrderAmount = account.QuantityFromPrice(m.settings.BaseBid, m.observations[m.observationsLength-1].ClosePrice)
		order, err := m.account.PlaceMarketOrder(m.lastOrderAmount, m.settings.Symbol, binance.SideTypeBuy)
		if err != nil {
			m.handler(marketData,order, nil, err)
			return
		}
		m.onTrade = true
	} else if delta < 1 && m.onTrade{
		order, err := m.account.PlaceMarketOrder(m.lastOrderAmount, m.settings.Symbol, binance.SideTypeSell)
		if err != nil {
			m.handler(marketData,order, nil, err)
			return
		}
		m.onTrade = false
	}


}

func (m *glideOnPrice) getLowestPrice() float64  {
	lowest := 9999999999999.
	for _, data := range m.observations {
		if lowest > data.MinPrice {
			lowest = data.MinPrice
		}
	}

	return lowest
}

func (m *glideOnPrice) processData(marketData *block.Block)  {
	m.observations = m.observations[1:]
	m.observations = append(m.observations, marketData)
}

func (m *glideOnPrice) calculatePriceChange() {
	delta := m.observations[2].ClosePrice - m.observations[1].ClosePrice

	if delta < 0 {
		delta = -delta
	}
	m.priceChangeObservations = m.priceChangeObservations[1:]
	m.priceChangeObservations = append(m.priceChangeObservations, delta)
}
