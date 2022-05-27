package fallx3

import (
	"fmt"
	"github.com/adshao/go-binance/v2/futures"
	"math"
	"newTradingBot/models/account"
	"newTradingBot/models/block"
	"newTradingBot/models/strategy"
	"time"
)

type fallx3 struct {

	observations []float64
	observationsLength int

	timeFrame time.Duration
	sdMultiplier float64

	handler strategy.HandlerFunc
	account account.Account
	settings *strategy.Settings

	isFutures bool

	lastOrder *account.InternalFuturesOrder

}

func NewFallx3Strategy(handler strategy.HandlerFunc, account account.Account,
	settings *strategy.Settings, timeFrame time.Duration,
	sdMultiplayer float64, monitorTimeFrame time.Duration) strategy.Strategy {

	fallx := &fallx3{
		handler: handler,
		account: account,
		settings: settings,
		observationsLength: int(timeFrame/monitorTimeFrame),
	}

	fallx.timeFrame = timeFrame
	fallx.sdMultiplier = sdMultiplayer

	fallx.observations = make([]float64, int(timeFrame/monitorTimeFrame))

	return fallx
}

func (m *fallx3) Execute(marketData *block.Block)  {
	m.processData(marketData)
	currentMean := m.mean()
	currentSD := m.standardDeviation(currentMean)

	fmt.Println(time.Now(),"\n")
	fmt.Println("SD: ", currentSD, "\tMEAN: ", currentMean, "\nCLOSE PRICE: ", marketData.ClosePrice,
		"\nINTERVAL:\n", (currentSD*m.sdMultiplier)+currentMean, " <=> ", currentMean - (currentSD*m.sdMultiplier),
		"\n======================================\n")

	if m.observations[0] == 0 {
		return
	}

	m.evaluate(currentSD, currentMean, marketData.ClosePrice,marketData)
}

func (m *fallx3) evaluate(sd, mean, closePrice float64, marketData *block.Block) {

	if m.lastOrder != nil {
		switch m.lastOrder.Side {
		case futures.SideTypeBuy:
			if marketData.MinPrice < m.lastOrder.StopLoss || marketData.MaxPrice > m.lastOrder.TakeProfit {
				m.lastOrder = nil
			}
		case futures.SideTypeSell:
			if marketData.MaxPrice > m.lastOrder.StopLoss || marketData.MinPrice < m.lastOrder.TakeProfit {
				m.lastOrder = nil
			}
		}
		return
	}

	goDown := closePrice > mean - (sd*m.sdMultiplier)
	deltaDown := (closePrice/mean) - 1
	goUp := closePrice < mean - (sd*m.sdMultiplier)
	deltaUp := (mean/closePrice) - 1

	amount := account.QuantityFromPrice(m.settings.BaseBid, closePrice)
	stopLoss := 0.
	takeProfits := make([]account.TakeProfit,0)
	var side futures.SideType

	if goUp && deltaUp >= minimalProfit {
		stopLoss = mean - (sd*m.sdMultiplier*1.33)
		takeProfits = append(takeProfits, account.TakeProfit{Price: mean + (1.5*sd*m.sdMultiplier),Sum: account.QuantityFromPrice(m.settings.BaseBid, closePrice)})
		side = futures.SideTypeBuy

	} else if goDown && deltaDown >= minimalProfit{
		stopLoss = mean + (sd*m.sdMultiplier*1.33)
		takeProfits = append(takeProfits, account.TakeProfit{Price: mean - (1.5*sd*m.sdMultiplier),Sum: account.QuantityFromPrice(m.settings.BaseBid, closePrice)})
		side = futures.SideTypeSell
	} else {
		return
	}

	order, err := m.account.OpenFuturesPosition(stopLoss, amount,takeProfits,m.settings.Symbol, side)
	if err != nil {
		m.handler(marketData,nil, order, err)
		m.lastOrder = &account.InternalFuturesOrder{
			Side: side,
			TakeProfit: takeProfits[0].Price,
			StopLoss: stopLoss,
		}
	}
}

func (m *fallx3) processData(marketData *block.Block)  {
	m.observations = m.observations[1:]
	m.observations = append(m.observations, marketData.AveragePrice)
}

func (m *fallx3) mean() float64 {
	sum := 0.
	for i := range m.observations {
		sum += m.observations[i]
	}

	return sum/float64(len(m.observations))
}

func (m *fallx3) standardDeviation(mean float64) float64  {
	sum := 0.
	for i := range m.observations {
		sum += math.Pow(m.observations[i] - mean,2)
	}

	return math.Sqrt(sum/float64(len(m.observations)))
}

func (m *fallx3) glide()  {

}

