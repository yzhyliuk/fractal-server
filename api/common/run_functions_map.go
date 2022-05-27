package common

import "newTradingBot/strategies/mac"

var RunStrategy = map[int]func(userID int, rawConfig []byte) error{
	1: mac.RunMovingAverageCrossover,
}
