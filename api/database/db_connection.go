package database

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const dsn = "host=localhost user=postgres password=080919 dbname=fractal_ai port=5432 sslmode=disable"

var db *gorm.DB

func GetDataBaseConnection() (*gorm.DB, error)  {
	if db == nil {
		dbConnection, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			return nil, err
		}
		db = dbConnection
	}
	return db, nil
}
