package mlp

import (
	"math"
	"math/rand"
	"newTradingBot/models/neural_networks/activation_funcs"
	"newTradingBot/models/neural_networks/linear"
)

const InputLayer = "input"
const OutputLayer = "output"

type Layer struct {
	Name               string   `json:"name"`
	Neurons            int      `json:"neurons"`
	ActivationFunction string   `json:"activation_function"`
	LearningRate       *float32 `json:"learningRate"`

	Inputs int `json:"inputs"`

	Weights linear.Frame  `json:"weights"`
	Biases  linear.Vector `json:"biases"`

	Value       linear.Vector `json:"value"`
	Activations linear.Vector `json:"activation_funcs"`

	Loss  linear.Frame  `json:"loss"`
	Error linear.Vector `json:"error"`
}

func (l *Layer) Initialize(neurons, inputs int, learningRate float32, layerName, activationFunction string) {
	l.Neurons = neurons
	l.Name = layerName
	l.Inputs = inputs
	l.ActivationFunction = activationFunction
	l.LearningRate = &learningRate

	// Initialize data structures for use in the network training and
	// predictions, providing them with random initial values.
	l.Weights = make(linear.Frame, l.Neurons)
	for i := range l.Weights {
		l.Weights[i] = make(linear.Vector, l.Inputs)
		for j := range l.Weights[i] {
			// We scale these based on the "connectedness" of the
			// node to avoid saturating the gradients in the
			// network, where really high values do not play nicely
			// with activation functions like sigmoid.
			weight := rand.NormFloat64() *
				math.Pow(float64(l.Inputs), -0.5)
			l.Weights[i][j] = float32(weight)
		}
	}
	l.Biases = make(linear.Vector, l.Neurons)
	for i := range l.Biases {
		l.Biases[i] = rand.Float32()
	}
	// Set up empty error and loss structures for use in backpropagation.
	l.Error = make(linear.Vector, l.Neurons)
	l.Loss = make(linear.Frame, l.Neurons)
	for i := range l.Loss {
		l.Loss[i] = make(linear.Vector, l.Inputs)
	}
}

// FeedForwardPropagation takes in a set of inputs from the previous layer and performs
// forward propagation for the current layer, returning the resulting
// activation_funcs. As a special case, if this Layer has no previous layer and is
// thus the input layer for the network, the values are passed through
// unmodified. Internal state from the calculation is persisted for later use
// in back propagation.
func (l *Layer) FeedForwardPropagation(input linear.Vector) linear.Vector {
	// If this is the input layer, pass through values unmodified.
	if l.Name == InputLayer {
		l.Activations = input
		return input
	}

	// Create vectors with state for each node in this layer.
	Z := make(linear.Vector, l.Neurons)
	activations := make(linear.Vector, l.Neurons)
	// For each node in the layer, perform feed-forward calculation.
	for i := range activations {
		// Vector of weights for each edge to this node, incoming from
		// the previous layer.
		nodeWeights := l.Weights[i]
		// Scalar bias value for the current node index.
		nodeBias := l.Biases[i]
		// Combine input with incoming edge weights, then apply bias.
		Z[i] = linear.DotProduct(input, nodeWeights) + nodeBias
		// Apply activation function for non-linearity.
		activations[i] = activation_funcs.ActivationFunctionsMap[l.ActivationFunction](Z[i])
	}
	// Capture state for use in back-propagation.
	l.Value = Z
	l.Activations = activations
	return activations
}

// BackProp performs the training process of back propagation on the layer for
// the given set of labels. Weights and biases are updated for this layer
// according to the computed error. Internal state on the backpropagation
// process is captured for further backpropagation in earlier layers of the
// network as well.
func (l *Layer) BackProp(label, previousActivations linear.Vector, nextLayerLoss linear.Frame) {
	// ∂L/∂a, derivative Loss w.r.t. activation:
	// 2 ( a1(L) - y1 )
	// First calculate the "Error" vector of last observed error values.
	if l.Name == OutputLayer { // Output layer
		// For the output layer, this is just the difference between
		// output value and label.
		l.Error = l.Activations.Subtract(label)
	} else {
		// Formula for propagated error in hidden layers:
		// ∑0-j ( 2(aj-yj) (g'(zj)) (wj2) )

		// Compute an error for this hidden node by summing up losses
		// attributed to it from the next layer down in the network.
		l.Error = make(linear.Vector, len(l.Error))
		for j := range l.Weights {
			for jn := range nextLayerLoss {
				// Add the loss from node jn in the next layer
				// that came from node j in this layer.
				l.Error[j] += nextLayerLoss[jn][j]
			}
		}
	}
	// Derivative of the squared error w.r.t. activation is applied to the
	// vector of computed errors for each node in this layer.
	dLdA := l.Error.Scalar(2)

	// ∂a/∂z, derivative of activation w.r.t. input:
	// g'(L)(z) ( z1(L) )
	// We apply the derivative of the activation function, specified at
	// network creation time, to the vector of "Z" values for each node
	// captured during forward propagation.
	dAdZ := l.Value.Apply(activation_funcs.ActivationFunctionDerivativesMap[l.ActivationFunction])

	// Capture the loss for this edge for use in the next layer up
	// of backprop. This references and feeds into the lastE
	// calculation above, used in the first derivative term.
	for j := range l.Weights {
		l.Loss[j] = l.Weights[j].Scalar(l.Error[j])
	}

	// Iterate over each weight for node "j" in this layer and "k" in the
	// previous layer and update it according to the computed derivatives.
	for j := range l.Weights {
		for k := range l.Weights[j] {
			// ∂z/∂w, derivative of input w.r.t. weight:
			// a2(L-1)
			// This comes out to the same thing as the formula for
			// the last used activations themselves.
			dZdW := previousActivations[k]

			// Total derivative, via chain rule ∂L/∂w,
			// derivative of loss w.r.t. weight
			dLdW := dLdA[j] * dAdZ[j] * dZdW

			// Update the weight according to the learning rate.
			l.Weights[j][k] -= dLdW * *l.LearningRate
		}
	}

	// Calculate bias updates along the gradient of the loss w.r.t. inputs.
	// ∂L/∂z = ∂L/∂a * ∂a/∂z
	biasUpdate := dLdA.ElementwiseProduct(dAdZ)
	// Update the bias according to the learning rate.
	l.Biases = l.Biases.Subtract(biasUpdate.Scalar(*l.LearningRate))
}
