package main

import (
	"log"
	"newTradingBot/api"
	"newTradingBot/api/common"
	"newTradingBot/logs"
	"newTradingBot/storage"
	"os"
	"os/signal"
)

const testnetApi = "wNjHUt25VGuUbtg5xSPWEWhcXnKvDRb7MKfFpGF4VBIyFzoN8LrxcDKcZuzt42Rp"
const testnetSecret = "bokQfryzyvDUmvdocBd0T6jpMiWLzK5o3mB4sfGSRQX6a9GQbXl1P8uB5WaDFEHA"

const futuresTestnetApi = "1727c6deddc958a983d169d20cad955a9896cb5a8bc6e405b99517ecae43a4ee"
const futuresTestnetSecret = "d0cb5106d2f17f7cb37cf4ec6568c9687ba76140fce0343e6d20a65395cedc99"

const realApi = "SNbfHhaCv47GdosKEuuD9nZVdG2uTJIgVdVkfeY4eIAyLiBawEfxfdPclOwXhbIw"
const realSecretKey = "BL7kS7hN7ry7DWzsoSSsJkQziI5yGBdsSZRyYCHzTO6jkKtDkSm5Z1mcbBOPBy47"

func main()  {

	c := make(chan os.Signal,1)
	signal.Notify(c, os.Interrupt)
	go func() {
		_ = <-c
		logs.LogDebug("Gracefull shutdown", nil)
		terminateAllStrategies()
	}()

	common.InitSpotTradingPairs()
	common.InitFuturesTradingPairs()

	// SERVER
	_, err := api.StartServer()
	if err != nil {
		log.Fatal(err)
	}

	//acc, _ := account.NewBinanceAccount(realApi, realSecretKey ,realApi, realSecretKey)
	//order, err := acc.OpenFuturesPosition(0.1, "BNBUSDT",futures.SideTypeSell, nil)
	//result, err := acc.CloseFuturesPosition(order)
	//fmt.Println(err,result)

}

func terminateAllStrategies()  {
	logs.LogDebug("Terminating all strategies...", nil)
	for _, v := range storage.StrategiesStorage {
		v.Stop()
	}
	logs.LogDebug("All strategies are terminated.", nil)
	os.Exit(0)
}