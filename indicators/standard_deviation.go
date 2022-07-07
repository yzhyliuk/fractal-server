package indicators

import "math"

func StandardDeviation(observations []float64) float64 {
	mean := Average(observations)
	sum := 0.

	for i := range observations {
		val := math.Pow(observations[i]-mean, 2)
		sum += val
	}

	return math.Sqrt(sum/float64(len(observations)))
}
