package activation_funcs

import "gorm.io/gorm"

type ActivationFunction struct {
	Name string `json:"name" gorm:"column:name"`
	Key  string `json:"key" gorm:"column:key"`
}

const ActivationFunctionsTableName = "activations"

func (a *ActivationFunction) TableName() string {
	return ActivationFunctionsTableName
}

func GetActivationFunctions(db *gorm.DB) ([]ActivationFunction, error) {
	var activations []ActivationFunction
	err := db.Find(&activations).Error
	return activations, err
}
