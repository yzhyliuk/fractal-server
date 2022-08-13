package recording

import (
	"gorm.io/gorm"
	"time"
)

const CapturedSessionTableName = "captured_session"

const StatusStopped = "stopped"
const StatusRecording = "recording"

type CapturedSession struct {
	ID int `json:"id" gorm:"column:id"`
	UserID int `json:"userId" gorm:"column:userid"`
	Symbol string `json:"symbol" gorm:"column:symbol"`
	TimeFrame int `json:"timeFrame" gorm:"column:timeframe"`
	StartDate *string `json:"startDate" gorm:"column:startdate"`
	EndDate *string `json:"endDate" gorm:"column:enddate"`
	Status string `json:"status" gorm:"column:status"`
	IsFutures bool `json:"isFutures" gorm:"column:isfutures"`
}

func (c *CapturedSession) TableName() string {
	return CapturedSessionTableName
}

func GetAllSessionForUser(db *gorm.DB, userID int) ([]*CapturedSession, error) {
	var sessions []*CapturedSession
	err := db.Where("userId = ?", userID).Find(&sessions).Error
	if err != nil {
		return nil, err
	}

	return sessions, nil
}

func CreateSession(db *gorm.DB, session *CapturedSession) (*CapturedSession, error) {
	session.Status = StatusRecording
	timeStamp := time.Now().Format("15:04:05  02.01.06")
	session.StartDate = &timeStamp
	err := db.Create(session).Error
	if err != nil {
		return nil, err
	}

	return session, err
}