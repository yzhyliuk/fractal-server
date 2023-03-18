package experimental

import (
	"fmt"
	"github.com/adshao/go-binance/v2/futures"
	"newTradingBot/api/database"
	"newTradingBot/indicators"
	"newTradingBot/logs"
	"newTradingBot/models/account"
	"newTradingBot/models/block"
	"newTradingBot/models/strategy"
	"newTradingBot/models/strategy/instance"
	"newTradingBot/models/users"
	"newTradingBot/strategies/common"
	"strconv"
)

type DF struct {
	Volume    float64
	Price     float64
	Direction string
	Type      string
}

const (
	UP      = "UP"
	DOWN    = "DOWN"
	NOTREND = "FLAT"
)

type experimentalContinuousStrategy struct {
	common.Strategy

	config      continuousConfig
	stopChannel chan struct{}

	priceObservations  []float64
	volumeObservations []float64

	highVolumeOrders []*DF

	trend          string
	lastPivotPrice float64

	volumeRatio        []float64
	buyersSellersRatio []float64

	balance float64
}

func NewContinuousExperimentalStrategy(config continuousConfig, keys *users.Keys, inst *instance.StrategyInstance) (strategy.Strategy, error) {
	acc, err := account.NewBinanceAccount(keys.ApiKey, keys.SecretKey, keys.ApiKey, keys.SecretKey)
	if err != nil {
		return nil, err
	}

	newStrategy := &experimentalContinuousStrategy{
		config: config,
	}
	newStrategy.Account = acc
	newStrategy.StopSignal = make(chan bool)
	newStrategy.StrategyInstance = inst
	newStrategy.ExperimentalHandler = newStrategy.Start

	newStrategy.priceObservations = make([]float64, 400)
	newStrategy.volumeObservations = make([]float64, 400)

	newStrategy.highVolumeOrders = make([]*DF, 50)

	newStrategy.volumeRatio = make([]float64, 50)
	newStrategy.buyersSellersRatio = make([]float64, 50)

	newStrategy.balance = 100

	return newStrategy, nil
}

