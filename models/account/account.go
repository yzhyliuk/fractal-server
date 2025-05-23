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

const BinanceFuturesTakerFeeRate = 0.0004
const BinanceSpotTakerFee = 0.00075

type Account interface {
	GetStableBalance() (*UserBalances, error)
	PlaceMarketOrder(sum float64, symbol string, side binance.SideType, inst *instance.StrategyInstance, prevTrade *trade.Trade) (*trade.Trade, error)
	OpenFuturesPosition(amount float64, symbol string, side futures.SideType, inst *instance.StrategyInstance) (*trade.Trade, error)
	CloseFuturesPosition(tradeFutures *trade.Trade) (*trade.Trade, error)
	PlaceRawSpotOrder(sum float64, symbol string, side binance.SideType) (*binance.CreateOrderResponse, error)
	GetOrder(orderID int64) (*futures.Order, error)
	GetOpenedFuturesTrades(symbol string, timeStamp time.Time) ([]*futures.PositionRisk, error)
	GetCurrentQuote(symbol string) (float64, error)
}

type TakeProfit struct {
	Sum   float64
	Price float64
}

type AssetPrecision struct {
	Price    int
	Quantity int
}

type BinanceAccount struct {
	apiKey    string
	secretKey string

	userID int

	client        *binance.Client
	futuresClient *futures.Client

	precision        map[string]AssetPrecision
	precisionFutures map[string]AssetPrecision
}

// NewBinanceAccount returns new entity of binance account for given keys
func NewBinanceAccount(apiKey, secretKey, futuresApiKey, futuresSecretKey string) (Account, error) {
	nba := &BinanceAccount{
		apiKey:    apiKey,
		secretKey: secretKey,
	}

	var err error

	// get new client cor account
	nba.client = binance.NewClient(apiKey, secretKey)
	nba.futuresClient = futures.NewClient(futuresApiKey, futuresSecretKey)

	// load precision for all trading symbols
	nba.precision, err = nba.getPrecisionMap()
	if err != nil {
		return nil, err
	}
	nba.precisionFutures, err = nba.getFuturesPrecisionMap()
	if err != nil {
		return nil, err
	}

	return nba, nil
}

func (b *BinanceAccount) GetCurrentQuote(symbol string) (float64, error) {
	v, err := b.futuresClient.NewListPricesService().Symbol(symbol).Do(context.Background())
	if err != nil {
		return 0, err
	}

	price, _ := strconv.ParseFloat(v[0].Price, 64)

	return price, nil
}

func (b *BinanceAccount) GetOrder(orderID int64) (*futures.Order, error) {
	return b.futuresClient.NewGetOrderService().OrderID(orderID).Do(context.Background())
}

