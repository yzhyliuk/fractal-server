package block

import (
	"github.com/lib/pq"
	"gorm.io/gorm"
	"newTradingBot/api/database"
	"time"
)

const CapturedDataTableName = "captured_data"

// Data - represents market info block
type Data struct {
	Symbol string `json:"symbol" gorm:"column:symbol"`

	TradesCount int           `json:"tradesCount" gorm:"column:tradescount"`
	Time        time.Duration `json:"-" gorm:"-"`

	Volume float64 `json:"volume" gorm:"column:volume"`

	High float64 `json:"high" gorm:"column:high"`
	Low  float64 `json:"low" gorm:"column:low"`

	OpenPrice  float64 `json:"openPrice" gorm:"column:openprice"`
	ClosePrice float64 `json:"closePrice" gorm:"column:closeprice"`

	AveragePrice float64 `json:"averagePrice" gorm:"column:averageprice"`

	// slice of all trades price for given time frame
	TradesArray []float64 `json:"trades" gorm:"-"`
}

// CapturedData - uses for strategy testing
type CapturedData struct {
	ID int `json:"id" gorm:"column:id"`
	Data
	CaptureID int             `json:"captureId" gorm:"column:captureid"`
	Trades    pq.Float64Array `json:"-" gorm:"type:numeric[]"`
}

func (c *CapturedData) ExtractData() *Data {
	dat := c.Data
	dat.TradesArray = c.Trades
	return &dat
}

func (c *CapturedData) ConvertToDbObject() *CapturedData {
	c.Trades = pq.Float64Array(c.TradesArray)
	return c
}

func (c *CapturedData) TableName() string {
	return CapturedDataTableName
}

func GetCaptureDataBySessionId(sessionID int, db *gorm.DB) ([]CapturedData, error) {
	db, err := database.GetDataBaseConnection()
	if err != nil {
		return nil, err
	}

	var marketData []CapturedData

	err = db.Where("captureid = ?", sessionID).Order("id asc").Find(&marketData).Error
	if err != nil {
		return nil, err
	}

	return marketData, nil
}
