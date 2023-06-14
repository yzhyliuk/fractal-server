package controllers

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"net/http"
	"newTradingBot/api/common"
	"newTradingBot/api/errors"
	"newTradingBot/logs"
	"newTradingBot/models/apimodels"
	"newTradingBot/models/auth"
	"newTradingBot/models/recording"
	"newTradingBot/models/recording/actions"
	"newTradingBot/models/testing"
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

	withPermissions := c.Query("withPermissions")

	var sessions []*recording.CapturedSession

	if withPermissions == "true" {
		sessions, err = recording.GetSessionsForUserWithPermissions(t.GetDB(), userInfo.UserID)
	} else {
		sessions, err = recording.GetAllSessionForUser(t.GetDB(), userInfo.UserID)
	}

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

	if !userInfo.Verified {
		return c.JSON(errors.NewBadRequestError("Please confirm your email first"))
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

	userinfo, err := t.GetUserInfo(c)
	if err != nil {
		return err
	}

	err = actions.DeleteCaptureSession(t.GetDB(), userinfo.UserID, captureSessionID)
	if err != nil {
		return err
	}

	return c.SendStatus(http.StatusOK)
}

func (t *TestingController) GetSelectDataOptions(c *fiber.Ctx) error {
	userinfo, err := t.GetUserInfo(c)
	if err != nil {
		return err
	}

	capt, err := actions.GetUsersDataOptions(t.GetDB(), userinfo.UserID)
	if err != nil {
		return err
	}

	return c.JSON(capt)
}

func (t *TestingController) HandleWS(c *websocket.Conn) {
}

func (t *TestingController) HandleWSv2(c *websocket.Conn) {
	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			logs.LogDebug("", err)
			return
		}

		var model apimodels.MassBackTesting

		err = json.Unmarshal(msg, &model)
		if err != nil {
			logs.LogDebug("", err)
			return
		}

		ui := c.Locals("userInfo")
		userInfo := ui.(*auth.Payload)

		//testing.RunTestsForAll(t.GetDB(), BTM, userInfo)

		// TODO : somehow rework this part of code to fit in separate method, currently it's causing import loop cycle

		rawConfig, err := json.Marshal(&model.Config)
		if err != nil {
			logs.LogDebug("", err)
			return
		}

		//TODO: separate method for this db request
		var sInfo apimodels.StrategyInfo
		err = t.GetDB().Where("id = ?", model.StrategyID).Find(&sInfo).Error
		if err != nil {
			logs.LogDebug("", err)
			return
		}

		captures, err := actions.GetCaptures(t.GetDB(), model.Pair, model.TimeFrame, userInfo.UserID)
		if err != nil {
			logs.LogDebug("", err)
			return
		}

		result := make([]*apimodels.BackTestingResult, 0)

		for _, capture := range captures {
			trades, _, err := common.RunFunction(userInfo.UserID, rawConfig, testing.BackTest, sInfo.ID, sInfo.StrategyName, &capture.ID)
			if err != nil {
				logs.LogDebug("", err)
				return
			}

			result = append(result, &apimodels.BackTestingResult{
				Trades:    trades,
				Pair:      capture.Symbol,
				TimeFrame: capture.TimeFrame,
			})
		}

		metrics, err := testing.GetMetricsForTrades(result)
		if err != nil {
			logs.LogError(err)
			return
		}

		respBytes, err := json.Marshal(&metrics)
		if err != nil {
			return
		}

		err = c.WriteMessage(websocket.TextMessage, respBytes)
		if err != nil {
			logs.LogDebug("", err)
			return
		}
	}
}
