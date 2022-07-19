package internal_arbitrage

import (
	"context"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"newTradingBot/logs"
	"newTradingBot/models/account"
	"strconv"
	"strings"
)

const ratioThreshold = 0.005

func RunArbitrageWithParams(targetCurrency, primarySymbol string, acc account.Account, buySum float64) error {
	secondarySymbols := make([]string, 0)
	allPairs, err := account.GetAllPairsForBinance(false)
	if err != nil {
		return err
	}

	for k,_ := range allPairs {
		if strings.Contains(k,primarySymbol) && !strings.Contains(k,targetCurrency){
			if primarySymbol == "XRP"{
				if k == "XRPBTC" || k == "XRPETH" {
					continue
				}
			}
			symbol := strings.Replace(k,primarySymbol,"",1)
			_, ok := allPairs[fmt.Sprintf("%s%s",symbol,targetCurrency)]
			if ok {
				secondarySymbols = append(secondarySymbols, k)
			}
		}
	}

	symbolOne := fmt.Sprintf("%s%s",primarySymbol,targetCurrency)

	for {

		bestRatio := make([]struct {
			Symbol string
			Ratio  float64
		}, 0)

		for i := range secondarySymbols {
			priceOne := getInstantPrice(symbolOne)
			priceTwo := getInstantPrice(secondarySymbols[i])
			symbolThree := fmt.Sprintf("%s%s", strings.Replace(secondarySymbols[i], primarySymbol, "", 1), targetCurrency)
			priceThree := getInstantPrice(symbolThree)

			ratio := (((1 / priceOne) / priceTwo) * priceThree) - (3*account.BinanceSpotTakerFee)
			if ratio >= 1+ratioThreshold {
				logs.LogDebug(fmt.Sprintf("%s %s %s RATIO: %f", secondarySymbols[i], primarySymbol, targetCurrency, ratio), nil)
				err = manageTrade(symbolOne, secondarySymbols[i], symbolThree, primarySymbol, priceOne, priceTwo, buySum, acc)
				if err != nil {
					logs.LogDebug("", err)
				}
				bestRatio = append(bestRatio, struct {
					Symbol string
					Ratio  float64
				}{Symbol: secondarySymbols[i], Ratio: ratio})
			}
		}
	}
	return nil
}

func getInstantPrice(symbol string) float64 {

	sum := 0.
	count := 10

	resp, err := binance.NewClient("","").NewRecentTradesService().Symbol(symbol).Limit(count).Do(context.Background())
	if err != nil {
		logs.LogDebug("", err)
	}
	for i := range resp {
		price, _ := strconv.ParseFloat(resp[i].Price, 64)
		sum += price
	}

	return sum/float64(count)
}

func manageTrade(symbolOne, symbolTwo, symbolThree, primary string, priceOne, priceTwo, sum float64, acc account.Account) error {
	quantity := sum/priceOne

	order, err := acc.PlaceRawSpotOrder(quantity, symbolOne,binance.SideTypeBuy)
	if err != nil {
		return err
	}
	logs.LogDebug(fmt.Sprintf("Step 1: %s - SUCCESS!", symbolOne),nil)

	quantityPrimary, _ := strconv.ParseFloat(order.ExecutedQuantity, 64)

	if strings.Index(symbolTwo,primary) != 0 {
		quantity = quantityPrimary / priceTwo
	} else  {
		quantity = quantityPrimary * priceTwo
	}

	order, err = acc.PlaceRawSpotOrder(quantity, symbolTwo,binance.SideTypeBuy)
	if err != nil {
		_, _ = acc.PlaceRawSpotOrder(quantityPrimary,symbolOne,binance.SideTypeSell)
		return err
	}
	logs.LogDebug(fmt.Sprintf("Step 2: %s - SUCCESS!", symbolTwo),nil)

	quantity, _ = strconv.ParseFloat(order.ExecutedQuantity, 64)

	order, err = acc.PlaceRawSpotOrder(quantity, symbolThree,binance.SideTypeSell)
	if err != nil {
		return err
	}

	logs.LogDebug(fmt.Sprintf("Step 3: %s - SUCCESS!", symbolThree),nil)

	return nil
}

