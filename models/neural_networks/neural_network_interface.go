package neural_networks

import "gonum.org/v1/gonum/mat"

type NeuralNetwork interface {
	TrainOnData(x,y *mat.Dense) error
	Predict(x *mat.Dense) (*mat.Dense, error)
}
