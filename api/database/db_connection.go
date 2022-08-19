package database

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"newTradingBot/configuration"
	"os"
)

const devDSN = "host=localhost user=postgres password=080919 dbname=fractal_ai port=5432 sslmode=disable"
const prodDebug = "host=164.92.143.224 user=postgres password=08091908fekla dbname=fractal_db port=1234 sslmode=disable"

var db *gorm.DB

func GetDataBaseConnection() (*gorm.DB, error)  {
	if db == nil {
		conn := devDSN
		if configuration.IsProduction() {
			conn = dbURL()
		} else if configuration.IsDebugProd() {
			conn = prodDebug
		}
		dbConnection, err := gorm.Open(postgres.Open(conn), &gorm.Config{})
		if err != nil {
			return nil, err
		}
		db = dbConnection
	}
	return db, nil
}

func dbURL() string {
	dbHost := os.Getenv("DATABASE_HOST")
	dbPort := os.Getenv("DATABASE_PORT")
	dbName := os.Getenv("POSTGRES_DB")
	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")

	return fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable", dbUser, dbPassword, dbName, dbHost, dbPort)
}
