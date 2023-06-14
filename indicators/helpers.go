package indicators

func GetSlicedArray(array []float64, length int) []float64 {
	return array[len(array)-1-length:]
}
