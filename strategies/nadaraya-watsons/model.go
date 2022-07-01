package nadaraya_watsons

import (
	"encoding/json"
	"fmt"
	"github.com/adshao/go-binance/v2/futures"
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

type nadarayaWatsons struct {
	strategyInstance *instance.StrategyInstance

	account account.Account
	config  MovingAverageCrossoverConfig

	monitorChannel chan *block.Block

	stopSignal chan bool
	observations []*dataToStore

	lastTrade *trade.Trade
}

type dataToStore struct {
	closePrice float64
}


type FrameData struct {
	Time time.Time `json:"time"`
	LongTerm float64 `json:"longTerm"`
	ShortTerm float64 `json:"shortTerm"`
}

func (n *nadarayaWatsons) GetInstance() *instance.StrategyInstance {
	return n.strategyInstance
}


// NewNadarayeWatsonsStrategy - creates new Moving Average crossover strategy
func NewNadarayeWatsonsStrategy(monitorChannel chan *block.Block, config MovingAverageCrossoverConfig, keys *users.Keys, historicalData []*block.Block, inst *instance.StrategyInstance) (strategy.Strategy, error) {

	acc, err := account.NewBinanceAccount(keys.ApiKey,keys.SecretKey, keys.ApiKey, keys.SecretKey)
	if err != nil {
		return nil, err
	}

	observations := make([]*dataToStore, 50)

	if historicalData[0] != nil {
		for _, dataFrame := range historicalData {
			sumPerFrame := 0.
			for i := range dataFrame.Trades {
				sumPerFrame += dataFrame.Trades[i]
			}
			//observations = observations[1:]
			//observations = append(observations, &dataToStore{Sum: sumPerFrame, Count: dataFrame.TradesCount})
		}
	}

	return &nadarayaWatsons{
		config: config,
		account: acc,
		stopSignal: make(chan bool),
		monitorChannel: monitorChannel,
		observations: observations,
		strategyInstance: inst,
	}, nil
}

func (n *nadarayaWatsons) Execute()  {
	go func() {
		for  {
			select {
			case <-n.stopSignal:
				if n.lastTrade != nil {
					// TODO: error handling
					err := n.sell(nil)
					if err != nil {
						log.Println(err.Error())
					}
				}
				return
			default:
				marketData := <- n.monitorChannel
				n.processData(marketData)

				logs.LogDebug(fmt.Sprintf("Data received by instance #%d", n.strategyInstance.ID), nil)

				if n.observations[0] != nil {

					if n.strategyInstance.Status == instance.StatusCreated {
						n.strategyInstance.Status = instance.StatusRunning
						db, _ := database.GetDataBaseConnection()
						_ = instance.UpdateStatus(db, n.strategyInstance.ID, instance.StatusRunning)
					}

					n.executeNW(marketData)
				}
			}
		}
	}()
}

func (n *nadarayaWatsons) executeNW(marketData *block.Block) {

}

func (n *nadarayaWatsons) evaluate(marketData *block.Block) {

	logs.LogDebug(fmt.Sprintf("Data processing is finished by instance #%d", n.strategyInstance.ID), nil)

}

func (n *nadarayaWatsons) Stop()  {
	go func() {
		n.stopSignal <- true
	}()

	db, err := database.GetDataBaseConnection()
	if err != nil {
		logs.LogDebug("", err)
	}

	err = instance.UpdateStatus(db, n.strategyInstance.ID, instance.StatusStopped)
	if err != nil {
		logs.LogDebug("", err)
	}
}

func (n *nadarayaWatsons) handleSell(marketData *block.Block) error {
	if n.strategyInstance.IsFutures {
		if n.lastTrade == nil {
			return n.sell(marketData)
		}
		if n.lastTrade.FuturesSide != futures.SideTypeSell{
			return n.sell(marketData)
		}
		return nil
	} else {
		return n.sell(marketData)
	}
}

func (n *nadarayaWatsons) handleBuy(marketData *block.Block) error {
	if n.strategyInstance.IsFutures {
		if n.lastTrade == nil {
			return n.buy(marketData)
		}
		if n.lastTrade.FuturesSide != futures.SideTypeBuy {
			return n.buy(marketData)
		}
		return nil
	} else {
		return n.buy(marketData)
	}
}

