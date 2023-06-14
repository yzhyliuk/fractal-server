package internal_arbitrage

import (
	"context"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"newTradingBot/logs"
	"newTradingBot/models/account"
	"strconv"
	"strings"
	"sync"
	"time"
)

const ratioThreshold = 0.015
const maxTrade = 3

var mutex sync.Mutex
var lastPrice = make(map[string]float64)

func RunArbitrageWithParams(targetCurrency, primarySymbol string, acc account.Account, buySum float64) error {
	secondarySymbols := make([]string, 0)
	allPairs, err := account.GetAllPairsForBinance(false)
	if err != nil {
		return err
	}
	for k, _ := range allPairs {
		if strings.Contains(k, primarySymbol) && !strings.Contains(k, targetCurrency) {
			if primarySymbol == "XRP" {
				if k == "XRPBTC" || k == "XRPETH" {
					continue
				}
			}
			symbol := strings.Replace(k, primarySymbol, "", 1)
			_, ok := allPairs[fmt.Sprintf("%s%s", symbol, targetCurrency)]
			if ok {
				secondarySymbols = append(secondarySymbols, k)
			}
		}
	}

	symbolOne := fmt.Sprintf("%s%s", primarySymbol, targetCurrency)
	//timer := time.Now()

	lastPrice[symbolOne] = getInstantPrice(symbolOne)
	logs.LogDebug("Start Initial get price", err)
	for i := range secondarySymbols {
		symbolThree := fmt.Sprintf("%s%s", strings.Replace(secondarySymbols[i], primarySymbol, "", 1), targetCurrency)
		lastPrice[symbolThree] = getInstantPrice(symbolThree)
		lastPrice[secondarySymbols[i]] = getInstantPrice(secondarySymbols[i])
	}

	RunMonitor(symbolOne)

	for i := range secondarySymbols {
		symbolThree := fmt.Sprintf("%s%s", strings.Replace(secondarySymbols[i], primarySymbol, "", 1), targetCurrency)
		RunMonitor(secondarySymbols[i])
		RunMonitor(symbolThree)
	}

	time.Sleep(10 * time.Second)

	logs.LogDebug("Start Monitoring", err)

	tC := 0
	for {
		if tC > maxTrade {
			break
		}
		for i := range secondarySymbols {
			symbolThree := fmt.Sprintf("%s%s", strings.Replace(secondarySymbols[i], primarySymbol, "", 1), targetCurrency)
			priceOne, priceTwo, priceThree, ratio, signal := GetStaticRatio(symbolOne, secondarySymbols[i], symbolThree, primarySymbol, targetCurrency)
			if signal {
				timestamp := time.Now()
				err = manageTrade(symbolOne, secondarySymbols[i], symbolThree, primarySymbol, priceOne, priceTwo, buySum, acc)
				if err != nil {
					logs.LogDebug("", err)
				}
				tC++
				logs.LogDebug(fmt.Sprintf("%s %s %s RATIO: %f", secondarySymbols[i], primarySymbol, targetCurrency, ratio), nil)
				logs.LogDebug(fmt.Sprintf("Trade execution time: %s", time.Since(timestamp)), nil)
				logs.LogDebug(fmt.Sprintf("P1: %f P2: %f P3: %f", priceOne, priceTwo, priceThree), nil)
				time.Sleep(5 * time.Second)
			}
		}
	}

	return nil
}

func RunMonitor(symbol string) {
	go func() {
		for {
			pricesum := 0.
			counter := 0
			wsAggTradeHandler := func(event *binance.WsAggTradeEvent) {
				price, _ := strconv.ParseFloat(event.Price, 64)
				pricesum += price
				counter++
			}
			_, stopC, _ := binance.WsAggTradeServe(symbol, wsAggTradeHandler, func(err error) {
				logs.LogError(err)
			})

			time.Sleep(100 * time.Millisecond)
			stopC <- struct{}{}

			if pricesum == 0 {
				continue
			}

			mutex.Lock()
			lastPrice[symbol] = pricesum / float64(counter)
			mutex.Unlock()
		}
	}()
}

func GetStaticRatio(s1, s2, s3, primary, target string) (ratio, priceOne, priceTwo, priceThree float64, up bool) {
	mutex.Lock()
	priceTwo = lastPrice[s2]
	priceThree = lastPrice[s3]
	priceOne = lastPrice[s1]
	mutex.Unlock()

	stepOne := 1 / priceOne
	stepTwo := 0.

	if strings.Index(s2, primary) != 0 {
		//buy
		stepTwo = stepOne / priceTwo
	} else {
		//sell
		stepTwo = stepOne * priceTwo
	}

	ratio = stepTwo*priceThree - (2 * account.BinanceSpotTakerFee)

	return priceOne, priceTwo, priceThree, ratio, ratio >= 1+ratioThreshold
}

func GetRatio(s1, s2, s3 string) (ratio, priceOne, priceTwo, priceThree float64, up bool) {
	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		priceTwo = getInstantPrice(s2)
		defer wg.Done()
	}()

	go func() {
		priceThree = getInstantPrice(s3)
		defer wg.Done()
	}()

	go func() {
		priceOne = getInstantPrice(s1)
		defer wg.Done()
	}()

	wg.Wait()

	ratio = (((1 / priceOne) / priceTwo) * priceThree) - (2 * account.BinanceSpotTakerFee)

	return priceOne, priceTwo, priceThree, ratio, ratio >= 1+ratioThreshold
}

func getInstantPrice(symbol string) float64 {
	count := 10
	resp, err := binance.NewClient("", "").NewRecentTradesService().Symbol(symbol).Limit(count).Do(context.Background())
	if err != nil {
		logs.LogDebug("", err)
	}

	min := 999999999999999999999999999.
	for i := range resp {
		price, _ := strconv.ParseFloat(resp[i].Price, 64)
		if price < min {
			min = price
		}
	}

	return min
}

func manageTrade(symbolOne, symbolTwo, symbolThree, primary string, priceOne, priceTwo, sum float64, acc account.Account) error {
	quantity := sum / priceOne

	order, err := acc.PlaceRawSpotOrder(quantity, symbolOne, binance.SideTypeBuy)
	if err != nil {
		return err
	}

	quantityPrimary, _ := strconv.ParseFloat(order.ExecutedQuantity, 64)

	sideType := binance.SideTypeBuy
	if strings.Index(symbolTwo, primary) != 0 {
		quantity = quantityPrimary / priceTwo
	} else {
		quantity = quantityPrimary * priceTwo
		sideType = binance.SideTypeSell
	}

	order, err = acc.PlaceRawSpotOrder(quantity, symbolTwo, sideType)
	if err != nil {
		_, _ = acc.PlaceRawSpotOrder(quantityPrimary, symbolOne, binance.SideTypeSell)
		return err
	}
	quantity, _ = strconv.ParseFloat(order.ExecutedQuantity, 64)

	order, err = acc.PlaceRawSpotOrder(quantity, symbolThree, binance.SideTypeSell)
	if err != nil {
		return err
	}

	logs.LogDebug(fmt.Sprintf("Step 1: %s - SUCCESS! ", symbolOne), nil)
	logs.LogDebug(fmt.Sprintf("Step 2: %s - SUCCESS! ", symbolTwo), nil)
	logs.LogDebug(fmt.Sprintf("Step 3: %s - SUCCESS", symbolThree), nil)

	return nil
}
