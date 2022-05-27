package indicators

func RSI(observations []float64, period int) float64 {
	if period > len(observations) {
		return 0
	}

	gain, loss := averageGainAndLoses(observations, period)

	if gain == 0 {
		return 0
	} else if loss == 0 {
		return 100
	}

	rs := gain/loss

	return 100-(100/(1+rs))
}

func averageGainAndLoses(observations []float64, period int) (gain, loss float64)  {
	gainList := make([]float64, 0)
	lossList := make([]float64, 0)

	count := len(observations)-1

	for i := count; i > count-(period-1); i-- {
		previousClosePrice := observations[i-1]
		delta := observations[i]-previousClosePrice
		if delta > 0 {
			gainList = append(gainList, delta)
			lossList = append(lossList, 0)
		} else if delta < 0 {
			lossList = append(lossList, -delta)
			gainList = append(gainList, 0)
		}
	}

	return Average(gainList), Average(lossList)
}

