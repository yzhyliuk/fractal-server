package indicators

// MACD returns the Moving Average Convergence Divergence for a given slice of data and two specified moving average lengths.
func MACD(data []float64, fastLength int, slowLength int) float64 {
	// Calculate the fast and slow moving averages
	var fastMA float64
	var slowMA float64
	for i, value := range data {
		if i < fastLength {
			fastMA += value
		}
		if i < slowLength {
			slowMA += value
		}
	}
	fastMA /= float64(fastLength)
	slowMA /= float64(slowLength)

	// Calculate the MACD
	macd := fastMA - slowMA
	return macd
}
