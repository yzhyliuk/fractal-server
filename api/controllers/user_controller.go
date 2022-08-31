package controllers

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"newTradingBot/configuration"
	"newTradingBot/models/auth"
	"newTradingBot/models/permissions"
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


// GetUser - returns user Info or 401 error TODO : Remove this test method
func (u *UserController) GetUser(c *fiber.Ctx) error  {
	userInfo, err := u.GetUserInfo(c)
	if err != nil {
		return c.SendStatus(http.StatusUnauthorized)
	}

	use, err := users.GetUserByID(u.GetDB(), userInfo.UserID)
	if err != nil {
		return err
	}

	return c.JSON(use)
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

func (u *UserController) UpdateUser(c *fiber.Ctx) error {
	var userUpdated *users.User

	err := c.BodyParser(&userUpdated)
	if err != nil {
		return err
	}

	userInfo, err := u.GetUserInfo(c)
	if err != nil {
		return err
	}

	err = users.UpdateUserName(u.GetDB(), userInfo.UserID, userUpdated.Username)
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

func (u *UserController) GetUserBalance(c *fiber.Ctx) error {
	userInfo, err := u.GetUserInfo(c)
	if err != nil {
		return err
	}

	finance, err := users.GetUserFinances(u.GetDB(), userInfo.UserID)
	if err != nil {
		return err
	}

	return c.JSON(finance)

}

func (u *UserController) CreatePermission(c *fiber.Ctx) error {
	userInfo, err := u.GetUserInfo(c)
	var permissisonTable permissions.PermissionTable

	err = c.BodyParser(&permissisonTable)
	if err != nil {
		return err
	}

	childUser, err := users.GetUserByUsername(u.GetDB(), permissisonTable.ChildUsername)
	if err != nil {
		return err
	}

	permissisonTable.OriginUser = userInfo.UserID
	permissisonTable.ChildUser = childUser.ID

	err = permissions.CreatePermission(u.GetDB(), permissisonTable)
	if err != nil {
		return err
	}

	return c.SendStatus(http.StatusOK)
}

func (u *UserController) UpdateUserPermission(c *fiber.Ctx) error {
	userInfo, err := u.GetUserInfo(c)
	var permissisonTable permissions.PermissionTable

	err = c.BodyParser(&permissisonTable)
	if err != nil {
		return err
	}

	permissisonTable.OriginUser = userInfo.UserID

	err = permissions.UpdatePermissionTable(u.GetDB(), permissisonTable)
	if err != nil {
		return err
	}

	return c.SendStatus(http.StatusOK)
}

func (u *UserController) GetAllowedUsers(c *fiber.Ctx) error {
	userInfo, err := u.GetUserInfo(c)
	if err != nil {
		return err
	}

	permissionsTable, err := permissions.GetAllowedUsers(u.GetDB(), userInfo.UserID)
	if err != nil {
		return err
	}

	return c.JSON(permissionsTable)
}

func (u *UserController) DeletePermission(c *fiber.Ctx) error {
	userInfo, err := u.GetUserInfo(c)
	var permissisonTable permissions.PermissionTable

	err = c.BodyParser(&permissisonTable)
	if err != nil {
		return err
	}

	permissisonTable.OriginUser = userInfo.UserID

	err = permissions.DeletePermissionTable(u.GetDB(), permissisonTable)
	if err != nil {
		return err
	}

	return c.SendStatus(http.StatusOK)
}

func (u *UserController) GetUserInfoDashboard(c *fiber.Ctx) error {
	userInfo, err := u.GetUserInfo(c)
	if err != nil {
		return err
	}

	dashboardInfo, err := users.GetUserDashboardInfo(u.GetDB(),userInfo.UserID)
	if err != nil {
		return err
	}

	return c.JSON(dashboardInfo)
}

func (u *UserController) GetUserStats(c *fiber.Ctx) error {
	userInfo, err := u.GetUserInfo(c)
	if err != nil {
		return err
	}

	stats, err := users.GetUserStats(u.GetDB(), userInfo.UserID)
	if err != nil {
		return err
	}

	return c.JSON(stats)
}

func (u *UserController) UploadPhoto(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return err
	}

	userInfo, err := u.GetUserInfo(c)
	if err != nil {
		return err
	}

	user, err := users.GetUserByID(u.GetDB(), userInfo.UserID)
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("%s-%s", user.Username, file.Filename)

	err = c.SaveFile(file, fmt.Sprintf("%s/%s",configuration.StaticFilesDir, filename))
	if err != nil {
		return err
	}

	err = users.UpdateProfilePhoto(u.GetDB(), user.ID, filename)
	if err != nil {
		return err
	}

	return c.SendStatus(http.StatusOK)
}
