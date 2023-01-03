package indicators

// AverageTrueRange Calculates the Average True Range (ATR) for a given set of high, low, and close prices, using a specified period.
func AverageTrueRange(highPrices, lowPrices, closePrices []float64, atrPeriod int) []float64 {
	atrValues := make([]float64, len(highPrices))

	// Loop through the input prices and calculate the ATR values
	for i := range highPrices {
		trueRange := 0.0
		if i == 0 {
			// For the first ATR value, use the range of the first period
			trueRange = highPrices[i] - lowPrices[i]
		} else {
			// For all other ATR values, use the maximum of the true range, the distance from the current high to the previous close, and the distance from the current low to the previous close
			trueRange = MaxOf3float(highPrices[i]-lowPrices[i], highPrices[i]-closePrices[i-1], closePrices[i-1]-lowPrices[i])
		}

		if i < atrPeriod-1 {
			// If we don't have enough data yet to calculate the ATR for the current period, just set the ATR value to 0
			atrValues[i] = 0
		} else if i == atrPeriod-1 {
			// If we have enough data to calculate the first ATR value, use the simple average of the true range over the previous period
			atrSum := 0.0
			for j := 0; j < atrPeriod; j++ {
				atrSum += trueRange
			}
			atrValues[i] = atrSum / float64(atrPeriod)
		} else {
			// For all other ATR values, use the exponential moving average to smooth the values
			atrValues[i] = (atrValues[i-1]*float64(atrPeriod-1) + trueRange) / float64(atrPeriod)
		}
	}

	return atrValues
}

// MaxOf3float helper function to return the maximum of 3 float64 values
func MaxOf3float(a, b, c float64) float64 {
	if a > b && a > c {
		return a
	} else if b > c {
		return b
	} else {
		return c
	}
}

// MaxOf3int helper function to return the maximum of 3 float64 values
func MaxOf3int(a, b, c int) int {
	if a > b && a > c {
		return a
	} else if b > c {
		return b
	} else {
		return c
	}
}
