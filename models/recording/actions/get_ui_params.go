package actions

import (
	"gorm.io/gorm"
	"newTradingBot/models/recording"
)

type CapturesParams struct {
	TimeFrames []int    `json:"timeFrames"`
	Pairs      []string `json:"pairs"`
}

func GetUsersDataOptions(db *gorm.DB, userID int) (*CapturesParams, error) {
	var capParams CapturesParams
	tf := make([]int, 0)
	sm := make([]string, 0)
	err := db.Raw("SELECT DISTINCT(timeframe) FROM public.captured_session WHERE userid IN ((SELECT user_origin FROM public.allowed_users WHERE user_child = ? AND allow_use_captures = true), ?)", userID, userID).Find(&tf).Error
	if err != nil {
		return nil, err
	}

	err = db.Raw("SELECT DISTINCT(symbol) FROM public.captured_session WHERE userid IN ((SELECT user_origin FROM public.allowed_users WHERE user_child = ? AND allow_use_captures = true), ?)", userID, userID).Find(&sm).Error
	if err != nil {
		return nil, err
	}

	capParams.Pairs = sm
	capParams.TimeFrames = tf
	return &capParams, nil
}

func GetCaptures(db *gorm.DB, pair string, timeFrame, userID int) ([]*recording.CapturedSession, error) {
	captures := make([]*recording.CapturedSession, 0)

	userIDs := make([]int, 0)

	err := db.Raw("SELECT user_origin FROM public.allowed_users WHERE user_child = ? AND allow_use_captures = true", userID).Find(&userIDs).Error
	if err != nil {
		return nil, err
	}

	userIDs = append(userIDs, userID)

	if pair != "all" {
		db = db.Where("symbol = ?", pair)
	}

	if timeFrame != 0 {
		db = db.Where("timeframe = ?", timeFrame)
	}

	db = db.Where("userid IN (?)", userIDs)

	err = db.Find(&captures).Error
	if err != nil {
		return nil, err
	}

	return captures, nil

}
