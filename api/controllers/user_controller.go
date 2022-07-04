package controllers

import (
	"github.com/gofiber/fiber/v2"
	"net/http"
	"newTradingBot/models/auth"
	"newTradingBot/models/users"
)

type UserController struct {
	*BaseController
}

func (u *UserController) CreateUser(c *fiber.Ctx) error {
	var newUser users.NewUser
	err := c.BodyParser(&newUser)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return err
	}

	user, err := users.Create(u.GetDB(), &newUser)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return err
	}

	token, err := auth.CreateToken(user)
	if err != nil {
		return err
	}
	c.Cookie(&fiber.Cookie{
		Name:     auth.AToken,
		Value:    *token,
		HTTPOnly: true,
	})

	return c.JSON(user)
}


// GetUser TODO : Remove this test method
func (u *UserController) GetUser(c *fiber.Ctx) error  {
	userInfo, err := u.GetUserInfo(c)
	if err != nil {
		return c.SendStatus(http.StatusUnauthorized)
	}



	return c.JSON(userInfo)
}

func (u *UserController) SetKeys(c *fiber.Ctx) error  {
	userInfo, err := u.GetUserInfo(c)
	if err != nil {
		return err

	}

	keysFromUser := users.Keys{}

	err = c.BodyParser(&keysFromUser)
	if err != nil {
		return err
	}

	keysFromUser.UserID = userInfo.UserID

	err = users.SaveUserKeys(u.GetDB(), &keysFromUser)
	if err != nil {
		return err
	}

	return c.SendStatus(http.StatusOK)
}

func (u *UserController) GetKeys(c *fiber.Ctx) error  {
	userInfo, err := u.GetUserInfo(c)
	if err != nil {
		return err

	}

	keys, err := users.GetUserKeys(u.GetDB(), userInfo.UserID)
	if err != nil {
		return err
	}

	keys.HideKeys()

	return c.JSON(keys)
}
