package api

import (
	"crypto/tls"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"golang.org/x/crypto/acme/autocert"
	"newTradingBot/api/middleware/jwt"
	"newTradingBot/configuration"
)

func StartServer() (*fiber.App, error) {
	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET, PATCH, PUT,HEAD, POST, DELETE, OPTIONS",
		AllowCredentials: true,
	}))

	app.Static("/static", configuration.StaticFilesDir)
	app.Use(recover.New())

	establishRoutes(app)

	// Certificate manager
	m := &autocert.Manager{
		Prompt: autocert.AcceptTOS,
		// Replace with your domain
		HostPolicy: autocert.HostWhitelist("fractal-server.com", "www.fractal-server.com"),
		// Folder to store the certificates
		Cache: autocert.DirCache("certs"),
	}

	// TLS Config
	m.TLSConfig()
	cfg := &tls.Config{
		// Get Certificate from Let's Encrypt
		GetCertificate: m.GetCertificate,
		// By default NextProtos contains the "h2"
		// This has to be removed since Fasthttp does not support HTTP/2
		// Or it will cause a flood of PRI method logs
		// http://webconcepts.info/concepts/http-method/PRI
		NextProtos: []string{
			"http/1.1", "acme-tls/1",
		},
	}

	if configuration.Mode == configuration.Prod {
		ln, err := tls.Listen("tcp", ":443", cfg)
		if err != nil {
			panic(err)
		}

		return app, app.Listener(ln)
	} else {
		return app, app.Listen(":8080")
	}
}

func establishRoutes(app *fiber.App) {
	publicRoutes(app)
	app.Use(jwt.New(jwt.Config{}))
	userRoutes(app)
	strategyRoutes(app)
	testingRoutes(app)
	notificationsRoutes(app)
	neuralNetworkRoutes(app)
}
