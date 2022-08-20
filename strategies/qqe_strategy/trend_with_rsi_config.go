package qqe

import "newTradingBot/models/strategy/configs"

type QQEStrategyConfig struct {
	configs.BaseStrategyConfig
	RSIPeriod int `json:"rsiPeriod"`
	RSISmoothing int `json:"rsiSmoothing"`
	FastQQE float64 `json:"fastQQE"`
	Threshold int `json:"threshold"`
}

