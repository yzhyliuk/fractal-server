package users

import (
	"fmt"
	"gorm.io/gorm"
)

const keysTableName = "keys"

// Keys - represents api keys for given user
type Keys struct {
	UserID int `gorm:"column:user_id"`
	ApiKey string `json:"apiKey" gorm:"column:api_key"`
	SecretKey string `json:"secretKey" gorm:"column:secret_key"`
}

func (k *Keys) TableName() string  {
	return keysTableName
}

func SaveUserKeys(db *gorm.DB, keys *Keys) error {
	return db.Where("user_id = ?", keys.UserID).Save(keys).Error
}

func GetUserKeys(db *gorm.DB, userID int) (Keys, error)  {
	var userKeys Keys
	err :=  db.Where("user_id = ?", userID).Find(&userKeys).Error
	if err != nil {
		return userKeys, err
	}

	return userKeys, nil
}

func (k *Keys) HideKeys()  {
	if k.ApiKey != "" && k.SecretKey != "" {
		k.ApiKey = fmt.Sprintf("%s***********%s", k.ApiKey[:5], k.ApiKey[len(k.ApiKey)-6:])
		k.SecretKey = fmt.Sprintf("%s**********%s", k.SecretKey[:5], k.SecretKey[len(k.SecretKey)-6:])
	}
}