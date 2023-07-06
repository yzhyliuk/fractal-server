package apimodels

type CreateTestingDataModel struct {
	Pair               string `json:"pair"`
	TimeFrame          int    `json:"timeFrame"`
	NumberOfTimeFrames int    `json:"numberOfTimeFrames"`
}