func (b *BinanceAccount) PlaceRawSpotOrder(sum float64, symbol string, side binance.SideType) (*binance.CreateOrderResponse, error) {
	quantity := b.formatQuantity(sum, symbol, false)
	order, err := b.client.NewCreateOrderService().Symbol(symbol).Side(side).Type(binance.OrderTypeMarket).Quantity(quantity).Do(context.Background())
	if err != nil {
		return nil, err
	}

	return order, nil
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

	price = price / float64(count)

	quantityFloat, _ := strconv.ParseFloat(quantity, 64)

	db, err := database.GetDataBaseConnection()
	if err != nil {
		return nil, err
	}

	if prevTrade == nil {
		spotTrade = &trade.Trade{
			InstanceID: inst.ID,
			UserID:     inst.UserID,
			StrategyID: inst.StrategyID,
			Pair:       symbol,
			USD:        quantityFloat * price,
			PriceOpen:  price,
			Quantity:   quantityFloat,
			TimeStamp:  time.Now(),
			Status:     trade.StatusActive,
		}
		spotTrade, err = trade.NewTrade(db, *spotTrade)
		if err != nil {
			return nil, err
		}
	} else {
		spotTrade = prevTrade
		spotTrade.PriceClose = price
		spotTrade.Status = trade.StatusClosed
		spotTrade.Profit = (spotTrade.Quantity * price) - spotTrade.USD
		spotTrade.ROI = spotTrade.Profit / spotTrade.USD

		err = trade.CloseTrade(db, spotTrade)
		if err != nil {
			return nil, err
		}
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

func (b *BinanceAccount) OpenFuturesPosition(amount float64, symbol string, side futures.SideType, inst *instance.StrategyInstance) (*trade.Trade, error) {

	quantity := b.formatQuantity(amount, symbol, true)

	res, err := b.futuresClient.NewCreateOrderService().Quantity(quantity).Symbol(symbol).Side(side).Type(futures.OrderTypeMarket).Do(context.Background())
	if err != nil {
		return nil, err
	}

	orderID := res.OrderID

	var order *futures.Order

	for order == nil {
		order, _ = b.futuresClient.NewGetOrderService().Symbol(symbol).OrderID(orderID).Do(context.Background())
		if err != nil {
			time.Sleep(10 * time.Second)
		}
	}

	futuresTrade := &trade.Trade{}

	price, err := strconv.ParseFloat(order.AvgPrice, 64)
	if err != nil {
		return nil, err
	}

	futuresTrade = &trade.Trade{
		InstanceID:     inst.ID,
		UserID:         inst.UserID,
		StrategyID:     inst.StrategyID,
		Pair:           symbol,
		USD:            amount * price,
		PriceOpen:      price,
		Quantity:       amount,
		TimeStamp:      time.Now(),
		Status:         trade.StatusActive,
		FuturesSide:    side,
		BinanceOrderID: orderID,
		Leverage:       inst.Leverage,
	}

	db, err := database.GetDataBaseConnection()
	if err != nil {
		return nil, err
	}

	futuresTrade, err = trade.NewTrade(db, *futuresTrade)
	if err != nil {
		return nil, err
	}

	return futuresTrade, nil
}

func (b *BinanceAccount) OpenNewTakeProfitOrder(takeProfit TakeProfit, symbol string, side futures.SideType) (*futures.CreateOrderResponse, error) {
	price := b.formatPrice(takeProfit.Price, symbol)
	quantity := b.formatQuantity(takeProfit.Sum, symbol, true)
	return b.futuresClient.NewCreateOrderService().ReduceOnly(true).TimeInForce(futures.TimeInForceTypeGTC).StopPrice(price).Quantity(quantity).Symbol(symbol).Side(side).Type(futures.OrderTypeTakeProfitMarket).Do(context.Background())
}

func (b *BinanceAccount) CloseFuturesPosition(futuresTrade *trade.Trade) (*trade.Trade, error) {
	quantity := b.formatQuantity(futuresTrade.Quantity, futuresTrade.Pair, true)
	resolvedSide := futures.SideTypeSell
	if futuresTrade.FuturesSide == futures.SideTypeSell {
		resolvedSide = futures.SideTypeBuy
	}

	res, err := b.futuresClient.NewCreateOrderService().Quantity(quantity).Symbol(futuresTrade.Pair).Side(resolvedSide).Type(futures.OrderTypeMarket).Do(context.Background())
	if err != nil {
		return nil, err
	}

	var order *futures.Order

	for order == nil {
		order, err = b.futuresClient.NewGetOrderService().Symbol(futuresTrade.Pair).OrderID(res.OrderID).Do(context.Background())
		if err != nil {
			time.Sleep(5 * time.Second)
		} else {
			break
		}
	}

	price, err := strconv.ParseFloat(order.AvgPrice, 64)
	if err != nil {
		return nil, err
	}

	roi := 0.
	profit := 0.
	fee := futuresTrade.USD * BinanceFuturesTakerFeeRate

	switch futuresTrade.FuturesSide {
	case futures.SideTypeBuy:
		profit = (futuresTrade.Quantity * price) - futuresTrade.USD
	case futures.SideTypeSell:
		profit = (futuresTrade.Quantity * futuresTrade.PriceOpen) - (futuresTrade.Quantity * price)
	}

	profit -= 2 * fee

	roi = profit / (futuresTrade.USD / float64(*futuresTrade.Leverage))

	futuresTrade.PriceClose = price
	futuresTrade.Status = trade.StatusClosed
	futuresTrade.Profit = profit
	futuresTrade.ROI = roi

	db, err := database.GetDataBaseConnection()

	err = trade.CloseTrade(db, futuresTrade)

	return futuresTrade, err
}

func (b *BinanceAccount) GetOpenedFuturesTrades(symbol string, timeStamp time.Time) ([]*futures.PositionRisk, error) {
	res, err := b.futuresClient.NewGetPositionRiskService().Symbol(symbol).Do(context.Background())
	if err != nil {
		return nil, err
	}

	return res, err
}

func (b *BinanceAccount) CloseOpenedFuturesTrade(orderID int64, symbol string) (*futures.CancelOrderResponse, error) {
	return b.futuresClient.NewCancelOrderService().OrderID(orderID).Symbol(symbol).Do(context.Background())
}

func (b *BinanceAccount) formatQuantity(sum float64, symbol string, isFutures bool) string {
	if isFutures {
		return fmt.Sprintf("%.*f", b.precisionFutures[symbol].Quantity, sum)
	} else {
		q := b.precision[symbol].Quantity
		return fmt.Sprintf("%.*f", q, sum)
	}
}

func (b *BinanceAccount) formatPrice(price float64, symbol string) string {
	return fmt.Sprintf("%.*f", b.precisionFutures[symbol].Price, price)
}

func QuantityFromPrice(bidSize, price float64) float64 {
	return bidSize / price
}
