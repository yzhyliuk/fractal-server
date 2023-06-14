package controllers

import (
	"encoding/json"
	errors2 "errors"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"newTradingBot/api/common"
	"newTradingBot/api/errors"
	"newTradingBot/internal_arbitrage"
	"newTradingBot/models/account"
	"newTradingBot/models/apimodels"
	"newTradingBot/models/strategy/actions"
	"newTradingBot/models/strategy/configs"
	"newTradingBot/models/strategy/instance"
	"newTradingBot/models/testing"
	"newTradingBot/models/trade"
	"newTradingBot/models/users"
	"newTradingBot/storage"
	"newTradingBot/strategies/experimental"
	"strconv"
	"strings"
)

type StrategyController struct {
	*BaseController
}

const FuturesType = "futures"

func (s *StrategyController) GetStrategies(c *fiber.Ctx) error {

	db := s.GetFilteredDB(c)

	strategies, err := apimodels.GetAllStrategies(db)
	if err != nil {
		return err
	}

	return c.JSON(strategies)
}

func (s *StrategyController) GetStrategyFields(c *fiber.Ctx) error {
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
	} else {
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

// RunStrategy - runs strategy with configuration from client side
func (s *StrategyController) RunStrategy(c *fiber.Ctx) error {
	strategyID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return err
	}

	//TODO: separate method for this db request
	var sInfo apimodels.StrategyInfo
	err = s.GetDB().Where("id = ?", strategyID).Find(&sInfo).Error
	if err != nil {
		return err
	}

	var commonConfig apimodels.CommonStrategyConfig

	err = c.BodyParser(&commonConfig)
	if err != nil {
		return err
	}

	userinfo, err := s.GetUserInfo(c)
	if err != nil {
		return err
	}

	if !userinfo.Verified {
		return c.JSON(errors.NewBadRequestError("Please confirm your email first"))
	}

	for _, pair := range commonConfig.Pairs {
		commonConfig.Config["pair"] = pair

		rawConfig, err := json.Marshal(&commonConfig.Config)
		if err != nil {
			return err
		}

		if sInfo.IsContinuous {
			err = experimental.RunExperimentalContinuousStrategy(userinfo.UserID, rawConfig, 0, nil)
			if err != nil {
				return err
			}
			return c.SendStatus(200)
		}

		_, instanceID, err := common.RunFunction(userinfo.UserID, rawConfig, testing.Disable, strategyID, sInfo.StrategyName, nil)
		if err != nil {
			return err
		}

		err = instance.CreateConfig(*instanceID, rawConfig, s.GetDB())
		if err != nil {
			return err
		}
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

	err = actions.StopStrategy(s.GetDB(), strategyInstance.StrategyInstance)
	if err != nil {
		return err
	}

	// TODO: if there is no running strategies for monitor - delete it

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

// DeleteSelected - delete list of selected strategies
func (s *StrategyController) DeleteSelected(c *fiber.Ctx) error {
	var ids []int
	err := c.BodyParser(&ids)
	if err != nil {
		return err
	}

	userInfo, err := s.GetUserInfo(c)
	if err != nil {
		return err
	}

	err = instance.DeleteSelectedInstances(s.GetDB(), ids, userInfo.UserID)
	if err != nil {
		return err
	}

	return c.SendStatus(http.StatusOK)
}

// GetTradesForInstance - returns all the trades for given strategy instance
func (s *StrategyController) GetTradesForInstance(c *fiber.Ctx) error {
	instanceID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return err
	}

	userInfo, err := s.GetUserInfo(c)
	if err != nil {
		return err
	}

	trades, err := trade.GetTradesByInstanceID(s.GetDB(), userInfo.UserID, instanceID)
	if err != nil {
		return err
	}

	for _, t := range trades {
		t.ConvertTime()
	}

	return c.JSON(trades)
}

func (s *StrategyController) ArchiveStrategies(c *fiber.Ctx) error {
	var ids []int
	err := c.BodyParser(&ids)
	if err != nil {
		return err
	}

	err = instance.MoveInstancesToArchive(s.GetDB(), ids)
	if err != nil {
		return err
	}

	return c.SendStatus(http.StatusOK)
}

func (s *StrategyController) SaveConfig(c *fiber.Ctx) error {
	strategyID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return err
	}

	var config apimodels.NewStrategyConfig

	err = c.BodyParser(&config)
	if err != nil {
		return err
	}

	rawConfig, err := json.Marshal(&config.Config)
	if err != nil {
		return err
	}

	userinfo, err := s.GetUserInfo(c)
	if err != nil {
		return err
	}

	err = configs.CreateConfig(s.GetDB(), rawConfig, userinfo.UserID, strategyID, config.Name)
	if err != nil {
		return err
	}

	return c.SendStatus(http.StatusOK)
}

