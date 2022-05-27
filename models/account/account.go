package account

import (
	"context"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"newTradingBot/api/database"
	"newTradingBot/models/strategy/instance"
	"newTradingBot/models/trade"
	"strconv"
	"time"
)

type Account interface {
	GetBalance(symbol string) float64
	PlaceMarketOrder(sum float64, symbol string, side binance.SideType, inst *instance.StrategyInstance, prevTrade *trade.Trade) (*trade.Trade, error)
	// PlaceLimitOrder(sum, price float64,symbol string, side binance.SideType) (*trade.Trade, error)
	OpenFuturesPosition(stopLoss, amount float64, takeProfits []TakeProfit, symbol string, side futures.SideType) (*futures.CreateOrderResponse, error)
	CloseFuturesPosition()
}

type TakeProfit struct {
	Sum   float64
	Price float64
}

type AssetPrecision struct {
	Price int
	Quantity int
}

type BinanceAccount struct {
	apiKey string
	secretKey string

	userID int

	client *binance.Client
	futuresClient *futures.Client

	precision map[string]int
	precisionFutures map[string]AssetPrecision
}

// NewBinanceAccount returns new entity of binance account for given keys
func NewBinanceAccount(apiKey, secretKey, futuresApiKey, futuresSecretKey string) (Account, error) {
	nba := &BinanceAccount{
		apiKey: apiKey,
		secretKey: secretKey,
	}

	var err error

	// get new client cor account
	nba.client = binance.NewClient(apiKey,secretKey)
	nba.futuresClient = futures.NewClient(futuresApiKey,futuresSecretKey)

	// load precision for all trading symbols
	nba.precision, err = nba.getPrecisionMap()
	if err !=nil {
		return nil, err
	}
	nba.precisionFutures, err = nba.getFuturesPrecisionMap()
	if err !=nil {
		return nil, err
	}

	return nba, nil
}

func (b *BinanceAccount) GetBalance(symbol string) float64 {
	return 0
}
func (b *BinanceAccount) PlaceMarketOrder(sum float64, symbol string, side binance.SideType, inst *instance.StrategyInstance, prevTrade *trade.Trade) (*trade.Trade, error) {
	quantity := b.formatQuantity(sum, symbol, false)
	order, err := b.client.NewCreateOrderService().Symbol(symbol).Side(side).Type(binance.OrderTypeMarket).Quantity(quantity).Do(context.Background())
	if err != nil {
		return nil, err
	}

	spotTrade := &trade.Trade{}

	price := 0.
	count := 0
	//price, _ := strconv.ParseFloat(order.Price, 64)
	for _, fill := range order.Fills {
		partPrice, _ := strconv.ParseFloat(fill.Price, 64)
		price += partPrice
		count++
	}

	price = price/float64(count)

	quantityFloat, _ := strconv.ParseFloat(quantity, 64)

	db, err := database.GetDataBaseConnection()
	if err != nil {
		return nil, err
	}

	if prevTrade == nil {
		spotTrade = &trade.Trade{
			InstanceID: inst.ID,
			UserID: inst.UserID,
			StrategyID: inst.StrategyID,
			Pair: symbol,
			USD: quantityFloat*price,
			IsFutures: false,
			PriceBuy: price,
			Quantity: quantityFloat,
			TimeStamp: time.Now(),
			Status: trade.StatusActive,
		}
		spotTrade, err = trade.NewTrade(db, *spotTrade)
		if err != nil {
			return nil, err
		}
	} else {
		spotTrade = prevTrade
		spotTrade.PriceSell = price
		spotTrade.Status = trade.StatusClosed
		spotTrade.Profit = (spotTrade.Quantity*price)-spotTrade.USD
		spotTrade.ROI = (spotTrade.Quantity*price)/spotTrade.USD-1

		err = trade.CloseTrade(db, spotTrade)
	}

	return spotTrade, nil
}

//func (b *BinanceAccount) PlaceLimitOrder(sum, price float64,symbol string,side binance.SideType) (*trade.Trade, error) {
//	quantity := b.formatQuantity(sum, symbol, false)
//	limit := fmt.Sprintf("%f", price)
//	order, err := b.client.NewCreateOrderService().Symbol(symbol).
//		Side(side).Type(binance.OrderTypeLimit).
//		TimeInForce(binance.TimeInForceTypeGTC).Quantity(quantity).
//		Price(limit).Do(context.Background())
//	if err != nil {
//		return nil, err
//	}
//	return
//}

func (b *BinanceAccount) OpenFuturesPosition(stopLoss, amount float64, takeProfits []TakeProfit, symbol string, side futures.SideType) (*futures.CreateOrderResponse, error) {
	quantity := b.formatQuantity(amount, symbol, true)
	res, err := b.futuresClient.NewCreateOrderService().Quantity(quantity).Symbol(symbol).Side(side).Type(futures.OrderTypeMarket).Do(context.Background())
	if err != nil {
		return nil, err
	}

	if side == futures.SideTypeBuy {
		side = futures.SideTypeSell
	} else {
		side = futures.SideTypeBuy
	}

	stopPrice := b.formatPrice(stopLoss,symbol)
	quantity = b.formatQuantity(amount*10,symbol,true)
	_, err = b.futuresClient.NewCreateOrderService().ReduceOnly(true).TimeInForce(futures.TimeInForceTypeGTC).StopPrice(stopPrice).Quantity(quantity).Symbol(symbol).Side(side).Type(futures.OrderTypeStopMarket).Do(context.Background())
	if err != nil {
		return  nil, err
	}

	for i := range takeProfits {
		_, err := b.OpenNewTakeProfitOrder(takeProfits[i], symbol, side)
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

func (b *BinanceAccount) OpenNewTakeProfitOrder(takeProfit TakeProfit, symbol string, side futures.SideType) (*futures.CreateOrderResponse, error) {
	price := b.formatPrice(takeProfit.Price ,symbol)
	quantity := b.formatQuantity(takeProfit.Sum, symbol, true)
	return b.futuresClient.NewCreateOrderService().ReduceOnly(true).TimeInForce(futures.TimeInForceTypeGTC).StopPrice(price).Quantity(quantity).Symbol(symbol).Side(side).Type(futures.OrderTypeTakeProfitMarket).Do(context.Background())
}

func (b *BinanceAccount) CloseFuturesPosition() {

}

func (b *BinanceAccount) formatQuantity(sum float64, symbol string, isFutures bool) string {
	if isFutures {
		return fmt.Sprintf("%.*f", b.precisionFutures[symbol].Quantity, sum)
	} else {
		return fmt.Sprintf("%.*f", b.precision[symbol], sum)
	}
}

func (b *BinanceAccount) formatPrice(price float64, symbol string) string {
	return fmt.Sprintf("%.*f", b.precisionFutures[symbol].Price, price)
}

func QuantityFromPrice(bidSize, price float64 ) float64 {
	return bidSize/price
}