func (e *experimentalContinuousStrategy) Start() {
	if e.StrategyInstance.Status == instance.StatusCreated {
		e.StrategyInstance.Status = instance.StatusRunning
		db, _ := database.GetDataBaseConnection()
		_ = instance.UpdateStatus(db, e.StrategyInstance.ID, instance.StatusRunning)
	}

	wsAggTradeHandler := func(event *futures.WsAggTradeEvent) {
		price, _ := strconv.ParseFloat(event.Price, 64)
		e.priceObservations = e.priceObservations[1:]
		e.priceObservations = append(e.priceObservations, price)

		quantity, _ := strconv.ParseFloat(event.Quantity, 64)
		e.volumeObservations = e.volumeObservations[1:]
		e.volumeObservations = append(e.volumeObservations, quantity)

		if e.priceObservations[0] == 0 {
			return
		}

		quantityMean := indicators.Average(e.volumeObservations)
		quantitySD := indicators.StandardDeviation(e.volumeObservations)

		up := quantityMean + quantitySD*3

		orderType := "LIMIT"
		if !event.Maker {
			orderType = "MARKET"
		}

		direction := DOWN
		// to do - price direction calculation for same price in row
		if e.priceObservations[len(e.priceObservations)-2] < price {
			direction = UP
		}

		if quantity > up {
			e.highVolumeOrders = e.highVolumeOrders[1:]
			e.highVolumeOrders = append(e.highVolumeOrders, &DF{
				Price:     price,
				Volume:    quantity,
				Type:      orderType,
				Direction: direction,
			})

			if e.highVolumeOrders[0] != nil {
				buyersVolume := 0.
				sellersVolume := 0.

				buyersCount := 0
				sellersCount := 0

				for _, order := range e.highVolumeOrders {
					if order.Direction == UP {
						buyersVolume += order.Volume
						buyersCount++
					} else {
						sellersVolume += order.Volume
						sellersCount++
					}
				}

				volumeRatio := buyersVolume / sellersVolume                        // more than 1 - more buyers volume
				buyersSellersRatio := float64(buyersCount) / float64(sellersCount) //more than 1 - more buyers

				e.buyersSellersRatio = e.buyersSellersRatio[1:]
				e.buyersSellersRatio = append(e.buyersSellersRatio, buyersSellersRatio)

				e.volumeRatio = e.volumeRatio[1:]
				e.volumeRatio = append(e.volumeRatio, volumeRatio)

				if e.volumeRatio[0] != 0 {
					meanVolumeRatio := indicators.Average(e.volumeRatio)
					meanCountRatio := indicators.Average(e.buyersSellersRatio)

					vSD := indicators.StandardDeviation(e.volumeRatio)
					cSD := indicators.StandardDeviation(e.buyersSellersRatio)

					upVolume := meanVolumeRatio + vSD*3
					donwVolume := meanVolumeRatio - vSD*3

					upCount := meanCountRatio + cSD*3
					donwCount := meanCountRatio - cSD*3

					if volumeRatio > upVolume && buyersSellersRatio > upCount {
						//buy
						if e.trend != UP {
							e.trend = UP
							e.PrintData(price, volumeRatio, buyersSellersRatio)

							e.CloseAllTrades()
							err := e.HandleBuy(&block.Data{ClosePrice: price})
							if err != nil {
								logs.LogError(err)
							}
						}
					} else if volumeRatio < donwVolume && buyersSellersRatio < donwCount {
						//sell
						if e.trend != DOWN {
							e.trend = DOWN
							e.PrintData(price, volumeRatio, buyersSellersRatio)

							e.CloseAllTrades()
							err := e.HandleSell(&block.Data{ClosePrice: price})
							if err != nil {
								logs.LogError(err)
							}
						}
					}
					//else {
					//	conditionOne := e.trend == UP && volumeRatio < meanVolumeRatio && buyersSellersRatio < meanCountRatio
					//	conditionTwo := e.trend == DOWN && volumeRatio > meanVolumeRatio && buyersSellersRatio > meanCountRatio
					//	if conditionOne || conditionTwo {
					//		e.CloseAllTrades()
					//	}
					//	if e.trend != NOTREND {
					//		e.trend = NOTREND
					//		e.PrintData(price, volumeRatio, buyersSellersRatio)
					//		e.CloseAllTrades()
					//	}
					//}

				}

				//if buyersVolume > sellersVolume && buyersCount > sellersCount {
				//	// Buy trend
				//	if e.trend != UP {
				//		e.trend = UP
				//		e.lastPivotPrice = price
				//
				//		fmt.Println(fmt.Sprintf("PRICE: %f TREND: %s \n BUERS: %d VOLUME: %f \n SEELLERS: %d VOLUME %f", price, e.trend, buyersCount, buyersVolume, sellersCount, sellersVolume))
				//
				//	}
				//} else if buyersVolume < sellersVolume && buyersCount < sellersCount {
				//	// Sell trend
				//	if e.trend != DOWN {
				//		e.trend = DOWN
				//		e.lastPivotPrice = price
				//
				//		fmt.Println(fmt.Sprintf("PRICE: %f TREND: %s \n BUERS: %d VOLUME: %f \n SEELLERS: %d VOLUME %f \n \n", price, e.trend, buyersCount, buyersVolume, sellersCount, sellersVolume))
				//	}
				//}
			}
		}
	}

	_, stopC, _ := futures.WsAggTradeServe(e.config.Pair, wsAggTradeHandler, errHandlerLog)

	e.stopChannel = stopC
}

func errHandlerLog(err error) {
	logs.LogDebug("", err)
}

func (e *experimentalContinuousStrategy) PrintData(price, volume, count float64) {
	fmt.Printf("================== \n")
	fmt.Printf("PRICE: %f TREND: %s \n Volume Ratio: %f Count Ratio: %f \n", price, e.trend, volume, count)
	fmt.Printf("================== \n \n")
}

func (e *experimentalContinuousStrategy) Stop() {

	db, err := database.GetDataBaseConnection()
	if err != nil {
		logs.LogDebug("", err)
	}
	err = instance.UpdateStatus(db, e.StrategyInstance.ID, instance.StatusStopped)
	if err != nil {
		logs.LogDebug("", err)
	}

	e.stopChannel <- struct{}{}
}
