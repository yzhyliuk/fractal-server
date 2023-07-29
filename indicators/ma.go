package indicators

const UpTrend = 1
const DownTrend = 2
const UnKnownTrend = 3

func GetMovingAverageTrend(observations []float64, candles int) int {
	ma := SimpleMA(observations, len(observations))
	high := 0
	low := 0
	for _, price := range observations[len(observations)-candles:] {
		if price > ma {
			high++
		} else {
			low++
		}
	}

	if high == candles {
		return UpTrend
	} else if low == candles {
		return DownTrend
	}

	return UnKnownTrend
}

func Average(array []float64) float64 {

	if len(array) == 0 {
		return 0
	}

	sum := 0.
	for _, val := range array {
		sum += val
	}

	return sum / float64(len(array))
}

func Sum(array []float64) float64 {
	sum := 0.

	for _, v := range array {
		sum += v
	}

	return sum
}

func SimpleMA(observations []float64, period int) float64 {
	if period > len(observations) {
		return 0
	}

	length := len(observations)
	sum := 0.

	for i := length - 1; i > length-1-period; i-- {
		sum += observations[i]
	}

	return sum / float64(period)
}

func ExponentialMA(period int, previousEMA, currentValue float64) float64 {
	multiplier := float64(2) / float64(period+1)

	ema := (currentValue * multiplier) + (previousEMA * (1 - multiplier))
	return ema
}
