package apimodels

type NewStrategyConfig struct {
	Name string `json:"name"`
	Config interface{} `json:"config"`
}

type CommonStrategyConfig struct {
	Pairs []string `json:"pairs"`
	Config map[string]interface{} `json:"config"`
}

type ChangeConfig struct {
	Bid float64 `json:"bid"`
	InstanceID int `json:"instanceId"`
}