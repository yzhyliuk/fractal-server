package apimodels

import "gorm.io/gorm"

const tableName = "strategy_fields"

type StrategyField struct {
	ID           int     `json:"id" gorm:"column:id"`
	StrategyID   int     `json:"strategyId" gorm:"column:strategy_id"`
	Name         string  `json:"name" gorm:"column:name"`
	DisplayName  string  `json:"displayName" gorm:"column:display_name"`
	Description  *string `json:"description" gorm:"column:description"`
	Min          *int    `json:"min" gorm:"column:min"`
	Max          *int    `json:"max" gorm:"column:max"`
	DefaultValue string  `json:"defaultValue" gorm:"column:default_value"`
	Type         string  `json:"type" gorm:"column:type"`
	UiType       string  `json:"uiType" gorm:"column:ui_type"`
	Dataset      string  `json:"dataset" gorm:"column:dataset"`
	FuturesOnly  bool    `json:"futuresOnly" gorm:"column:futuresonly"`
}

func (s StrategyField) TableName() string {
	return tableName
}

func GetStrategyFields(db *gorm.DB, strategyID int) ([]*StrategyField, error) {
	var fields []*StrategyField
	err := db.Where("strategy_id = ?", strategyID).Order("id").Find(&fields).Error
	if err != nil {
		return nil, err
	}

	return fields, nil
}

func GetDefaultFields(db *gorm.DB) ([]*StrategyField, error) {
	return GetStrategyFields(db, 0)
}
