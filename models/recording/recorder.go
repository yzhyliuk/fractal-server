package recording

import (
	"newTradingBot/api/database"
	"newTradingBot/logs"
	"newTradingBot/models/block"
	"time"
)

func NewRecorder(inputChannel chan *block.Data, captureSession *CapturedSession) *Recorder {
	return &Recorder{
		Session: captureSession,
		inputChannel: inputChannel,
		stopChannel: make(chan struct{}),
	}
}

type Recorder struct {
	Session *CapturedSession
	inputChannel chan *block.Data
	stopChannel chan struct{}
}

func (r *Recorder) Start() error {
	go func() {
		for  {
			select {
			case <-r.stopChannel:
				return
			default:
				marketData := <- r.inputChannel
				capturedData := &block.CapturedData{
					CaptureID: r.Session.ID,
				}
				capturedData.Data = *marketData

				db, err := database.GetDataBaseConnection()
				if err != nil {
					logs.LogDebug("",err)
				}

				err = db.Create(capturedData.ConvertToDbObject()).Error
				if err != nil {
					logs.LogDebug("",err)
				}
			}
		}
	}()
	return nil
}

func (r *Recorder )Stop() error {
	go func() {
		r.stopChannel <- struct{}{}
	}()

	db, err := database.GetDataBaseConnection()
	if err != nil {
		logs.LogDebug("", err)
		return err
	}
	r.Session.Status = StatusStopped
	timeStamp := time.Now().Format("15:04:05  02.01.06")
	r.Session.EndDate = &timeStamp
	err = db.Save(r.Session).Error
	if err != nil {
		logs.LogDebug("", err)
		return err
	}

	return nil
}
