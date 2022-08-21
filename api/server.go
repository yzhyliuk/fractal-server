package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"newTradingBot/api/middleware/jwt"
	"newTradingBot/configuration"
)

func StartServer() (*fiber.App, error) {
	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: 	  "*",
		AllowMethods:     "GET, PATCH, PUT, POST, DELETE, OPTIONS",
		AllowCredentials: true,
	}))

	app.Static("/static", configuration.StaticFilesDir)

	app.Use(recover.New())

	establishRoutes(app)

	return app, app.Listen(":8080")
}

func establishRoutes(app *fiber.App)  {
	publicRoutes(app)
	app.Use(jwt.New(jwt.Config{}))
	userRoutes(app)
	strategyRoutes(app)
	testingRoutes(app)
}
