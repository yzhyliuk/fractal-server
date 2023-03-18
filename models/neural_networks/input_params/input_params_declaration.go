package input_params

import "newTradingBot/models/apimodels"

const RSI = "rsi"
const SMA = "sma"
const Volume = "volume"
const PriceChange = "priceChange"

var params = []apimodels.UiParam{
	{"RSI", RSI}, {"SMA", SMA}, {"Volume", Volume}, {"Price change", PriceChange},
}

func GetInputParams() []apimodels.UiParam {
	return params
}
