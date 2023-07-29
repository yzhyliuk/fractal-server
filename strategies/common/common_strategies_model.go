package common

import (
	"newTradingBot/api/database"
	"newTradingBot/logs"
	"newTradingBot/models/account"
	"newTradingBot/models/block"
	"newTradingBot/models/strategy/instance"
	"newTradingBot/models/testing"
	"newTradingBot/models/trade"
	"time"
)

type Strategy struct {
	StrategyInstance *instance.StrategyInstance
	Account          account.Account
	MonitorChannel   chan *block.Data
	StopSignal       chan bool
	LastTrade        *trade.Trade

	trades      []*trade.Trade
	TotalProfit float64

	prevMarketData *block.Data

	currentMarketData   *block.Data
	Stopped             bool
	HandlerFunction     func(marketData *block.Data)
	DataProcessFunction func(marketData *block.Data)
	DataLoadEndpoint    func() error
	ExperimentalHandler func()

	TakeProfitPrice float64
	StopLossPrice   float64
}

func (s *Strategy) Execute() {
	s.TotalProfit = 0
	s.trades = make([]*trade.Trade, 0)

	if s.StrategyInstance.Testing != testing.BackTest {
		s.UpdateLastPingTime()
		s.LivePriceMonitoring()
		if s.DataLoadEndpoint != nil {
			err := s.DataLoadEndpoint()
			if err != nil {
				logs.LogError(err)
			}
		}

		// Restore from cache
		if s.StrategyInstance.Status == instance.StatusRestarting {
			db, err := database.GetDataBaseConnection()
			if err != nil {
				return
			}
			err = RestoreFromCache(db, s)
			if err != nil {
				return
			}

			err = DeleteCache(db, s.StrategyInstance.ID)
			if err != nil {
				return
			}
		}
	}

	go func() {
		for {
			select {
			case <-s.StopSignal:
				return
			default:
				marketData := <-s.MonitorChannel
				if s.Stopped {
					return
				}

				s.currentMarketData = marketData

				s.CalculateTradeData(marketData)
				s.HandleStrategyDefinedStopLoss(marketData)

				if s.MaxLossPerStrategyCondition() {
					return
				}
				s.HandleTPansSL(marketData)

				s.DataProcessFunction(marketData)
				s.HandlerFunction(marketData)

				s.prevMarketData = marketData

				//logs.LogDebug(fmt.Sprintf("Data received by instance #%d", s.StrategyInstance.ID), nil)
				if s.StrategyInstance.Testing != testing.BackTest {
					s.UpdateLastPingTime()
				}

			}
		}
	}()
}

func (s *Strategy) UpdateLastPingTime() {
	db, err := database.GetDataBaseConnection()
	if err != nil {
		logs.LogError(err)
		return
	}

	err = db.Model(&instance.StrategyInstance{}).Where("id = ?", s.StrategyInstance.ID).Update("last_ping_time", time.Now()).Error
	if err != nil {
		logs.LogError(err)
		return
	}
}

// ChangeBid - changes bid for strategy
func (s *Strategy) ChangeBid(bid float64) error {
	s.StrategyInstance.Bid = bid
	db, err := database.GetDataBaseConnection()
	if err != nil {
		return err
	}

	return db.Save(s.StrategyInstance).Error
}

func (s *Strategy) CloseTrade() {
	if s.LastTrade != nil {
		s.CloseAllTrades()
	}
}

// ExecuteExperimental - runs experimental handler
func (s *Strategy) ExecuteExperimental() {
	s.ExperimentalHandler()
}

// GetInstance - returns instance entity from strategy
func (s *Strategy) GetInstance() *instance.StrategyInstance {
	return s.StrategyInstance
}

// GetTestingTrades - returns trades from back testing
func (s *Strategy) GetTestingTrades() []*trade.Trade {
	return s.trades
}

// PrepareForRestart - prepares strategy instance to restart
func (s *Strategy) PrepareForRestart() {
	db, err := database.GetDataBaseConnection()
	if err != nil {
		logs.LogDebug("", err)
	}
	err = instance.UpdateStatus(db, s.StrategyInstance.ID, instance.StatusRestarting)
	if err != nil {
		logs.LogDebug("", err)
	}

	err = SaveCache(db, s)
	if err != nil {
		// Stop if not able to save cache
		s.Stop()
		return
	}

	s.Stopped = true
	go func() { s.StopSignal <- true }()
}

// Stop - terminates running of current strategy and closes all the trades
func (s *Strategy) Stop() {
	db, err := database.GetDataBaseConnection()
	if err != nil {
		logs.LogDebug("", err)
	}
	err = instance.UpdateStatus(db, s.StrategyInstance.ID, instance.StatusStopped)
	if err != nil {
		logs.LogDebug("", err)
	}

	s.Stopped = true
	go func() { s.StopSignal <- true }()

	s.CloseAllTrades()
}
