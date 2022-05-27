package instance

import "gorm.io/gorm"

const dataFrameTableName = "instance_data"

type DataFrame struct {
	InstanceID int `json:"instance_id"`
	Data []byte `json:"data"`
}

func (d *DataFrame) TableName() string {
	return dataFrameTableName
}

// GetDataFrames - returns dataFrames for given strategy instance
func GetDataFrames(db *gorm.DB, instanceID int) ([]*DataFrame, error) {
	var dataFrames []*DataFrame
	err := db.Where("instance_id = ?", instanceID).Find(&dataFrames).Error
	if err != nil {
		return nil, err
	}
	return dataFrames, nil
}

func NewDataFrame(db *gorm.DB, frame *DataFrame) error {
	return db.Create(frame).Error
}
