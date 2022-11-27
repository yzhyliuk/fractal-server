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
	"newTradingBot/models/trade"
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
	} else  {
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

	err = actions.DeleteCaptureSession(t.GetDB(), userinfo.UserID,captureSessionID)
	if err != nil {
		return err
	}

	return c.SendStatus(http.StatusOK)
}

func (t *TestingController) RunBackTest(c *fiber.Ctx) error {
	strategy, err := strconv.Atoi(c.Params("strategy"))
	if err != nil {
		return err
	}

	session, err := strconv.Atoi(c.Params("session"))
	if err != nil {
		return err
	}

	var config interface{}

	err = c.BodyParser(&config)
	if err != nil {
		return err
	}

	rawConfig, err := json.Marshal(&config)
	if err != nil {
		return err
	}

	userinfo, err := t.GetUserInfo(c)
	if err != nil {
		return err
	}

	trades, _, err := common.RunStrategy[strategy](userinfo.UserID, rawConfig, testing.BackTest, &session)
	if err != nil {
		return err
	}

	profit, winRate, roi := testing.GetProfitWinRateAndRoiForTrades(trades)

	if len(trades) > 200 {
		trades = trades[:200]
	}

	return c.JSON(struct {
		Profit float64 `json:"profit"`
		WinRate float64 `json:"winRate"`
		Roi float64 `json:"roi"`
		TradesClosed int `json:"tradesClosed"`
		Trades []*trade.Trade `json:"trades"`
	}{
		Profit: profit,
		WinRate: winRate,
		Roi: roi,
		TradesClosed: len(trades),
		Trades: trades,
	})
}

func (t *TestingController) HandleWS(c *websocket.Conn) {
	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			logs.LogDebug("", err)
			return
		}

		var BTM apimodels.BackTestModel

		err = json.Unmarshal(msg, &BTM)
		if err != nil {
			logs.LogDebug("", err)
			return
		}

		ui := c.Locals("userInfo")
		userInfo := ui.(*auth.Payload)


		rawConfig, err := json.Marshal(&BTM.Config)
		if err != nil {
			logs.LogDebug("", err)
			return
		}

		trades, _, err := common.RunStrategy[BTM.StrategyID](userInfo.UserID, rawConfig, testing.BackTest, &BTM.CaptureID)
		if err != nil {
			return
		}

		profit, winRate, roi := testing.GetProfitWinRateAndRoiForTrades(trades)
		tradesClosed := len(trades)

		if tradesClosed > 200 {
			trades = trades[:200]
		}

		res := struct {
			Profit float64 `json:"profit"`
			WinRate float64 `json:"winRate"`
			Roi float64 `json:"roi"`
			TradesClosed int `json:"tradesClosed"`
			Trades []*trade.Trade `json:"trades"`
		}{
			Profit: profit,
			WinRate: winRate,
			Roi: roi,
			TradesClosed: tradesClosed,
			Trades: trades,
		}


		resp := struct {
			Data interface{} `json:"data"`
			Status int `json:"status"`
		}{
			Data: res,
			Status: http.StatusOK,
		}

		respBytes, err := json.Marshal(&resp)
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