package controllers

import (
	"github.com/gofiber/fiber/v2"
	"newTradingBot/models/apimodels"
	"newTradingBot/models/neural_networks/activation_funcs"
	"newTradingBot/models/neural_networks/input_params"
)

type UiController struct {
	*BaseController
}

func (u *UiController) GetFormFields(c *fiber.Ctx) error {
	formName := c.Params("form")

	fields, err := apimodels.GetFieldsByFormName(u.GetDB(), formName)
	if err != nil {
		return err
	}

	return c.JSON(fields)
}

func (u *UiController) GetNeuralNetworkParams(c *fiber.Ctx) error {
	return c.JSON(input_params.GetInputParams())
}

func (u *UiController) GetActivationsParams(c *fiber.Ctx) error {
	return c.JSON(activation_funcs.GetActivationFunctionsParams())
}
