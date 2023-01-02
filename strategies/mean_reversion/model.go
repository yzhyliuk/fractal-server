package mean_reversion

import (
	"encoding/json"
	"github.com/adshao/go-binance/v2/futures"
	"newTradingBot/api/database"
	"newTradingBot/indicators"
	"newTradingBot/logs"
	"newTradingBot/models/account"
	"newTradingBot/models/block"
	"newTradingBot/models/strategy"
	"newTradingBot/models/strategy/instance"
	"newTradingBot/models/testing"
	"newTradingBot/models/users"
	"newTradingBot/strategies/common"
)

const StrategyName = "mean_reversion"

type meanReversion struct {
	common.Strategy

	config                 MeanReversionConfig
	closePriceObservations []float64

	prevMean float64
}

// NewMeanReversion - creates new Moving Average crossover strategy
func NewMeanReversion(monitorChannel chan *block.Data, configRaw []byte, keys *users.Keys, historicalData []*block.Data, inst *instance.StrategyInstance) (strategy.Strategy, error) {

	var config MeanReversionConfig

	err := json.Unmarshal(configRaw, &config)
	if err != nil {
		return nil,err
	}

	acc, err := account.NewBinanceAccount(keys.ApiKey,keys.SecretKey, keys.ApiKey, keys.SecretKey)
	if err != nil {
		return nil, err
	}


	newStrategy := &meanReversion{
		config: config,
	}
	newStrategy.Account = acc
	newStrategy.StopSignal = make(chan bool)
	newStrategy.MonitorChannel = monitorChannel
	newStrategy.StrategyInstance = inst
	newStrategy.HandlerFunction = newStrategy.HandlerFunc
	newStrategy.DataProcessFunction = newStrategy.ProcessData


	newStrategy.closePriceObservations = make([]float64, newStrategy.config.MeanPeriod)

	return newStrategy, nil
}

func (t *meanReversion) HandlerFunc(marketData *block.Data)  {
	if t.closePriceObservations[0] != 0 {
		if t.StrategyInstance.Status == instance.StatusCreated && t.StrategyInstance.Testing == testing.Disable {
			t.StrategyInstance.Status = instance.StatusRunning
			db, _ := database.GetDataBaseConnection()
			_ = instance.UpdateStatus(db, t.StrategyInstance.ID, instance.StatusRunning)
		}

		if t.prevMean == 0 {
			t.prevMean = marketData.ClosePrice
		}

		mean := indicators.SimpleMA(t.closePriceObservations, len(t.closePriceObservations))
		// handle close trade
		if t.EvaluateExit(marketData, mean) {
			return
		}

		sd := indicators.StandardDeviation(t.closePriceObservations)

		outOfMeanUp := mean + (float64(t.config.SDMultiplier)*sd)
		outOfMeanDown := mean - (float64(t.config.SDMultiplier)*sd)

		t.Evaluate(marketData, outOfMeanDown, outOfMeanUp)
	}
}

func (t *meanReversion) Evaluate(marketData *block.Data, down, up float64)  {

	if down > marketData.ClosePrice {
		err := t.HandleBuy(marketData)
		if err != nil {
			logs.LogDebug("", err)
			return
		}
	}

	if up < marketData.ClosePrice {
		err := t.HandleSell(marketData)
		if err != nil {
			logs.LogDebug("", err)
			return
		}
	}
}

func (t *meanReversion) EvaluateExit(marketData *block.Data, mean float64) bool {
	if t.LastTrade != nil {
		if t.LastTrade.IsFutures && t.LastTrade.FuturesSide == futures.SideTypeBuy {
			if marketData.ClosePrice > mean {
				t.CloseAllTrades()
				return true
			}
		} else if t.LastTrade.IsFutures && t.LastTrade.FuturesSide == futures.SideTypeSell {
			if marketData.ClosePrice < mean {
				t.CloseAllTrades()
				return true
			}
		}
	}

	return false
}

func (t *meanReversion) ProcessData(marketData *block.Data)  {
	t.closePriceObservations = t.closePriceObservations[1:]
	t.closePriceObservations = append(t.closePriceObservations, marketData.ClosePrice)
}
