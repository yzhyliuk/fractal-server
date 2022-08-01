package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"newTradingBot/api/middleware/jwt"
)

func StartServer() (*fiber.App, error) {
	app := fiber.New()
	app.Use(cors.New(cors.Config{
		//AllowOrigins:     "http://localhost:3000",
		AllowOrigins:     "*",
		AllowHeaders:     "Origin, X-Requested-With, Content-Type, Accept, x-access-token, X-Auth-Token",
		AllowMethods:     "GET, PUT, POST, DELETE, OPTIONS",
		AllowCredentials: true,
	}))

	app.Use(recover.New())

	establishRoutes(app)

	return app, app.Listen(":8080")
}

func establishRoutes(app *fiber.App)  {
	publicRoutes(app)
	app.Use(jwt.New(jwt.Config{}))
	userRoutes(app)
	strategyRoutes(app)

}
