package indicators

func Min(observations []float64) float64 {
	min := 999999999999999999999999999.
	for i := range observations {
		if observations[i] < min {
			min = observations[i]
		}
	}

	return min
}

func Max(observations []float64) float64 {
	max := -999999999999999999999999999.
	for i := range observations {
		if observations[i] > max {
			max = observations[i]
		}
	}

	return max
}