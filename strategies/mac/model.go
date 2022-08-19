package mac

import (
	"fmt"
	"newTradingBot/api/database"
	"newTradingBot/indicators"
	"newTradingBot/logs"
	"newTradingBot/models/account"
	"newTradingBot/models/block"
	"newTradingBot/models/strategy"
	"newTradingBot/models/strategy/instance"
	"newTradingBot/models/users"
	"newTradingBot/strategies/common"
	"time"
)

type movingAverageCrossover struct {
	common.Strategy

	config MovingAverageCrossoverConfig
	observations []float64

	emaPrevious emaData
}

type dataToStore struct {
	Sum float64
	Count int
}

type emaData struct {
	LongEMA float64
	ShortEMA float64
}

type FrameData struct {
	Time time.Time `json:"time"`
	LongTerm float64 `json:"longTerm"`
	ShortTerm float64 `json:"shortTerm"`
}


// NewMacStrategy - creates new Moving Average crossover strategy
func NewMacStrategy(monitorChannel chan *block.Data, config MovingAverageCrossoverConfig, keys *users.Keys, historicalData []*block.Data, inst *instance.StrategyInstance) (strategy.Strategy, error) {

	acc, err := account.NewBinanceAccount(keys.ApiKey,keys.SecretKey, keys.ApiKey, keys.SecretKey)
	if err != nil {
		return nil, err
	}

	//observations := make([]*dataToStore, config.LongTermPeriod)
	observations := make([]float64, config.LongTermPeriod)

	strat := &movingAverageCrossover{
		config: config,
		observations: observations,
	}
	strat.Account = acc
	strat.StopSignal = make(chan bool)
	strat.MonitorChannel = monitorChannel
	strat.StrategyInstance = inst
	strat.HandlerFunction = strat.HandlerFunc
	strat.DataProcessFunction = strat.processData

	return strat, nil
}

func (m *movingAverageCrossover) HandlerFunc(marketData *block.Data)  {
	if m.observations[0] != 0 {

		if m.StrategyInstance.Status == instance.StatusCreated {
			m.StrategyInstance.Status = instance.StatusRunning
			db, _ := database.GetDataBaseConnection()
			_ = instance.UpdateStatus(db, m.StrategyInstance.ID, instance.StatusRunning)
		}

		switch m.config.MovingAverageType {
		case "sma":
			m.executeSimpleMA(marketData)
		case "ema":
			m.executeExponentialMA(marketData)
		}
	}
}

func (m *movingAverageCrossover) executeSimpleMA(marketData *block.Data) {
	longTerm := m.getAverage(m.config.LongTermPeriod)
	shortTerm := m.getAverage(m.config.ShortTermPeriod)

	m.evaluate(marketData,longTerm,shortTerm)
}

func (m *movingAverageCrossover) executeExponentialMA(marketData *block.Data) {
	if m.emaPrevious.LongEMA == 0 || m.emaPrevious.ShortEMA == 0 {
		m.emaPrevious.LongEMA = m.getAverage(m.config.LongTermPeriod)
		m.emaPrevious.ShortEMA = m.getAverage(m.config.ShortTermPeriod)
	}

	longEMA := indicators.ExponentialMA(m.config.LongTermPeriod, m.emaPrevious.LongEMA, m.getAverage(m.config.LongTermPeriod))
	shortEMA := indicators.ExponentialMA(m.config.ShortTermPeriod, m.emaPrevious.ShortEMA, m.getAverage(m.config.ShortTermPeriod))

	m.emaPrevious.LongEMA = longEMA
	m.emaPrevious.ShortEMA = shortEMA

	m.evaluate(marketData, longEMA, shortEMA)

}

func (m *movingAverageCrossover) evaluate(marketData *block.Data, longTerm, shortTerm float64)  {
	if shortTerm > longTerm {
		logs.LogDebug("Buy order", nil)
		err := m.HandleBuy(marketData)
		if err != nil {
			logs.LogDebug("",err)
		}
	} else if shortTerm < longTerm {
		logs.LogDebug("Sell order", nil)
		err := m.HandleSell(marketData)
		if err != nil {
			logs.LogDebug("",err)
		}
	}

	logs.LogDebug(fmt.Sprintf("Data processing is finished by instance #%d", m.StrategyInstance.ID), nil)
	logs.LogDebug(fmt.Sprintf("LONG: %f \t SHORT: %f",longTerm, shortTerm), nil)
}

func (m *movingAverageCrossover) processData(marketData *block.Data)  {
	//sumPerFrame := 0.
	//for i := range marketData.TradesArray {
	//	sumPerFrame += marketData.TradesArray[i]
	//}
	m.observations = m.observations[1:]
	m.observations = append(m.observations, marketData.ClosePrice)
}

func (m *movingAverageCrossover) getAverage(timeFrames int) float64  {
	frame := make([]float64, 0)
	length := len(m.observations)-1
	for i := length; i > length - timeFrames; i-- {
		frame = append(frame, m.observations[i])
	}

	return indicators.Average(frame)
}

