package apimodels

type BackTestModel struct {
	Config interface{} `json:"config"`
	StrategyID int `json:"strategyId"`
	CaptureID int `json:"sessionId"`
}
