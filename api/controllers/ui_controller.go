package controllers

import (
	"github.com/gofiber/fiber/v2"
	"newTradingBot/models/apimodels"
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
