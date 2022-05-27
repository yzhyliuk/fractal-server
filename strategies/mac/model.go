package mac

import (
	"encoding/json"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"log"
	"newTradingBot/api/database"
	"newTradingBot/indicators"
	"newTradingBot/logs"
	"newTradingBot/models/account"
	"newTradingBot/models/block"
	"newTradingBot/models/strategy"
	"newTradingBot/models/strategy/instance"
	"newTradingBot/models/trade"
	"newTradingBot/models/users"
	"time"
)

type mac struct {
	strategyInstance *instance.StrategyInstance

	account account.Account
	config MovingAverageCrossoverConfig

	monitorChannel chan *block.Block

	stopSignal chan bool
	observations []*dataToStore

	lastTrade *trade.Trade

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

func (m *mac) GetInstance() *instance.StrategyInstance {
	return m.strategyInstance
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

	return &mac{
		config: config,
		account: acc,
		stopSignal: make(chan bool),
		monitorChannel: monitorChannel,
		observations: observations,
		strategyInstance: inst,
	}, nil
}

func (m *mac) Execute()  {
	go func() {
		for  {
			select {
			case <-m.stopSignal:
				return
			default:
				marketData := <- m.monitorChannel
				m.processData(marketData)

				logs.LogDebug(fmt.Sprintf("Data received by instance #%d", m.strategyInstance.ID), nil)

				if m.observations[0] != nil {

					if m.strategyInstance.Status == instance.StatusCreated {
						m.strategyInstance.Status = instance.StatusRunning
						db, _ := database.GetDataBaseConnection()
						_ = instance.UpdateStatus(db, m.strategyInstance.ID, instance.StatusRunning)
					}

					switch m.config.MovingAverageType {
					case "sma":
						m.executeSimpleMA(marketData)
					case "ema":
						m.executeExponentialMA(marketData)
					}
				}
			}
		}
	}()
}

func (m *mac) executeSimpleMA(marketData *block.Block) {
	longTerm := m.getAverage(m.config.LongTermPeriod)
	shortTerm := m.getAverage(m.config.ShortTermPeriod)
	if shortTerm > longTerm && m.lastTrade == nil {
		logs.LogDebug("Buy order", nil)
		err := m.buy(marketData)
		if err != nil {
			logs.LogDebug("", err)
		}

	} else if shortTerm < longTerm && m.lastTrade != nil {
		// Sell
		logs.LogDebug("Sell order", nil)
		err := m.sell()
		if err != nil {
			logs.LogDebug("", err)
		}
	}

	logs.LogDebug(fmt.Sprintf("Data processing is finished by instance #%d", m.strategyInstance.ID), nil)

	df := &FrameData{Time: time.Now(), LongTerm: longTerm, ShortTerm: shortTerm}
	data, err := json.Marshal(df)
	if err != nil {
		log.Println(err.Error())
	}

	db, err := database.GetDataBaseConnection()
	if err != nil {
		log.Println(err.Error())
	}
	err = instance.NewDataFrame(db, &instance.DataFrame{InstanceID: m.strategyInstance.ID, Data: data})
	if err != nil {
		log.Println(err.Error())
	}
}

func (m *mac) executeExponentialMA(marketData *block.Block) {
	if m.emaPrevious.LongEMA == 0 || m.emaPrevious.ShortEMA == 0 {
		m.emaPrevious.LongEMA = m.getAverage(m.config.LongTermPeriod)
		m.emaPrevious.ShortEMA = m.getAverage(m.config.ShortTermPeriod)
	}

	longEMA := indicators.ExponentialMA(m.config.LongTermPeriod, m.emaPrevious.LongEMA, m.getAverage(m.config.LongTermPeriod))
	shortEMA := indicators.ExponentialMA(m.config.ShortTermPeriod, m.emaPrevious.ShortEMA, m.getAverage(m.config.ShortTermPeriod))

	if shortEMA > longEMA && m.lastTrade == nil{
		// Buy
		// TODO: error handling
		m.buy(marketData)
	} else if shortEMA < longEMA && m.lastTrade != nil{
		// Sell
		// TODO: error handling
		m.sell()
	}

}

func (m *mac) processData(marketData *block.Block)  {
	sumPerFrame := 0.
	for i := range marketData.Trades {
		sumPerFrame += marketData.Trades[i]
	}
	m.observations = m.observations[1:]
	m.observations = append(m.observations, &dataToStore{Sum: sumPerFrame, Count: marketData.TradesCount})
}

func (m *mac) getAverage(timeFrames int) float64  {
	sum := 0.
	count := 0
	length := len(m.observations)-1
	for i := length; i > length - timeFrames; i-- {
		sum += m.observations[i].Sum
		count += m.observations[i].Count
	}

	return sum/float64(count)
}

func (m *mac) Stop()  {
	m.stopSignal <- true

	if m.lastTrade != nil {
		// TODO: error handling
		err := m.sell()
		if err != nil {
			log.Println(err.Error())
		}
	}
}

func (m *mac) buy(marketData *block.Block) error {
	quantity := account.QuantityFromPrice(m.config.BidSize, marketData.ClosePrice )
	spotTrade, err := m.account.PlaceMarketOrder(quantity,m.config.Pair,binance.SideTypeBuy, m.GetInstance(), m.lastTrade)
	if err != nil {
		return err
	}

	m.lastTrade = spotTrade

	return nil
}

func (m *mac) sell() error {
	_, err := m.account.PlaceMarketOrder(m.lastTrade.Quantity,m.config.Pair,binance.SideTypeSell, m.GetInstance(), m.lastTrade)
	if err != nil {
		return err
	}

	m.lastTrade = nil
	return nil
}

