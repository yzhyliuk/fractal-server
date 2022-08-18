package api

import (
	"github.com/gofiber/fiber/v2"
	"newTradingBot/api/controllers"
)

func publicRoutes(app *fiber.App)  {
	userController := &controllers.UserController{}
	authController := &controllers.AuthController{}
	uiController := &controllers.UiController{}
	base := &controllers.BaseController{}

	app.Get("/ping", base.Ping)
	app.Get("/ui/:form", uiController.GetFormFields)

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
	usersGroup.Post("/update", userController.UpdateUser)
	usersGroup.Post("/permission", userController.CreatePermission)
	usersGroup.Put("/permission", userController.UpdateUserPermission)
	usersGroup.Get("/permission", userController.GetAllowedUsers)
	usersGroup.Post("/permission/delete", userController.DeletePermission)
}

func testingRoutes(app *fiber.App) {
	testingController := &controllers.TestingController{}
	testingGroup := app.Group("/testing")
	testingGroup.Post("/start-capture", testingController.StartCapture)
	testingGroup.Get("/stop-capture", testingController.StopCapture)
	testingGroup.Get("/sessions", testingController.GetSessionsForUser)
	testingGroup.Get("/delete-session", testingController.DeleteCapture)
	testingGroup.Post("/back-test/:strategy/:session", testingController.RunBackTest)
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