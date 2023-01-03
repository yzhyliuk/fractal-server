package indicators

import "math"

// BollingerBands returns the upper and lower Bollinger Bands for a given slice of data and a specified length and standard deviation multiplier.
func BollingerBands(data []float64, length int, multiplier float64) (float64, float64) {
	// Calculate the simple moving average of the data
	var sum float64
	for _, value := range data {
		sum += value
	}
	sma := sum / float64(length)

	// Calculate the standard deviation of the data
	var variance float64
	for _, value := range data {
		variance += math.Pow(value-sma, 2)
	}
	stdDev := math.Sqrt(variance / float64(length))

	// Calculate the upper and lower Bollinger Bands
	upperBand := sma + (multiplier * stdDev)
	lowerBand := sma - (multiplier * stdDev)

	return upperBand, lowerBand
}
