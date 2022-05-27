package apimodels

import "gorm.io/gorm"

const strategyTableName = "strategy_info"

type StrategyInfo struct {
	ID int `json:"id" gorm:"column:id"`
	Name string `json:"name" gorm:"column:"`
	Description string `json:"description" gorm:"column:description"`
}

func (s *StrategyInfo) TableName() string  {
	return strategyTableName
}

func GetAllStrategies(db *gorm.DB) ([]*StrategyInfo, error){
	var strategies []*StrategyInfo
	err := db.Where("id != 0").Find(&strategies).Error
	return strategies, err
}