package mlp

import (
	"errors"
	"fmt"
	"newTradingBot/models/neural_networks/linear"
)

// MLP provides a Multi-Layer Perceptron which can be configured for
// any network architecture within that paradigm.
type MLP struct {
	// Layers is a list of layers in the network, where the first is the
	// input and last is the output, with inner layers acting as hidden
	// layers.
	//
	// These must not be modified after initialization/training.
	Layers []*Layer `json:"layers"`

	// LearningRate is the rate at which learning occurs in back
	// propagation, relative to the error calculations.
	LearningRate float32 `json:"learningRate"`

	// Introspect provides a way for the caller of this network to
	// check the status of network learning over time and witness
	// convergence (or lack thereof).
	Introspect func(step Step) `json:"-"`
}

// Step captures status updates that happens within a single Epoch, for use in
// introspecting models.
type Step struct {
	// Monotonically increasing counter of which training epoch this step
	// represents.
	Epoch int
	// Loss is the sum of normalized error values to for the epoch using
	// the loss function for the network.
	Loss float32
}

//// Initialize sets up network layers with the needed memory allocations and
//// references for proper operation. It is called automatically during training,
//// provided separately only to facilitate more precise use of the network from
//// a performance analysis perspective.
//func (n *MLP) Initialize() {
//	var prev *Layer
//	for i, layer := range n.Layers {
//		var next *Layer
//		if i < len(n.Layers)-1 {
//			next = n.Layers[i+1]
//		}
//		// Idempotent initialization of the layer, passing in the
//		// previous and next layers for reference in training.
//		layer.initialize(n, prev, next)
//		prev = layer
//	}
//}

func (n *MLP) NewMultilayerPerceptron(layers []int, layersActivations []string, learningRate float32) {
	n.LearningRate = learningRate
	n.Layers = make([]*Layer, len(layers))
	var inputs *int
	for i := range layers {
		layerName := fmt.Sprintf("Hidden Layer %d", i)
		if i == 0 {
			layerName = InputLayer
			inputs = &layers[0]
		} else if i+1 == len(layers) {
			layerName = OutputLayer
		}

		newLayer := &Layer{}
		newLayer.Initialize(layers[i], *inputs, learningRate, layerName, layersActivations[i])

		n.Layers[i] = newLayer
		inputs = &layers[i]
	}
}

// Train takes in a set of inputs and a set of labels and trains the network
// using backpropagation to adjust internal weights to minimize loss, over the
// specified number of epochs. The final loss value is returned after training
// completes.
func (n *MLP) Train(epochs int, inputs, labels linear.Frame) (float32, error) {
	// Validate that the inputs match the network configuration.
	if err := n.check(inputs, labels); err != nil {
		return 0, err
	}

	// Run the training process for the specified number of epochs.
	var loss float32
	for e := 0; e < epochs; e++ {
		predictions := make(linear.Frame, len(inputs))

		// Iterate over each inputs to train in SGD fashion.
		for i, input := range inputs {
			// Iterate FORWARDS through the network.
			activations := input
			for _, layer := range n.Layers {
				activations = layer.FeedForwardPropagation(activations)
			}
			predictions[i] = activations

			// Iterate BACKWARDS through the network.
			for step := range n.Layers {
				l := len(n.Layers) - (step + 1)
				layer := n.Layers[l]

				if l == 0 {
					// If we are at the input layer, nothing to do.
					continue
				}

				var nextLayerLoss linear.Frame
				var prevLayerActivation linear.Vector
				if step != 0 {
					nextLayerLoss = n.Layers[l+1].Loss
				}
				prevLayerActivation = n.Layers[l-1].Activations

				layer.BackProp(labels[i], prevLayerActivation, nextLayerLoss)
			}
		}

		// Calculate loss
		loss = Loss(predictions, labels)
		if n.Introspect != nil {
			n.Introspect(Step{
				Epoch: e,
				Loss:  loss,
			})
		}

	}

	return loss, nil
}

// Predict takes in a set of input rows with the width of the input layer, and
// returns a frame of prediction rows with the width of the output layer,
// representing the predictions of the network.
func (n *MLP) Predict(inputs linear.Vector) linear.Vector {
	// Iterate FORWARDS through the network
	activations := inputs
	for _, layer := range n.Layers {
		activations = layer.FeedForwardPropagation(activations)
	}
	// Activations from the last layer are our predictions
	return activations
}

func (n *MLP) check(inputs linear.Frame, outputs linear.Frame) error {
	if len(n.Layers) == 0 {
		return errors.New("ann must have at least one layer")
	}

	if len(inputs) != len(outputs) {
		return fmt.Errorf(
			"inputs count %d mismatched with outputs count %d",
			len(inputs), len(outputs),
		)
	}
	return nil
}

// Loss function, mean squared error.
//
// Mean(Error^2)
func Loss(pred, labels linear.Frame) float32 {
	var squaredError, count float32
	pred.ForEachPairwise(labels, func(o, l float32) {
		count += 1.0
		// squared error
		squaredError += (o - l) * (o - l)
	})
	return squaredError / count
}
