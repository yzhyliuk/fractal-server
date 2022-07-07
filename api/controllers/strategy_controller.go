package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"newTradingBot/api/common"
	"newTradingBot/models/apimodels"
	"newTradingBot/models/strategy/instance"
	"newTradingBot/models/trade"
	"newTradingBot/storage"
	"strconv"
	"strings"
)

type StrategyController struct {
	*BaseController
}

func (s *StrategyController) GetStrategies(c *fiber.Ctx) error  {

	db := s.GetFilteredDB(c)

	strategies, err := apimodels.GetAllStrategies(db)
	if err != nil {
		return err
	}

	return c.JSON(strategies)
}

func (s *StrategyController) GetStrategyFields(c *fiber.Ctx) error  {
	id, err := strconv.Atoi(c.Query("id"))
	if err != nil {
		return err
	}

	fields, err := apimodels.GetStrategyFields(s.GetDB(), id)
	if err != nil {
		return err
	}

	defaultFields, err := apimodels.GetDefaultFields(s.GetDB())
	if err != nil {
		return err
	}

	fields = append(defaultFields, fields...)

	return c.JSON(fields)
}

func (s *StrategyController) GetPairs(c *fiber.Ctx) error {
	filter := c.Query("filter")
	typeOfPairs := c.Query("type")

	var tradingPairs = make([]apimodels.TradingPair, 0)

	tradingPairsSource := make([]apimodels.TradingPair, 0)

	if typeOfPairs == "futures" {
		tradingPairsSource = common.FuturesTradingPairs
	} else  {
		tradingPairsSource = common.SpotTradingPairs
	}

	if filter == "" {
		return c.JSON(tradingPairsSource)
	}

	for _, pair := range tradingPairsSource {
		if strings.Contains(pair.Value, filter) {
			tradingPairs = append(tradingPairs, pair)
		}
	}

	return c.JSON(tradingPairs)
}


// RunStrategy - runs strategy with config from client side
func (s *StrategyController) RunStrategy(c *fiber.Ctx) error {
	strategyID, err := strconv.Atoi(c.Params("id"))
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

	userinfo, err := s.GetUserInfo(c)
	if err != nil {
		return err
	}

	err = common.RunStrategy[strategyID](userinfo.UserID, rawConfig)
	if err != nil {
		return err
	}

	return nil
}

// StopStrategy - stops strategy instance
func (s *StrategyController) StopStrategy(c *fiber.Ctx) error {
	instanceID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return err
	}

	strategyInstance, err := instance.GetInstanceByID(s.GetDB(), instanceID)
	if err != nil {
		return err
	}

	userInfo, err := s.GetUserInfo(c)
	if err != nil {
		return err
	}

	if strategyInstance.UserID != userInfo.UserID {
		return err
	}

	err = instance.UpdateStatus(s.GetDB(), instanceID, instance.StatusStopped)
	if err != nil {
		return err
	}

	monitorName := fmt.Sprintf("%s:%d:%t",strategyInstance.Pair, strategyInstance.TimeFrame, strategyInstance.IsFutures)

	storage.MonitorsBinance[monitorName].UnSubscribe(strategyInstance.ID)

	go storage.StrategiesStorage[instanceID].Stop()
	delete(storage.StrategiesStorage, instanceID)

	return c.SendStatus(http.StatusOK)
}

// GetInstances - returns list of all instances for user
func (s *StrategyController) GetInstances(c *fiber.Ctx) error {
	userInfo, err := s.GetUserInfo(c)
	if err != nil {
			return err
	}

	instances, err := instance.ListInstancesForUser(s.GetFilteredDB(c), userInfo.UserID)
	if err != nil {
		return err
	}

	return c.JSON(instances)
}

// GetInstance - get strategy monitoring instance by its id
func (s *StrategyController) GetInstance(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return err
	}

	strategyInstance, err := instance.GetInstanceByID(s.GetDB(), id)
	if err != nil {
		return err
	}

	return c.JSON(strategyInstance)
}

// Delete - delete instance from data base
func (s *StrategyController) Delete(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return err
	}

	userInfo, err := s.GetUserInfo(c)
	if err != nil {
		return err
	}

	err = instance.DeleteInstance(s.GetDB(), id, userInfo.UserID)
	if err != nil {
		return err
	}

	return c.SendStatus(http.StatusOK)
}


// GetTradesForInstance - returns all the trades for given strategy instance
func (s *StrategyController) GetTradesForInstance(c *fiber.Ctx) error  {
	instanceID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return err
	}

	userInfo, err := s.GetUserInfo(c)
	if err != nil {
		return err
	}

	trades, err := trade.GetTradesByInstanceID(s.GetDB(),userInfo.UserID, instanceID)
	if err != nil {
		return err
	}

	for _, t := range trades {
		t.ConvertTime()
	}

	return c.JSON(trades)
}