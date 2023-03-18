package neural_networks

import (
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"math"
	"newTradingBot/indicators"
	"newTradingBot/logs"
	"newTradingBot/models/block"
	"newTradingBot/models/neural_networks/activation_funcs"
	"newTradingBot/models/neural_networks/linear"
	"newTradingBot/models/neural_networks/mlp"
)

func TestMlp(db *gorm.DB) {
	captureID := 20

	data, _ := block.GetCaptureDataBySessionId(captureID, db)

	features := make(linear.Frame, 0)
	labels := make(linear.Frame, 0)

	// as target, we'll use delta by price in 5 time frames
	frameL := 5
	rsiL := 14
	maLL := 25
	maSL := 7

	maxD := 30

	clObservations := make([]float64, 0)

	for i := range data {
		clObservations = append(clObservations, data[i].ClosePrice)
		if i+frameL >= len(data) || i <= maxD {
			continue
		}

		target := data[i+frameL].ClosePrice / data[i].ClosePrice

		treshold := 0.025

		d0 := data[i].ClosePrice / data[i].OpenPrice
		d1 := data[i].ClosePrice / data[i-1].ClosePrice
		d2 := data[i].ClosePrice / data[i-2].ClosePrice
		d3 := data[i].ClosePrice / data[i-3].ClosePrice
		d4 := data[i].ClosePrice / data[i-4].ClosePrice
		d5 := data[i].ClosePrice / data[i-5].ClosePrice
		rsi := indicators.RSI(clObservations, rsiL)

		maS := indicators.SimpleMA(clObservations, maSL)
		maL := indicators.SimpleMA(clObservations, maLL)
		dMA := maS / maL

		slope, _ := indicators.LinearRegressionForTimeSeries(clObservations[len(clObservations)-(1+30):])

		nV := linear.Vector{
			float32(math.Abs(1 - (d0 - 1))), float32(math.Abs(1 - (d1 - 1))), float32(math.Abs(1 - (d2 - 1))), float32(math.Abs(1 - (d3 - 1))), float32(math.Abs(1 - (d4 - 1))), float32(math.Abs(1 - (d5 - 1))),
			float32(rsi / 100), float32(math.Abs(1 - (dMA - 1))), float32(slope * 10),
		}

		features = append(features, nV)

		var nL linear.Vector

		absVal := math.Abs(1 - (target - 1))

		if target > 1 && absVal > treshold {
			nL = linear.Vector{1, 0, 0}
		} else if target < 1 && absVal > treshold {
			nL = linear.Vector{0, 1, 0}
		} else {
			nL = linear.Vector{0, 0, 1}
		}

		labels = append(labels, nL)

	}
	// 9 inp
	// 3 out

	testSampleSize := 10000
	testSample := features[len(features)-testSampleSize:]
	testLabels := labels[len(labels)-testSampleSize:]

	trainSample := features[:len(features)-testSampleSize]
	trainLabels := labels[:len(labels)-testSampleSize]

	// Define Layers
	structure := []int{
		9, // input layer
		12,
		12,
		3, // outputLayer
	}

	activations := []string{
		activation_funcs.Sigmoid,
		activation_funcs.Sigmoid,
		activation_funcs.Sigmoid,
		activation_funcs.Sigmoid,
	}

	learningRate := float32(0.1)

	perceptron := mlp.MLP{
		Introspect: func(s mlp.Step) {
			fmt.Printf("%+v \n", s)
		},
	}
	perceptron.NewMultilayerPerceptron(structure, activations, learningRate)

	someValue, err := perceptron.Train(5, trainSample, trainLabels)
	if err != nil {
		logs.LogError(err)
	}

	fmt.Printf("Loss Value: %f", someValue)

	correct := 0

	for i := range testSample {
		pred := perceptron.Predict(testSample[i])
		actual := testLabels[i]

		_, maxIndex := indicators.Max32(pred)

		if actual[maxIndex] == 1 {
			correct++
		}
	}

	fmt.Printf("\n Correctness: %f%", (float64(correct)/float64(len(testSample)))*100)

	datatatat, err := json.Marshal(perceptron)
	if err != nil {
		logs.LogError(err)
	}

	var renewedPerceptron mlp.MLP

	err = json.Unmarshal(datatatat, &renewedPerceptron)
	if err != nil {
		logs.LogError(err)
	}

}
