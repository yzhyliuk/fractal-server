package indicators

import (
	"math"
	"newTradingBot/models/block"
)

func LinearRegressionForTimeSeries(observations []float64) (slope, intercept float64) {
	yMean := Average(observations)
	xMean := MeanForSeries(len(observations))

	upperSum := 0.
	lowerSum := 0.

	for x, y := range observations {
		//upperSum += (x - xMean)*(y - yMean)
		//lowerSum += math.Pow(x - xMean, 2)
		upperSum += (float64(x) - xMean) * (y - yMean)
		lowerSum += math.Pow(float64(x)-xMean, 2)
	}

	slope = upperSum / lowerSum

	intercept = yMean - slope*xMean

	return
}

func LinearRegressionForDataFrameClosePrice(df []*block.Data) (slope, intercept float64) {
	priceObservations := make([]float64, len(df))
	for idx := range df {
		priceObservations[idx] = df[idx].ClosePrice
	}

	return LinearRegressionForTimeSeries(priceObservations)
}

func MeanForSeries(num int) float64 {
	sum := 0.

	for i := 1; i <= num; i++ {
		sum += float64(i)
	}

	return sum / float64(num)
}
