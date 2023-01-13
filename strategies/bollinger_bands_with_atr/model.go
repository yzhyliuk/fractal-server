package bollinger_bands_with_atr

import (
	"encoding/json"
	"newTradingBot/indicators"
	"newTradingBot/logs"
	"newTradingBot/models/account"
	"newTradingBot/models/block"
	"newTradingBot/models/strategy"
	"newTradingBot/models/strategy/instance"
	"newTradingBot/models/users"
	"newTradingBot/strategies/common"
)

type BollingerBandsWithATR struct {
	common.Strategy
	config BollingerBandsWithATRConfig

	closePrice []float64
	openPrice	[]float64
	lowPrice []float64
	highPrice []float64

	orderTargetPrice float64
	orderStopLossPrice float64

	observationsLength int
}

const StrategyName = "bollinger_bands_with_atr"

// NewBollingerBandsWithATR - creates new Moving Average crossover strategy
func NewBollingerBandsWithATR(monitorChannel chan *block.Data, configRaw []byte, keys *users.Keys, historicalData []*block.Data, inst *instance.StrategyInstance) (strategy.Strategy, error) {

	var config BollingerBandsWithATRConfig

	err := json.Unmarshal(configRaw, &config)
	if err != nil {
		return nil,err
	}

	observationsLength := indicators.MaxOf3int(config.MALength, config.BBLength, config.ATRLength)

	acc, err := account.NewBinanceAccount(keys.ApiKey,keys.SecretKey, keys.ApiKey, keys.SecretKey)
	if err != nil {
		return nil, err
	}


	newStrategy := &BollingerBandsWithATR{
		config: config,
		observationsLength: observationsLength,
	}
	newStrategy.Account = acc
	newStrategy.StopSignal = make(chan bool)
	newStrategy.MonitorChannel = monitorChannel
	newStrategy.StrategyInstance = inst
	newStrategy.HandlerFunction = newStrategy.HandlerFunc
	newStrategy.DataProcessFunction = newStrategy.ProcessData
	newStrategy.closePrice = make([]float64, observationsLength)
	newStrategy.openPrice = make([]float64, observationsLength)
	newStrategy.highPrice = make([]float64, observationsLength)
	newStrategy.lowPrice = make([]float64, observationsLength)

	return newStrategy, nil
}

func (m *BollingerBandsWithATR) HandlerFunc(marketData *block.Data)  {

	//slope := m.closePrice[m.observationsLength-1]/m.closePrice[m.observationsLength-2]

	//longTermMA := indicators.SimpleMA(m.closePrice,m.config.MALength)
	//shortTermMA := indicators.SimpleMA(m.closePrice,m.config.MALength/2)

	upperBB, lowerBB := indicators.BollingerBands(m.closePrice, m.config.BBLength, m.config.BBMultiplier)

	atr := indicators.AverageTrueRange(m.highPrice, m.lowPrice, m.closePrice, m.config.ATRLength)

	//if atr[m.observationsLength-1] < atr[m.observationsLength-2] {
	//	m.CloseAllTrades()
	//}

	if (marketData.High > upperBB) && (atr[m.observationsLength-1] > atr[m.observationsLength-2]) {
		// Sell
		err := m.HandleSell(marketData)
		if err != nil {
			logs.LogError(err)
		}
		return
	}

	if (marketData.Low < lowerBB) && (atr[m.observationsLength-1] > atr[m.observationsLength-2]) {
		//Sell
		err := m.HandleBuy(marketData)
		if err != nil {
			logs.LogError(err)
		}
		return
	}
}

func (m *BollingerBandsWithATR) ProcessData(marketData *block.Data)  {
	m.closePrice = m.closePrice[1:]
	m.closePrice = append(m.closePrice, marketData.ClosePrice)

	m.openPrice = m.openPrice[1:]
	m.openPrice = append(m.openPrice, marketData.OpenPrice)

	m.highPrice = m.highPrice[1:]
	m.highPrice = append(m.highPrice, marketData.High)

	m.lowPrice = m.lowPrice[1:]
	m.lowPrice = append(m.lowPrice, marketData.Low)
}

func (m *BollingerBandsWithATR) CustomTakeProfitAndStopLoss()  {

}