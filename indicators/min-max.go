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

func Max32(observations []float32) (float32, int) {
	max := float32(-999999999999999999999999999.)
	maxIndex := 0
	for i := range observations {
		if observations[i] > max {
			max = observations[i]
			maxIndex = i
		}
	}

	return max, maxIndex
}
