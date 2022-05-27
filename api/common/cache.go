package common

import (
	"log"
	"newTradingBot/models/account"
	"newTradingBot/models/apimodels"
)

var SpotTradingPairs = make([]apimodels.TradingPair,0)

func InitSpotTradingPairs() {
	pairs, err := account.GetAllPairsForBinance(false)
	if err != nil {
		log.Println(err)
	}

	for k, _ := range pairs {
		SpotTradingPairs = append(SpotTradingPairs, apimodels.TradingPair{Option: k, Value: k})
	}
}