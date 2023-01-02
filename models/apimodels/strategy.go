package apimodels

import "gorm.io/gorm"

const strategyTableName = "strategy_info"

type StrategyInfo struct {
	ID int `json:"id" gorm:"column:id"`
	Name string `json:"name" gorm:"column:"`
	Description string `json:"description" gorm:"column:description"`
	IsContinuous bool `json:"isContinuous" gorm:"column:is_continuous"`
	IsHidden bool `json:"isHidden" gorm:"column:is_hidden"`
	Disabled bool `json:"disabled" gorm:"column:disabled"`
	StrategyName string `json:"strategy_name" gorm:"column:strategy_name"`
}

func (s *StrategyInfo) TableName() string  {
	return strategyTableName
}

func GetAllStrategies(db *gorm.DB) ([]*StrategyInfo, error){
	var strategies []*StrategyInfo
	err := db.Where("id != 0").Find(&strategies).Error
	return strategies, err
}