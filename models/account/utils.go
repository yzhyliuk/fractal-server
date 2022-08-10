package account

import (
	"context"
	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"strings"
)

const filterType = "filterType"
const lotSizeFilter = "LOT_SIZE"
const stepSize = "stepSize"

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

func (b *BinanceAccount) getPrecisionMap() (map[string]AssetPrecision, error)  {

	//Getting limits for min quantity of current trading pair
	exchangeInfo, err := b.client.NewExchangeInfoService().Do(context.Background())
	if err != nil {
		return nil, err
	}

	precisionMap := make(map[string]AssetPrecision)

	for _, elem := range exchangeInfo.Symbols {
		filters := elem.Filters
		var stepSizeFilter interface{}

		for _, val := range filters {
			if val["filterType"] == lotSizeFilter {
				stepSizeFilter = val["stepSize"]
			}
		}

		str := stepSizeFilter.(string)
		str = strings.Replace(str,".","",1)
		prec := strings.Index(str,"1")
		precisionMap[elem.Symbol] = AssetPrecision{
			Quantity: prec,
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

