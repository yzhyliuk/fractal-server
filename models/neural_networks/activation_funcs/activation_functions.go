package activation_funcs

import (
	"math"
	"newTradingBot/models/apimodels"
)

var activations = []apimodels.UiParam{
	{"Sigmoid", Sigmoid}, {"Tanh", Tanh},
}

func GetActivationFunctionsParams() []apimodels.UiParam {
	return activations
}

// Sigmoid - used to represent 'sigmoid' activation function
const Sigmoid = "sigmoid"
const Tanh = "tanh"

// ActivationFunctionsMap - maps named constants to actual activation functions
var ActivationFunctionsMap = map[string]func(x float32) float32{
	Sigmoid: sigmoid,
	Tanh:    tanh,
}

// ActivationFunctionDerivativesMap maps named constants to actual activation functions derivatives
var ActivationFunctionDerivativesMap = map[string]func(x float32) float32{
	Sigmoid: sigmoidPrime,
	Tanh:    tanhPrime,
}

// sigmoid implements the sigmoid function
// for use in activation functions.
func sigmoid(x float32) float32 {
	return 1 / (1 + float32(math.Exp(-float64(x))))
}

// sigmoidPrime implements the derivative
// of the sigmoid function for backpropagation.
func sigmoidPrime(x float32) float32 {
	return sigmoid(x) * (1.0 - sigmoid(x))
}

func tanh(x float32) float32 {
	return (float32(math.Exp(float64(x))) - float32(math.Exp(float64(-x)))) / (float32(math.Exp(float64(x))) + float32(math.Exp(float64(-x))))
}

func tanhPrime(x float32) float32 {
	return 1 - tanh(x)*tanh(x)
}
