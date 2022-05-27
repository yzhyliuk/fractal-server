package main

import (
	"log"
	"newTradingBot/api"
	"newTradingBot/api/common"
)

const testnetApi = "wNjHUt25VGuUbtg5xSPWEWhcXnKvDRb7MKfFpGF4VBIyFzoN8LrxcDKcZuzt42Rp"
const testnetSecret = "bokQfryzyvDUmvdocBd0T6jpMiWLzK5o3mB4sfGSRQX6a9GQbXl1P8uB5WaDFEHA"

const futuresTestnetApi = "1727c6deddc958a983d169d20cad955a9896cb5a8bc6e405b99517ecae43a4ee"
const futuresTestnetSecret = "d0cb5106d2f17f7cb37cf4ec6568c9687ba76140fce0343e6d20a65395cedc99"

const realApi = "BacybA8f4Tw6yBZ7Ot782tGeHOZVxHxxbEK6dvr4jA2QsnG7rvKc2GiWZA2v2Kex"
const realSecretKey = "7UHbxI33kpEQikhIYXpzbSzBT3885BJSd0gCjxAg1rIEWwGmwQ7oARGw1Har1Syo"

func main()  {

	common.InitSpotTradingPairs()

	// SERVER
	log.Fatal(api.StartServer())

	//acc, err := account.NewBinanceAccount(realApi,realSecretKey, realApi, realSecretKey)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//
	//handler := func(block *block.Block, orderSpot *binance.CreateOrderResponse, orderFutures *futures.CreateOrderResponse, err error){
	//	if err == nil {
	//		if orderFutures != nil {
	//			fmt.Println(fmt.Sprintf("%s Position opened. \n Current price: %f \t %t",orderFutures.Side, block.ClosePrice,time.Now()))
	//		} else  {
	//			fmt.Println(fmt.Sprintf("%s Position opened. \n Current price: %f \t %t",orderSpot.Side, block.ClosePrice,time.Now()))
	//		}
	//	} else {
	//		fmt.Println("ERROR: ", err.Error())
	//	}
	//}
	//
	//monitorTimeFrame := time.Minute * 5
	//observationsTimeFrame := time.Minute * 5
	//symbol := "NEARUSDT"
	//
	//strat := glide_on_price.NewGlideOnPrice(handler,acc, &strategy.Settings{Symbol: symbol, BaseBid: 50}, observationsTimeFrame, monitorTimeFrame)
	//
	//
	//monitor := monitoring.NewBinanceMonitor(symbol, realApi,realSecretKey,monitorTimeFrame)
	//monitor.RunWithStrategy(strat)

	//signChan := make(chan os.Signal)
	//signal.Notify(signChan, os.Kill)
	//signal.Notify(signChan, os.Interrupt)
	//
	////Block next part of code via waiting response from channel
	//<-signChan

	//monitor.Stop()


	///////////////////////////////////////////// UP VALID

	//binance.UseTestnet = true
	//
	//acc, err := account.NewBinanceAccount(testnetApi,testnetSecret)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//orderId, err := acc.PlaceLimitOrder(10,400, "BNBUSDT", binance.SideTypeSell)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//fmt.Println(orderId)
	//
	//monitor := monitoring.NewBinanceMonitor("BTCUSDT",testnetApi,testnetSecret,time.Second*10)
	//
	//controllers := func(block *block.Block, order *binance.CreateOrderResponse, err error) {
	//	if err == nil {
	//		fmt.Println(order.Type)
	//		fmt.Println(block.ClosePrice)
	//	} else {
	//		fmt.Println(err)
	//	}
	//}
	//
	//strat := fallx3.NewFallx3Strategy(controllers,acc, &strategy.Settings{Symbol: "BTCUSDT", BaseBid: 9},time.Minute*35, 3, 10 *time.Second )
	//
	//monitor.RunWithStrategy(strat)
	//time.Sleep(time.Minute)
	//monitor.Stop()
}
