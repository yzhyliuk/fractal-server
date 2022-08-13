package controllers

import (
	"github.com/gofiber/fiber/v2"
	"net/http"
	"newTradingBot/models/recording"
	"newTradingBot/models/recording/actions"
	"strconv"
)

type TestingController struct {
	*BaseController
}

func (t *TestingController) GetSessionsForUser(c *fiber.Ctx) error {
	userInfo, err := t.GetUserInfo(c)
	if err != nil {
		return err
	}

	sessions, err := recording.GetAllSessionForUser(t.GetDB(), userInfo.UserID)
	if err != nil {
		return err
	}

	return c.JSON(sessions)
}

func (t *TestingController) StartCapture(c *fiber.Ctx) error {
	var captureInstance recording.CapturedSession
	err := c.BodyParser(&captureInstance)
	if err != nil {
		return err
	}

	userInfo, err := t.GetUserInfo(c)
	if err != nil {
		return err
	}

	captureInstance.UserID = userInfo.UserID

	recordSession, err := recording.CreateSession(t.GetDB(), &captureInstance)
	if err != nil {
		return err
	}

	err = actions.StartCapture(recordSession)
	if err != nil {
		return err
	}

	return c.SendStatus(http.StatusOK)
}

func (t *TestingController) StopCapture(c *fiber.Ctx) error {
	captureSessionID, err := strconv.Atoi(c.Query("sessionId"))
	if err != nil {
		return err
	}

	err = actions.StopRecording(t.GetDB(), captureSessionID)
	if err != nil {
		return err
	}

	return c.SendStatus(http.StatusOK)
}

func (t *TestingController) DeleteCapture(c *fiber.Ctx) error {
	captureSessionID, err := strconv.Atoi(c.Query("sessionId"))
	if err != nil {
		return err
	}

	err = actions.DeleteCaptureSession(t.GetDB(), captureSessionID)
	if err != nil {
		return err
	}

	return c.SendStatus(http.StatusOK)
}
