package indicators

func Average(array []float64) float64  {

	if len(array) == 0 {
		return 0
	}

	sum := 0.
	for _, val := range array {
		sum += val
	}

	return sum/float64(len(array))
}

func SimpleMA(observations []float64, period int) float64  {
	if period > len(observations) {
		return 0
	}

	length := len(observations)
	sum := 0.

	for i := length-1; i > length-1-period; i-- {
		sum += observations[i]
	}

	return sum/float64(period)
}

func ExponentialMA(period int, previousEMA, currentValue float64) float64  {
	multiplier := float64(2) / float64(period+1)

	ema := currentValue*multiplier + previousEMA*(1-multiplier)
	return ema
}

