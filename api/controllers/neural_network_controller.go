package controllers

import (
	"github.com/gofiber/fiber/v2"
	"newTradingBot/models/neural_networks"
)

type NeuralNetworkController struct {
	*BaseController
}

func (n *NeuralNetworkController) RunTestOnNeuralNetwork(ctx *fiber.Ctx) error {
	neural_networks.TestMlp(n.GetDB())
	return nil
}

func (n *NeuralNetworkController) CreateModel(ctx *fiber.Ctx) error {
	return nil
}
