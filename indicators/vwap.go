package indicators

// VWAP returns the Volume Weighted Average Price for a given slice of data.
func VWAP(data []float64, volumes []int) float64 {
	// Calculate the total volume and the volume-weighted sum of the data
	var totalVolume int
	var volumeWeightedSum float64
	for i, value := range data {
		totalVolume += volumes[i]
		volumeWeightedSum += value * float64(volumes[i])
	}

	// Calculate the VWAP
	vwap := volumeWeightedSum / float64(totalVolume)
	return vwap
}
