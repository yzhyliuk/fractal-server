package controllers

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"net/http"
	"newTradingBot/logs"
	"newTradingBot/models/auth"
	"newTradingBot/models/notifications"
	"time"
)

type NotificationController struct {
	*BaseController
}

func (n *NotificationController) ListNotificationsForUser(c *fiber.Ctx) error {
	userInfo, err := n.GetUserInfo(c)
	if err != nil {
		return err
	}

	nf, err := notifications.ListNotificationsForUser(n.GetDB(), userInfo.UserID)
	if err != nil {
		return err
	}

	return c.JSON(nf)
}

func (n *NotificationController) DismissAll(c *fiber.Ctx) error {
	userInfo, err := n.GetUserInfo(c)
	if err != nil {
		return err
	}

	err = notifications.DismissAll(n.GetDB(), userInfo.UserID)
	if err != nil {
		return err
	}

	return c.SendStatus(http.StatusOK)
}

func (n *NotificationController) NotificationsWS(c *websocket.Conn)  {
	notificationList := make([]*notifications.Notification, 0)

	for {

		ui := c.Locals("userInfo")
		userInfo := ui.(*auth.Payload)

	    nfList, err := notifications.ListNotificationsForUser(n.GetDB(),userInfo.UserID)
		if err != nil{
			logs.LogError(err)
		}

		if len(notificationList) != len(nfList) {

			notificationList = nfList

			bytes, err := json.Marshal(&nfList)
			if err != nil {
				logs.LogError(err)
			}

			err = c.WriteMessage(websocket.TextMessage, bytes)
			if err != nil {
				logs.LogError(err)
				return
			}
		}

		time.Sleep(10*time.Second)
	}
}
