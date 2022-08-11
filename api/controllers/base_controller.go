package controllers

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"log"
	"newTradingBot/api/database"
	"newTradingBot/api/helpers"
	"newTradingBot/models/auth"
	"strconv"
	"strings"
)

type BaseController struct {

}

func (b *BaseController) GetDB() *gorm.DB {
	db, err := database.GetDataBaseConnection()
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func (b *BaseController) GetUserInfo(c *fiber.Ctx) (*auth.Payload, error) {
	// TODO: type assertion
	uInfo := c.Locals(auth.UserInfo)
	var userInfo auth.Payload
	err := helpers.DeepCopyJSON(&uInfo, &userInfo)
	if err != nil {
		return nil, err
	}
	return &userInfo, nil
}

func (b *BaseController) GetFilteredDB(c *fiber.Ctx) *gorm.DB {
	query := map[string]string{}
	str := string(c.Request().URI().QueryString())
	if len(str) == 0 {
		return b.GetDB()
	}
	params := strings.Split(str, "&")
	for _, param := range params {
		keyValue := strings.Split(param, "=")
		key := keyValue[0]
		//if key == helpers.Pagination ||
		//	key == helpers.ItemsPerPage ||
		//	key == helpers.CurrentPage {
		//	continue
		//}
		query[key] = c.Query(key)
	}

	db := b.GetDB()

	for key, value := range query {
		if value == "true" || value == "false" {
			db = db.Where(fmt.Sprintf("%s = %s",key, value))
			continue
		}
		_, err := strconv.Atoi(value)
		if err == nil {
			db = db.Where(fmt.Sprintf(`%s = %s`, key, value))
		} else {
			db = db.Where(fmt.Sprintf(`LOWER(%s) LIKE LOWER('%s')`, key, fmt.Sprintf("%%%s%%", value)))
		}
	}

	return db
}

func (b *BaseController) Ping(c *fiber.Ctx) error {
	return c.SendString("Pong")
}
