package controllers

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"newTradingBot/api/common"
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

	trades, err := common.RunStrategy[strategy](userinfo.UserID, rawConfig, testing.BackTest, &session)
	if err != nil {
		return err
	}

	profit, winRate, roi := testing.GetProfitWinRateAndRoiForTrades(trades)

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
