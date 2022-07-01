package mac

import (
	"encoding/json"
	"fmt"
	"log"
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
	observations []*dataToStore

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
func NewMacStrategy(monitorChannel chan *block.Block, config MovingAverageCrossoverConfig, keys *users.Keys, historicalData []*block.Block, inst *instance.StrategyInstance) (strategy.Strategy, error) {

	acc, err := account.NewBinanceAccount(keys.ApiKey,keys.SecretKey, keys.ApiKey, keys.SecretKey)
	if err != nil {
		return nil, err
	}

	observations := make([]*dataToStore, config.LongTermPeriod)

	if historicalData[0] != nil {
		for _, dataFrame := range historicalData {
			sumPerFrame := 0.
			for i := range dataFrame.Trades {
				sumPerFrame += dataFrame.Trades[i]
			}
			observations = observations[1:]
			observations = append(observations, &dataToStore{Sum: sumPerFrame, Count: dataFrame.TradesCount})
		}
	}

	strategy := &movingAverageCrossover{
		config: config,
		observations: observations,
	}
	strategy.Account = acc
	strategy.StopSignal = make(chan bool)
	strategy.MonitorChannel = monitorChannel
	strategy.StrategyInstance = inst
	strategy.HandlerFunction = strategy.HandlerFunc
	strategy.DataProcessFunction = strategy.processData

	return strategy, nil
}

func (m *movingAverageCrossover) HandlerFunc(marketData *block.Block)  {
	if m.observations[0] != nil {

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

func (m *movingAverageCrossover) executeSimpleMA(marketData *block.Block) {
	longTerm := m.getAverage(m.config.LongTermPeriod)
	shortTerm := m.getAverage(m.config.ShortTermPeriod)

	m.evaluate(marketData,longTerm,shortTerm)
}

func (m *movingAverageCrossover) executeExponentialMA(marketData *block.Block) {
	if m.emaPrevious.LongEMA == 0 || m.emaPrevious.ShortEMA == 0 {
		m.emaPrevious.LongEMA = m.getAverage(m.config.LongTermPeriod)
		m.emaPrevious.ShortEMA = m.getAverage(m.config.ShortTermPeriod)
	}

	longEMA := indicators.ExponentialMA(m.config.LongTermPeriod, m.emaPrevious.LongEMA, m.getAverage(m.config.LongTermPeriod))
	shortEMA := indicators.ExponentialMA(m.config.ShortTermPeriod, m.emaPrevious.ShortEMA, m.getAverage(m.config.ShortTermPeriod))

	m.evaluate(marketData, longEMA, shortEMA)

}

func (m *movingAverageCrossover) evaluate(marketData *block.Block, longTerm, shortTerm float64)  {
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

	df := &FrameData{Time: time.Now(), LongTerm: longTerm, ShortTerm: shortTerm}
	data, err := json.Marshal(df)
	if err != nil {
		log.Println(err.Error())
	}

	db, err := database.GetDataBaseConnection()
	if err != nil {
		log.Println(err.Error())
	}
	err = instance.NewDataFrame(db, &instance.DataFrame{InstanceID: m.StrategyInstance.ID, Data: data})
	if err != nil {
		log.Println(err.Error())
	}
}

func (m *movingAverageCrossover) processData(marketData *block.Block)  {
	sumPerFrame := 0.
	for i := range marketData.Trades {
		sumPerFrame += marketData.Trades[i]
	}
	m.observations = m.observations[1:]
	m.observations = append(m.observations, &dataToStore{Sum: sumPerFrame, Count: marketData.TradesCount})
}

func (m *movingAverageCrossover) getAverage(timeFrames int) float64  {
	sum := 0.
	count := 0
	length := len(m.observations)-1
	for i := length; i > length - timeFrames; i-- {
		sum += m.observations[i].Sum
		count += m.observations[i].Count
	}

	return sum/float64(count)
}

