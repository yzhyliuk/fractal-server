package replay

import (
	"newTradingBot/api/database"
	"newTradingBot/models/block"
)

type MonitorReplay struct {
	OutputChannel chan *block.Data
	sessionID int
}

func NewMonitorReplay(sessionID int) *MonitorReplay {
	return &MonitorReplay{
		OutputChannel: make(chan *block.Data),
		sessionID: sessionID,
	}
}

func (m *MonitorReplay) Start() error {
	data, err := m.GetMonitorData()
	if err != nil {
		return err
	}

	for _, frame := range data {
		m.OutputChannel <- frame.ExtractData()
	}

	return nil
}

func (m *MonitorReplay) GetMonitorData() ([]*block.CapturedData, error) {
	db, err := database.GetDataBaseConnection()
	if err != nil {
		return nil, err
	}

	var marketData []*block.CapturedData

	err = db.Where("captureid = ?", m.sessionID).Order("id asc").Find(&marketData).Error
	if err != nil {
		return nil, err
	}

	return marketData, nil
}
