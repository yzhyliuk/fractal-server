package common

import (
	"newTradingBot/strategies/glide_on_price"
	"newTradingBot/strategies/mac"
)

var RunStrategy = map[int]func(userID int, rawConfig []byte) error{
	1: mac.RunMovingAverageCrossover,
	2: glide_on_price.RunGlideOnPrice,
}
