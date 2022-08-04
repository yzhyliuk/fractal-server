package jwt

import (
	"github.com/gofiber/fiber/v2"
	"net/http"
	"newTradingBot/models/auth"
	"time"
)

type Config struct {
	Filter       func(*fiber.Ctx) bool
	SignatureKey string
}

func New(config ...Config) func(*fiber.Ctx) error {
	// Init configuration
	var cfg Config
	if len(config) > 0 {
		cfg = config[0]
	}
	return func(c *fiber.Ctx) error {
		// Filter request to skip middleware
		if cfg.Filter != nil && cfg.Filter(c) {
			return c.Next()
		}

		token := c.Cookies(auth.AToken, "")
		valid, err := auth.VerifyToken(token)
		if err != nil {
			return err
		}
		if valid {
			// Get tkp payload data
			tp, err := auth.GetTokenPayload(token)
			if err != nil {
				return err
			}
			if tp.Expires.Before(time.Now()) {
				return c.SendStatus(http.StatusUnauthorized)

			}
			c.Locals(auth.UserInfo, tp)
			return c.Next()
		}
		return c.SendStatus(http.StatusUnauthorized)
	}
}


