package indicators

// RMA calculates the Relative Moving Average
func RMA(current, previous float64, length int) float64 {
	return (current + (float64(length-1) * previous)) / float64(length)
}
