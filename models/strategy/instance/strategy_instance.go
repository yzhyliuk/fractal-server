package instance

import "gorm.io/gorm"

const strategyInstancesTableName = "strategy_instances"
const strategyMonitoringTableName = "strategy_instance_monitoring"

const StatusRunning = "running"
const StatusCreated = "created"
const StatusStopped = "stopped"

// StrategyInstance - represents running instance of strategy
type StrategyInstance struct {
	ID int `json:"id" gorm:"column:id"`
	StrategyID int `json:"strategyId" gorm:"column:strategy_id"`
	UserID int `json:"userId" gorm:"column:user_id"`
	Pair string `json:"pair" gorm:"column:pair"`
	Bid float64 `json:"bid" gorm:"column:bid"`
	TimeFrame int `json:"timeFrame" gorm:"column:time_frame"`
	Status string `json:"status" gorm:"column:status"`
	IsFutures bool `json:"isFutures" gorm:"column:is_futures"`
	Leverage *int `json:"leverage" gorm:"column:leverage"`
}

type StrategyMonitoring struct {
	*StrategyInstance
	Name string `json:"name" gorm:"name"`
	Profit float64 `json:"profit" gorm:"column:profit"`
	WinRate float64 `json:"winRate" gorm:"column:win_rate" `
}

func (s StrategyInstance) TableName() string  {
	return strategyInstancesTableName
}

func (s StrategyMonitoring) TableName() string  {
	return strategyMonitoringTableName

}

// CreateStrategyInstance - creates new instance of strategy
func CreateStrategyInstance(db *gorm.DB, strategyInstance *StrategyInstance) (*StrategyInstance, error)  {
	err := db.Create(&strategyInstance).Error
	if err != nil {
		return nil, err
	}
	return strategyInstance, err
}


// ListInstancesForUser - returns list of instances for given user
func ListInstancesForUser(db *gorm.DB, userID int) ([]*StrategyMonitoring, error) {
	var instances []*StrategyMonitoring
	err := db.Where("user_id = ?", userID).Find(&instances).Error
	if err != nil {
		return nil, err
	}
	return instances, err
}

func GetInstanceByID(db *gorm.DB, id int) (*StrategyMonitoring, error)  {
	var instance *StrategyMonitoring
	err := db.Where("id = ?", id).Find(&instance).Error
	if err != nil {
		return nil, err
	}
	return instance, nil
}

// DeleteInstance - deletes given instance for good
func DeleteInstance(db *gorm.DB, instanceID, userID int) error {
	err := db.Delete(&StrategyInstance{
		ID: instanceID,
		UserID: userID,
	}).Error
	return err
}

// UpdateStatus - updates status of running instance
func UpdateStatus(db *gorm.DB, instanceID int, status string) error {
	err := db.Table(strategyInstancesTableName).Where("id = ?", instanceID).Update("status",status).Error
	if err != nil {
		return err
	}
	return nil
}
