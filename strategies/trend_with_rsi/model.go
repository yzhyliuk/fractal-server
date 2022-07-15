package trend_with_rsi

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

type trendWithRSIStrategy struct {
	common.Strategy

	config TrendWithRSIConfig
	closePriceObservations []float64
	movingAverageObservations []*dataToStore

	orderUptrend bool
}

type dataToStore struct {
	Sum float64
	Count int
}

type FrameData struct {
	Time time.Time `json:"time"`
}

// NewTemplateStrategy - creates new Moving Average crossover strategy
func NewTemplateStrategy(monitorChannel chan *block.Block, config TrendWithRSIConfig, keys *users.Keys, historicalData []*block.Block, inst *instance.StrategyInstance) (strategy.Strategy, error) {

	acc, err := account.NewBinanceAccount(keys.ApiKey,keys.SecretKey, keys.ApiKey, keys.SecretKey)
	if err != nil {
		return nil, err
	}


	newStrategy := &trendWithRSIStrategy{
		config: config,
	}
	newStrategy.Account = acc
	newStrategy.StopSignal = make(chan bool)
	newStrategy.MonitorChannel = monitorChannel
	newStrategy.StrategyInstance = inst
	newStrategy.HandlerFunction = newStrategy.HandlerFunc
	newStrategy.DataProcessFunction = newStrategy.ProcessData

	newStrategy.closePriceObservations = make([]float64, newStrategy.config.RSIPeriod)
	newStrategy.movingAverageObservations = make([]*dataToStore, newStrategy.config.TrendMAPeriod)

	return newStrategy, nil
}

func (t *trendWithRSIStrategy) HandlerFunc(marketData *block.Block)  {
	if t.movingAverageObservations[0] != nil && t.closePriceObservations[0] != 0{
		if t.StrategyInstance.Status == instance.StatusCreated {
			t.StrategyInstance.Status = instance.StatusRunning
			db, _ := database.GetDataBaseConnection()
			_ = instance.UpdateStatus(db, t.StrategyInstance.ID, instance.StatusRunning)
		}

		rsi := indicators.RSI(t.closePriceObservations, t.config.RSIPeriod)
		mean := getMean(t.movingAverageObservations)
		uptrend := marketData.ClosePrice > mean

		rsiOverbought := rsi > float64(t.config.RSIOverboughtLevel)
		rsiOversold := rsi < float64(t.config.RSIOversoldLevel)

		logs.LogDebug(fmt.Sprintf("RSI: %f, Uptrend: %t, OrderUptrend: %t", rsi, uptrend, t.orderUptrend),nil)

		t.Evaluate(marketData,rsi,uptrend,rsiOverbought,rsiOversold)
	}
}

func (t *trendWithRSIStrategy) Evaluate(marketData *block.Block, rsi float64, uptrend, rsiOverbought, rsiOversold bool)  {
	if t.LastTrade != nil {
		if t.orderUptrend && rsi > 50{
			t.CloseAllTrades()
		} else if !t.orderUptrend && rsi < 50 {
			t.CloseAllTrades()
		}
	}
	if uptrend && rsiOversold{
		t.orderUptrend = uptrend
		err := t.HandleBuy(marketData)
		if err != nil {
			logs.LogDebug("", err)
			return
		}
	} else if !uptrend && rsiOverbought{
		t.orderUptrend = uptrend
		err := t.HandleSell(marketData)
		if err != nil {
			logs.LogDebug("", err)
			return
		}

	}
}

func (t *trendWithRSIStrategy) ProcessData(marketData *block.Block)  {
	sumPerFrame := 0.
	for i := range marketData.Trades {
		sumPerFrame += marketData.Trades[i]
	}
	t.movingAverageObservations = t.movingAverageObservations[1:]
	t.movingAverageObservations = append(t.movingAverageObservations, &dataToStore{Sum: sumPerFrame, Count: marketData.TradesCount})

	t.closePriceObservations = t.closePriceObservations[1:]
	t.closePriceObservations = append(t.closePriceObservations, marketData.ClosePrice)
}

func getMean(data []*dataToStore) float64 {
	sum := 0.
	count := 0
	for i := range data {
		sum += data[i].Sum
		count += data[i].Count
	}

	return sum/float64(count)
}
