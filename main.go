package main

import (
	"flag"
	"log"
	"newTradingBot/api"
	"newTradingBot/api/common"
	"newTradingBot/configuration"
	"newTradingBot/logs"
	"newTradingBot/storage"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main()  {

	mode := flag.String("mode", "dev", "launch mode: dev (development) | prod (production)")
	flag.Parse()

	if mode != nil && *mode == configuration.Prod {
		configuration.Mode = configuration.Prod
	} else if *mode == configuration.Dev {
		configuration.Mode = configuration.Dev
	} else {
		configuration.Mode =configuration.DebugProd
	}

	c := make(chan os.Signal,1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGINT)
	signal.Notify(c, syscall.SIGKILL)
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
}

func terminateAllStrategies()  {
	logs.LogDebug("Terminating all strategies...", nil)
	for _, v := range storage.StrategiesStorage {
		v.Stop()
	}
	logs.LogDebug("All strategies are terminated.", nil)
	time.Sleep(3*time.Second)
	os.Exit(0)
}