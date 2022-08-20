package qqe

import (
	"math"
	"newTradingBot/indicators"
	"newTradingBot/logs"
	"newTradingBot/models/account"
	"newTradingBot/models/block"
	"newTradingBot/models/strategy"
	"newTradingBot/models/strategy/instance"
	"newTradingBot/models/users"
	"newTradingBot/strategies/common"
)

type qqeStrategy struct {
	common.Strategy

	config                 QQEStrategyConfig
	closePriceObservations []float64

	wildersPeriod int

	prevRSI float64
	prevAtrRsi float64
	prevAtrRSIMa float64

	prevRSIndex float64
	prevLongband float64
	prevShortband float64

	crossLong bool
	crossShort bool

	prevTrend int

	prevQQExlong int
	prevQQExshort int

	prevFastAtrRsiTL float64
}

// NewQQEStrategy - creates new Moving Average crossover strategy
func NewQQEStrategy(monitorChannel chan *block.Data, config QQEStrategyConfig, keys *users.Keys, historicalData []*block.Data, inst *instance.StrategyInstance) (strategy.Strategy, error) {

	acc, err := account.NewBinanceAccount(keys.ApiKey,keys.SecretKey, keys.ApiKey, keys.SecretKey)
	if err != nil {
		return nil, err
	}


	newStrategy := &qqeStrategy{
		config: config,
	}
	newStrategy.Account = acc
	newStrategy.StopSignal = make(chan bool)
	newStrategy.MonitorChannel = monitorChannel
	newStrategy.StrategyInstance = inst
	newStrategy.HandlerFunction = newStrategy.HandlerFunc
	newStrategy.DataProcessFunction = newStrategy.ProcessData

	newStrategy.wildersPeriod = (config.RSIPeriod * 2) - 1

	newStrategy.closePriceObservations = make([]float64, newStrategy.config.RSIPeriod)

	return newStrategy, nil
}

func (t *qqeStrategy) HandlerFunc(marketData *block.Data)  {
	rsi := 0.

	if t.closePriceObservations[0] != 0 {
		rsi = indicators.RSI(t.closePriceObservations, t.config.RSIPeriod)

		// Start here
		if t.prevRSI == 0 {
			t.prevRSI = rsi
			return
		}
		rsiMA := indicators.ExponentialMA(t.config.RSISmoothing, t.prevRSI, rsi)
		atrRsi := math.Abs(t.prevRSI - rsiMA)

		if t.prevAtrRsi == 0 {
			t.prevAtrRsi = atrRsi
			return
		}

		MaAtrRsi := indicators.ExponentialMA(t.wildersPeriod, t.prevAtrRsi, atrRsi)

		if t.prevAtrRSIMa == 0 {
			t.prevAtrRSIMa = MaAtrRsi
			return
		}

		dar := indicators.ExponentialMA(t.wildersPeriod, MaAtrRsi, t.config.FastQQE)

		longband := 0.0
		shortband := 0.0
		trend := 0

		DeltaFastAtrRsi := dar
		RSIndex := rsiMA

		newshortband := RSIndex + DeltaFastAtrRsi
		newlongband := RSIndex - DeltaFastAtrRsi

		if t.prevLongband == 0 {
			t.prevLongband = newlongband
			t.prevShortband = newshortband
			t.prevRSIndex = RSIndex
			return
		}

		// calculate longband
		if t.prevRSIndex > t.prevLongband && RSIndex > t.prevLongband {
			longband = indicators.Max([]float64{t.prevLongband, newlongband})
		} else {
			longband = newlongband
		}

		// calculate shortband

		if t.prevRSIndex < t.prevShortband && RSIndex < t.prevShortband {
			shortband = indicators.Min([]float64{t.prevShortband, newshortband})
		} else {
			shortband = newshortband
		}

		crossOne := t.CrossLong(RSIndex)
		crossTwo := t.CrossShort(RSIndex)

		if crossTwo {
			trend = 1
		} else if crossOne {
			trend = -1
		} else {
			trend = t.prevTrend
		}

		// set back
		t.prevTrend = trend

		FastAtrRsiTL := shortband

		if trend == 1 {
			FastAtrRsiTL = longband
		}

		if FastAtrRsiTL < RSIndex {
			t.prevQQExlong += 1
		} else {
			t.prevQQExlong = 0
		}

		if FastAtrRsiTL > RSIndex {
			t.prevQQExshort += 1
		} else {
			t.prevQQExshort = 0
		}

		if t.prevFastAtrRsiTL == 0 {
			t.prevFastAtrRsiTL = FastAtrRsiTL
			return
		}

		var qqeLong *float64
		var qqeShort *float64

		if t.prevQQExlong == 1 {
			val := t.prevFastAtrRsiTL -50
			qqeLong = &val
		}

		if t.prevQQExshort == 1 {
			val := t.prevFastAtrRsiTL
			qqeShort = &val
		}


		// setback all values
		t.prevRSI = rsi
		t.prevAtrRsi = atrRsi
		t.prevAtrRSIMa = MaAtrRsi
		t.prevLongband = newlongband
		t.prevShortband = newshortband
		t.prevRSIndex = RSIndex
		t.prevFastAtrRsiTL = FastAtrRsiTL

		t.Evaluate(marketData, qqeLong, qqeShort)

	}
}

func (t *qqeStrategy) Evaluate(marketData *block.Data, qqeLong, qqeShort *float64)  {
	if qqeLong != nil {
		err := t.HandleBuy(marketData)
		if err != nil {
			logs.LogDebug("", err)
			return
		}
	} else if qqeShort != nil {
		err := t.HandleSell(marketData)
		if err != nil {
			logs.LogDebug("", err)
			return
		}
	}
}

func (t *qqeStrategy) CrossLong(rsiIndex float64) bool {
	curr := (t.prevLongband / rsiIndex) > 1
	if curr != t.crossLong {
		t.crossLong = curr
		return true
	}
	t.crossLong = curr
	return false
}

func (t *qqeStrategy) CrossShort(rsiIndex float64) bool {
	curr := (t.prevShortband / rsiIndex) > 1
	if curr != t.crossShort {
		t.crossShort = curr
		return true
	}
	t.crossShort = curr
	return false
}

func (t *qqeStrategy) ProcessData(marketData *block.Data)  {
	t.closePriceObservations = t.closePriceObservations[1:]
	t.closePriceObservations = append(t.closePriceObservations, marketData.ClosePrice)
}
