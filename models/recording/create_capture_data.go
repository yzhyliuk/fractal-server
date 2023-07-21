package recording

import (
	"context"
	"errors"
	"fmt"
	"github.com/adshao/go-binance/v2/futures"
	"gorm.io/gorm"
	"newTradingBot/models/block"
	"strconv"
	"time"
)

const framesLimit = 1000
const maxBatchInsert = 4000

func GetBinanceData(db *gorm.DB, pair string, userID, timeframe, numberOfTimeFrames int) (*CapturedSession, error) {
	interval, err := formatTimeFrameToInterval(timeframe)
	if err != nil {
		return nil, err
	}

	timeframe = timeframe / 60
	timeframeDuration := time.Duration(timeframe)

	client := futures.NewClient("", "")

	dataSet := make([]*block.CapturedData, 0)

	current := 0
	max := numberOfTimeFrames

	now := time.Now()
	previous := now.Add(-time.Duration(now.Minute()%timeframe)*time.Minute - time.Duration(now.Second())*time.Second)
	previous = previous.Truncate(timeframeDuration * time.Minute)

	endTime := previous.Add(time.Duration(numberOfTimeFrames) * time.Minute * -timeframeDuration)

	toBeSavedStartTime := endTime.Format("15:04:05  02.01.06")
	toBeSavedEndTime := time.Now().Format("15:04:05  02.01.06")

	capture, err := CreateSession(db, &CapturedSession{
		UserID:    userID,
		TimeFrame: timeframe * 60,
		Symbol:    pair,
		Status:    StatusStopped,
		IsFutures: true,
		StartDate: &toBeSavedStartTime,
		EndDate:   &toBeSavedEndTime,
	})

	for current < max {

		startTime := endTime.Add((framesLimit) * (time.Minute * timeframeDuration))
		sts := endTime.UnixMilli()
		ets := startTime.UnixMilli()

		res, err := client.NewKlinesService().EndTime(ets).StartTime(sts).Symbol(pair).Limit(framesLimit).Interval(interval).Do(context.Background())
		if err != nil {
			return nil, err
		}

		for i := range res {
			closePrice, _ := strconv.ParseFloat(res[i].Close, 64)
			high, _ := strconv.ParseFloat(res[i].High, 64)
			low, _ := strconv.ParseFloat(res[i].Low, 64)
			openPrice, _ := strconv.ParseFloat(res[i].Open, 64)
			volume, _ := strconv.ParseFloat(res[i].Open, 64)
			tradesNum := res[i].TradeNum
			data := &block.Data{
				Symbol:      pair,
				ClosePrice:  closePrice,
				Low:         low,
				High:        high,
				OpenPrice:   openPrice,
				Volume:      volume,
				TradesArray: nil,
				TradesCount: int(tradesNum),
			}

			capturedData := &block.CapturedData{
				CaptureID: capture.ID,
				Data:      *data,
			}

			capturedData = capturedData.ConvertToDbObject()

			dataSet = append(dataSet, capturedData)
		}

		current += framesLimit
		endTime = startTime

		time.Sleep(time.Millisecond * 500)
	}

	counter := 0
	for counter < len(dataSet) {
		toSaveSlice := make([]*block.CapturedData, 0)
		if counter+maxBatchInsert < len(dataSet) {
			toSaveSlice = dataSet[counter : counter+maxBatchInsert]
		} else {
			toSaveSlice = dataSet[counter:]
		}

		err = db.Save(&toSaveSlice).Error
		if err != nil {
			return nil, err
		}

		counter += maxBatchInsert
	}

	capture.Status = StatusStopped
	capture.StartDate = &toBeSavedStartTime
	err = db.Save(&capture).Error
	if err != nil {
		return nil, err
	}

	return capture, nil
}

func formatTimeFrameToInterval(timeFrame int) (string, error) {
	defaultMark := "m"
	if timeFrame > 60 {
		// minutes
		timeFrame = timeFrame / 60
		if timeFrame > 60 {
			// hours
			timeFrame = timeFrame / 60
			defaultMark = "h"
		}

		return fmt.Sprintf("%d%s", timeFrame, defaultMark), nil
	}
	return "", errors.New("invalid time frame. Use 1m 3m 5m 15m frames")
}
