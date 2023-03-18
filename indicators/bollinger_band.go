package indicators

// BollingerBands returns the upper and lower Bollinger Bands for a given slice of data and a specified length and standard deviation multiplier.
func BollingerBands(data []float64, length int, multiplier float64) (float64, float64) {
	sma := SimpleMA(data, length)
	stdDev := StandardDeviation(data[len(data)-length:])

	// Calculate the upper and lower Bollinger Bands
	upperBand := sma + (multiplier * stdDev)
	lowerBand := sma - (multiplier * stdDev)

	return upperBand, lowerBand
}
