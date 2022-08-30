package configs

import (
	"gorm.io/gorm"
)

type SavedConfig struct {
	ID int `json:"id" gorm:"column:id"`
	StrategyID int `json:"strategyId" gorm:"column:strategy_id"`
	UserID int `json:"userId" gorm:"column:user_id"`
	Name string `json:"name" gorm:"column:name"`
	Config []byte `json:"-" gorm:"column:config"`
	ConfigParsed string `json:"config" gorm:"-"`
}

const SavedConfigTableName = "saved_config"

func (s *SavedConfig) TableName() string {
	return SavedConfigTableName
}

func (s *SavedConfig) PrepareConfig()  {
	s.ConfigParsed = string(s.Config)
}

func CreateConfig(db *gorm.DB, config []byte, userID, strategyID int, name string) error {
	var cfg *SavedConfig

	_ = db.Where("name = ? AND strategy_id = ? AND user_id = ?", name, strategyID, userID).Find(&cfg)
	if cfg.ID != 0 {
		cfg.Config = config
		return db.Save(&cfg).Error
	}

	cfg = &SavedConfig{
		StrategyID: strategyID,
		UserID: userID,
		Name: name,
		Config: config,
	}

	return db.Create(cfg).Error
}

func DeleteSavedConfig(db *gorm.DB, configID, userID int) error {
	return db.Where("user_id = ? AND id = ?", userID, configID).Delete(&SavedConfig{}).Error
}

func GetConfigsForStrategyPerUser(db *gorm.DB, userID, strategyID int) ([]*SavedConfig, error) {
	cfgs := make([]*SavedConfig, 0)
	err := db.Where("user_id = ? AND strategy_id = ?", userID, strategyID).Find(&cfgs).Error
	if err != nil {
		return nil, err
	}

	for _, c := range cfgs {
		c.PrepareConfig()
	}

	return cfgs, nil
}
