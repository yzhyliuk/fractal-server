package indicators

import "math"

// ADX returns the Average Directional Index for a given slice of high, low, and close prices and a specified length.
func ADX(high []float64, low []float64, close []float64, length int) float64 {
	// Calculate the Positive Directional Index (+DI) and Negative Directional Index (-DI)
	var plusDI float64
	var minusDI float64
	for i := 0; i < length; i++ {
		// Calculate the +DM and -DM values
		plusDM := high[i] - high[i-1]
		minusDM := low[i-1] - low[i]

		// Calculate the +DI and -DI values
		if plusDM > minusDM && plusDM > 0 {
			plusDI += plusDM
		}
		if minusDM > plusDM && minusDM > 0 {
			minusDI += minusDM
		}
	}
	plusDI /= float64(length)
	minusDI /= float64(length)

	// Calculate the Average True Range (ATR)
	var atr float64
	for i := 0; i < length; i++ {
		trueRange := math.Max(high[i]-low[i], math.Abs(high[i]-close[i-1]))
		trueRange = math.Max(trueRange, math.Abs(low[i]-close[i-1]))
		atr += trueRange
	}
	atr /= float64(length)

	// Calculate the ADX
	adx := 100 * math.Abs((plusDI - minusDI) / (plusDI + minusDI)) / atr
	return adx
}
