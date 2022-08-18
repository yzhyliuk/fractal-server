package glide_on_price

import (
	"encoding/json"
	"fmt"
	"log"
	"newTradingBot/api/database"
	"newTradingBot/logs"
	"newTradingBot/models/account"
	"newTradingBot/models/block"
	"newTradingBot/models/strategy"
	"newTradingBot/models/strategy/instance"
	"newTradingBot/models/users"
	"newTradingBot/strategies/common"
	"time"
)

type glideOnPrice struct {
	common.Strategy
	config GlideOnPriceConfig

	observations []float64
	prevMarketData *block.Data
}

type FrameData struct {
	Time time.Time `json:"time"`
	Volatility float64 `json:"volatility"`
	Slope float64 `json:"slope"`
}

// NewGlideOnPriceStrategy - creates new Moving Average crossover strategy
func NewGlideOnPriceStrategy(monitorChannel chan *block.Data, config GlideOnPriceConfig, keys *users.Keys, historicalData []*block.Data, inst *instance.StrategyInstance) (strategy.Strategy, error) {

	acc, err := account.NewBinanceAccount(keys.ApiKey,keys.SecretKey, keys.ApiKey, keys.SecretKey)
	if err != nil {
		return nil, err
	}


	strategy := &glideOnPrice{
		config: config,
	}
	strategy.Account = acc
	strategy.StopSignal = make(chan bool)
	strategy.MonitorChannel = monitorChannel
	strategy.StrategyInstance = inst
	strategy.HandlerFunction = strategy.HandlerFunc
	strategy.DataProcessFunction = strategy.ProcessData
	
	strategy.observations = make([]float64, config.VolatilityObservationsTimeFrame)

	return strategy, nil
}

func (g *glideOnPrice) HandlerFunc(marketData *block.Data)  {
	if g.observations[0] != 0 {

		if g.StrategyInstance.Status == instance.StatusCreated {
			g.StrategyInstance.Status = instance.StatusRunning
			db, _ := database.GetDataBaseConnection()
			_ = instance.UpdateStatus(db, g.StrategyInstance.ID, instance.StatusRunning)
		}

		lastIndex := len(g.observations)-1
		prevIndex := len(g.observations)-g.config.SlopeTimeFrameDifference

		volatility := g.getHighest()/g.getLowestPrice()
		slope := g.observations[lastIndex]/g.observations[prevIndex]

		if volatility > g.config.VolatilityLimit {
			if slope > 1 {
				logs.LogDebug("Buy order", nil)
				err := g.HandleBuy(marketData)
				if err != nil {
					logs.LogDebug("",err)
				}
			} else {
				logs.LogDebug("Sell order", nil)
				err := g.HandleSell(marketData)
				if err != nil {
					logs.LogDebug("",err)
				}
			}
		} else {
			g.CloseAllTrades()
		}

		logs.LogDebug(fmt.Sprintf("Data processing is finished by instance #%d", g.StrategyInstance.ID), nil)

		df := &FrameData{Time: time.Now(), Volatility: volatility, Slope: slope}
		data, err := json.Marshal(df)
		if err != nil {
			log.Println(err.Error())
		}

		db, err := database.GetDataBaseConnection()
		if err != nil {
			log.Println(err.Error())
		}
		err = instance.NewDataFrame(db, &instance.DataFrame{InstanceID: g.StrategyInstance.ID, Data: data})
		if err != nil {
			log.Println(err.Error())
		}
	}
}

func (g *glideOnPrice) ProcessData(marketData *block.Data)  {
	g.observations = g.observations[1:]
	g.observations = append(g.observations, marketData.ClosePrice)
}

func (g *glideOnPrice) getLowestPrice() float64 {
	lowest := 99999999999999999999999.
	for i := range g.observations {
		if lowest > g.observations[i] {
			lowest = g.observations[i]
		}
	}

	return lowest
}

func (g *glideOnPrice) getHighest() float64 {
	highest := 0.
	for i := range g.observations {
		if highest < g.observations[i] {
			highest = g.observations[i]
		}
	}

	return highest
}