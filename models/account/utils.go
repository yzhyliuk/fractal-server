package account

import (
	"context"
	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"strings"
)

// GetAllPairsForBinance returns map with symbols for all available currencies pairs
func GetAllPairsForBinance(isFutures bool) (map[string]struct{}, error) {

	pairsMap := make(map[string]struct{})

	if !isFutures {
		client := binance.NewClient("", "")

		ExchangeInfo, err := client.NewExchangeInfoService().Do(context.Background())
		if err != nil {
			return nil, err
		}

		for _, pair := range ExchangeInfo.Symbols {
			if pair.Status == statusTrading {
				pairsMap[pair.Symbol] = struct{}{}
			}
		}

		return pairsMap, nil
	} else  {
		client := futures.NewClient("", "")

		ExchangeInfo, err := client.NewExchangeInfoService().Do(context.Background())
		if err != nil {
			return nil, err
		}

		for _, pair := range ExchangeInfo.Symbols {
			if pair.Status == statusTrading {
				pairsMap[pair.Symbol] = struct{}{}
			}
		}

		return pairsMap, nil
	}
}

func (b *BinanceAccount) getPrecisionMap() (map[string]int, error)  {

	//Getting limits for min quantity of current trading pair
	exchangeInfo, err := b.client.NewExchangeInfoService().Do(context.Background())
	if err != nil {
		return nil, err
	}

	allPairs, err := GetAllPairsForBinance(false)
	if err != nil {
		return nil, err
	}

	precisionMap := make(map[string]int)

	for _, elem := range exchangeInfo.Symbols {
		_, exists := allPairs[elem.Symbol]
		if exists {

			precision := 0
			spl := strings.Split(elem.LotSizeFilter().MinQuantity, ".")

			for _, v := range spl[1] {
				if v != '1' {
					precision++
				} else {
					precision++
					break
				}
			}
			precisionMap[elem.Symbol] = precision
		}
	}
	return precisionMap, nil
}

func (b *BinanceAccount) getFuturesPrecisionMap() (map[string]AssetPrecision, error)  {

	//Getting limits for min quantity of current trading pair
	exchangeInfo, err := b.futuresClient.NewExchangeInfoService().Do(context.Background())
	if err != nil {
		return nil, err
	}

	precisionMap := make(map[string]AssetPrecision)

	for _, elem := range exchangeInfo.Symbols {
		precisionMap[elem.Symbol] = AssetPrecision{
			Price: elem.PricePrecision,
			Quantity: elem.QuantityPrecision,
		}
	}
	return precisionMap, nil
}

