package mlp

type MultiLayerPerceptronInstance struct {
	Id     int                           `json:"id" gorm:"column:id"`
	Name   string                        `json:"name" gorm:"column:name"`
	Struct string                        `json:"-" gorm:"column:struct"`
	UserID int                           `json:"userID" gorm:"column:user_id"`
	Model  *MultiLayerPerceptronInstance `json:"-" gorm:"-"`
}

func (m *MultiLayerPerceptronInstance) Load() {

}
