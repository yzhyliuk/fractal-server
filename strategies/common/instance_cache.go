package common

import (
	"gorm.io/gorm"
)

type InstanceCache struct {
	ID              int     `gorm:"column:instance_id"`
	TakeProfitPrice float64 `gorm:"column:last_take_profit_price"`
	StopLossPrice   float64 `gorm:"column:last_stop_loss_price"`
}

// SaveCache - saves some field of strategy to database
func SaveCache(db *gorm.DB, strategy *Strategy) error {
	inst := &InstanceCache{
		ID:              strategy.StrategyInstance.ID,
		TakeProfitPrice: strategy.TakeProfitPrice,
		StopLossPrice:   strategy.StopLossPrice,
	}

	return db.Create(&inst).Error
}

// RestoreFromCache - restores fields from strategy
func RestoreFromCache(db *gorm.DB, strategy *Strategy) error {
	inst := &InstanceCache{
		ID: strategy.StrategyInstance.ID,
	}
	err := db.Find(inst).Error
	if err != nil {
		return err
	}

	strategy.StopLossPrice = inst.StopLossPrice
	strategy.TakeProfitPrice = inst.StopLossPrice
	return nil
}

// DeleteCache - deletes cache from db
func DeleteCache(db *gorm.DB, instanceID int) error {
	return db.Delete(&InstanceCache{ID: instanceID}).Error
}
