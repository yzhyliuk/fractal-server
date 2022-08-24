package notifications

import (
	"gorm.io/gorm"
	"newTradingBot/models/users"
	"time"
)

type NotifyType int

var Warning NotifyType = 1
var Info NotifyType = 2
var Error NotifyType = 3

const TableName = "notifications"

type Notification struct {
	ID int `json:"id" gorm:"column:id"`
	UserID int `json:"userId" gorm:"column:user_id"`
	TimeStamp time.Time `json:"-" gorm:"column:time_stamp"`
	Message string `json:"message" gorm:"column:message"`
	Type NotifyType `json:"type" gorm:"column:type"`
	Time string `json:"time" gorm:"-"`
}

func (n *Notification) TableName() string {
	return TableName
}

func (n *Notification) FormatDate()  {
	n.Time = n.TimeStamp.Format("15:04     02.01.06")
}

func CreateUserNotification(db *gorm.DB, userID int, notifyType NotifyType, message string) error {
	var notification = Notification{
		UserID: userID,
		TimeStamp: time.Now(),
		Message: message,
		Type: notifyType,
	}

	return db.Create(&notification).Error
}

func CreateGeneralNotification(db *gorm.DB, nType NotifyType, message string) error {
	var notification = Notification{
		UserID: 0,
		TimeStamp: time.Now(),
		Message: message,
		Type: nType,
	}

	list, err := users.ListAllUsers(db)
	if err != nil {
		return err
	}

	for _, u := range list {
		notification.UserID = u.ID
		err = db.Create(&notification).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func ListNotificationsForUser(db *gorm.DB, userID int) ([]*Notification, error) {
	var Notifications []*Notification
	err := db.Where("user_id = 0 OR user_id = ?", userID).Order("time_stamp desc").Find(&Notifications).Error
	for _, n := range Notifications {
		n.FormatDate()
	}
	return Notifications, err
}

func DismissAll(db *gorm.DB, userID int) error {
	return db.Where("user_id = ?", userID).Delete(&Notification{}).Error
}