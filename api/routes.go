package api

import (
	"github.com/gofiber/fiber/v2"
	"newTradingBot/api/controllers"
)

func publicRoutes(app *fiber.App)  {
	userController := &controllers.UserController{}
	authController := &controllers.AuthController{}
	base := &controllers.BaseController{}

	app.Get("/ping", base.Ping)

	app.Post("/auth", authController.Login)
	usersGroup := app.Group("/users")
	usersGroup.Post("/new", userController.CreateUser)

}

func userRoutes(app *fiber.App)  {
	userController := &controllers.UserController{}
	usersGroup := app.Group("/users")
	usersGroup.Get("/get-finances", userController.GetUserBalance)
	usersGroup.Get("/my-info", userController.GetUser)
	usersGroup.Post("/set-keys", userController.SetKeys)
	usersGroup.Get("/get-keys", userController.GetKeys)
}

func strategyRoutes(app *fiber.App)  {
	strategyController := &controllers.StrategyController{}
	strategiesGroup := app.Group("/strategies")
	strategiesGroup.Get("/list", strategyController.GetStrategies)
	strategiesGroup.Get("/fields", strategyController.GetStrategyFields)
	strategiesGroup.Get("/pairs", strategyController.GetPairs)
	strategiesGroup.Post("/run/:id", strategyController.RunStrategy)
	strategiesGroup.Get("/instances", strategyController.GetInstances)
	strategiesGroup.Get("/instances/:id", strategyController.GetInstance)
	strategiesGroup.Delete("/instances/:id", strategyController.Delete)
	strategiesGroup.Post("/instances/delete-selected", strategyController.DeleteSelected)
	strategiesGroup.Get("/instances/:id/trades",strategyController.GetTradesForInstance)
	strategiesGroup.Get("/instances/:id/stop", strategyController.StopStrategy)
	strategiesGroup.Post("/run-arbitrage", strategyController.RunArbitrage)
	strategiesGroup.Post("/move-to-archive", strategyController.ArchiveStrategies)
}