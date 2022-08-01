package account

import (
	"context"
	"strconv"
)

const usdt = "USDT"
const busd = "BUSD"
const usdc = "USDC"

type UserBalances struct {
	BUSD float64 `json:"busd"`
	USDT float64 `json:"usdt"`
	USDC float64 `json:"usdc"`
}

func (b *BinanceAccount) GetStableBalance() (*UserBalances, error) {
	balance, err := b.futuresClient.NewGetBalanceService().Do(context.Background())
	if err != nil {
		return nil, err
	}

	uBalance := &UserBalances{
		BUSD: 0,
		USDT: 0,
		USDC: 0,
	}

	for i := range balance{
		sum, _ := strconv.ParseFloat(balance[i].Balance,64)
		switch balance[i].Asset {
			case busd:
				uBalance.BUSD = sum
			case usdc:
				uBalance.USDC = sum
			case usdt:
				uBalance.USDT = sum
		}
	}

	return uBalance, err
}