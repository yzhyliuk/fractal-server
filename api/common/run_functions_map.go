package common

import (
	"newTradingBot/strategies/nadaraya-watsons"
)

var RunStrategy = map[int]func(userID int, rawConfig []byte) error{
	1: nadaraya_watsons.RunMovingAverageCrossover,
}
