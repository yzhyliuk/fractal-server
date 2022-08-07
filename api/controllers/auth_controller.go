package controllers

import (
	"github.com/gofiber/fiber/v2"
	"net/http"
	"newTradingBot/api/security"
	"newTradingBot/models/auth"
	"newTradingBot/models/users"
)

type AuthController struct {
	*BaseController
}

func (a *AuthController) Login(c *fiber.Ctx) error {
	var userCredentials users.UserCredentials
	err := c.BodyParser(&userCredentials)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return err
	}

	userData, err := users.GetByEmail(a.GetDB(), userCredentials.Email)
	if err != nil {
		return err
	}

	valid, err := security.VerifyHashedString(userCredentials.Password, userData.Password)
	if err != nil {
		return err
	}

	if valid {
		token, err := auth.CreateToken(userData)
		if err != nil {
			return err
		}
		c.Cookie(&fiber.Cookie{
			Name:     auth.AToken,
			Value:    *token,
			HTTPOnly: true,
			Secure: true,
			SameSite: "None",
		})
		return nil
	}

	return c.SendStatus(http.StatusBadRequest)
}