func (s *StrategyController) LoadConfigs(c *fiber.Ctx) error {
	strategyID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return err
	}

	userinfo, err := s.GetUserInfo(c)
	if err != nil {
		return err
	}

	cfgs, err := configs.GetConfigsForStrategyPerUser(s.GetDB(), userinfo.UserID, strategyID)
	if err != nil {
		return err
	}

	return c.JSON(cfgs)
}

func (s *StrategyController) DeleteConfig(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return err
	}

	userinfo, err := s.GetUserInfo(c)
	if err != nil {
		return err
	}

	err = configs.DeleteSavedConfig(s.GetDB(), id, userinfo.UserID)
	if err != nil {
		return err
	}

	return c.SendStatus(http.StatusOK)
}

func (s *StrategyController) RunArbitrage(c *fiber.Ctx) error {
	var config struct {
		Target  string `json:"target"`
		Primary string `json:"primary"`
	}

	err := c.BodyParser(&config)
	if err != nil {
		return err
	}

	userinfo, err := s.GetUserInfo(c)
	if err != nil {
		return err
	}

	keys, err := users.GetUserKeys(s.GetDB(), userinfo.UserID)
	if err != nil {
		return err
	}

	acc, err := account.NewBinanceAccount(keys.ApiKey, keys.SecretKey, keys.ApiKey, keys.SecretKey)
	go func() {
		err = internal_arbitrage.RunArbitrageWithParams(config.Target, config.Primary, acc, 15)
	}()

	return nil
}

func (s *StrategyController) GetInstanceConfig(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return err
	}

	config, err := instance.GetInstanceConfig(id, s.GetDB())
	if err != nil {
		return err
	}

	return c.JSON(config)
}

func (s *StrategyController) ChangeConfig(c *fiber.Ctx) error {
	var cfg apimodels.ChangeConfig
	err := c.BodyParser(&cfg)
	if err != nil {
		return err
	}

	insts := &instance.StrategyInstance{
		ID: cfg.InstanceID,
	}
	err = s.GetDB().Find(insts).Error
	if err != nil {
		return err
	}

	userInfo, err := s.GetUserInfo(c)
	if err != nil {
		return err
	}

	if insts.UserID != userInfo.UserID {
		return errors2.New("you don't have permission to perform this action")
	}

	err = storage.StrategiesStorage[cfg.InstanceID].ChangeBid(cfg.Bid)
	if err != nil {
		return err
	}

	return c.SendStatus(http.StatusOK)
}

func (s *StrategyController) CloseCurrentTrade(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return err
	}

	insts := &instance.StrategyInstance{
		ID: id,
	}
	err = s.GetDB().Find(insts).Error
	if err != nil {
		return err
	}

	userInfo, err := s.GetUserInfo(c)
	if err != nil {
		return err
	}

	if insts.UserID != userInfo.UserID {
		return errors2.New("you don't have permission to perform this action")
	}

	storage.StrategiesStorage[id].CloseTrade()

	return c.SendStatus(http.StatusOK)
}
