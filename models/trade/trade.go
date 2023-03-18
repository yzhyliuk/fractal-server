package trade

import (
	"github.com/adshao/go-binance/v2/futures"
	"gorm.io/gorm"
	"time"
)

const TradesTableName = "trades"
const StatusActive = "active"
const StatusClosed = "closed"

const BinanceFuturesTakerFeeRate = 0.0004
const BinanceSpotTakerFee = 0.00075

// Trade - represents trade instance
type Trade struct {
	ID             int              `json:"id" gorm:"column:id"`
	Pair           string           `json:"pair" gorm:"column:pair"`
	StrategyID     int              `json:"strategyID" gorm:"column:strategy_id"`
	InstanceID     int              `json:"instanceID" gorm:"column:instance_id"`
	UserID         int              `json:"userID" gorm:"column:user_id"`
	IsFutures      bool             `json:"isFutures" gorm:"column:is_futures"`
	PriceOpen      float64          `json:"priceOpen" gorm:"column:price_open"`
	PriceClose     float64          `json:"priceClose" gorm:"column:price_close"`
	USD            float64          `json:"usd" gorm:"column:usd"`
	Quantity       float64          `json:"quantity" gorm:"column:quantity"`
	Profit         float64          `json:"profit" gorm:"column:profit"`
	ROI            float64          `json:"roi" gorm:"column:roi"`
	Status         string           `json:"status" gorm:"column:status"`
	Leverage       *int             `json:"leverage" gorm:"column:leverage"`
	TimeStamp      time.Time        `json:"-" gorm:"column:time_stamp"`
	TimeString     string           `json:"time" gorm:"-"`
	BinanceOrderID int64            `json:"-" gorm:"-"`
	FuturesSide    futures.SideType `json:"futuresSide" gorm:"column:futures_side"`
	LengthCounter  int              `json:"lengthCounter" gorm:"column:length_counter"`
	MaxDropDown    float64          `json:"maxDropDown" gorm:"column:max_drop"`
	MaxHeadUp      float64          `json:"maxHeadUp" gorm:"column:max_up"`
}

func (t *Trade) TableName() string {
	return TradesTableName
}

func (t *Trade) ConvertTime() {
	t.TimeString = t.TimeStamp.Format("15:04:05  02.01.06")
}

// NewTrade - creates new trade for given user with params
func NewTrade(db *gorm.DB, trade Trade) (*Trade, error) {
	err := db.Create(&trade).Error
	if err != nil {
		return nil, err
	}
	return &trade, nil
}

// CloseTrade - closes trade with price
func CloseTrade(db *gorm.DB, trade *Trade) error {
	err := db.Save(&trade).Error
	if err != nil {
		return err
	}
	return err
}

// GetTradesForUser - returns list of trades for given user
func GetTradesForUser(db *gorm.DB, userID int) ([]*Trade, error) {
	var trades []*Trade
	err := db.Where("user_id = ?", userID).Find(&trades).Error
	if err != nil {
		return nil, err
	}
	return trades, nil
}

// GetTradesByInstanceID - returns trades list for given instance
func GetTradesByInstanceID(db *gorm.DB, userID, instanceID int) ([]*Trade, error) {
	var trades []*Trade
	err := db.Where("user_id = ? AND instance_id = ?", userID, instanceID).Order("id desc").Find(&trades).Error
	if err != nil {
		return nil, err
	}
	return trades, nil
}

func (t *Trade) CalculateProfitRoi() {
	roi := 0.
	profit := 0.
	fee := t.USD * BinanceFuturesTakerFeeRate

	if t.IsFutures {
		switch t.FuturesSide {
		case futures.SideTypeBuy:
			profit = (t.Quantity * t.PriceClose) - t.USD
		case futures.SideTypeSell:
			profit = (t.Quantity * t.PriceOpen) - (t.Quantity * t.PriceClose)
		}
	} else {
		profit = (t.Quantity * t.PriceClose) - t.USD
	}

	profit -= 2 * fee

	if t.IsFutures {
		roi = profit / (t.USD / float64(*t.Leverage))
	} else {
		roi = profit / t.USD
	}

	t.Profit = profit
	t.ROI = roi
}
