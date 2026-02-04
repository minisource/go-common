package middleware

import (
	"errors"
	"net/http"
	"time"

	helper "github.com/minisource/go-common/http/helper"
	"github.com/minisource/go-common/limiter"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/time/rate"
)

type OtpConfig struct {
	ExpireTime time.Duration
	Digits     int
	Limiter    time.Duration
}

func OtpLimiter(cfg *OtpConfig) fiber.Handler {
	ipLimiter := limiter.NewIPRateLimiter(rate.Every(cfg.Limiter*time.Second), 1)
	return func(c *fiber.Ctx) error {
		// Get the IP address of the client
		clientIP := c.IP()

		// Get the rate limiter for the client's IP
		limiter := ipLimiter.GetLimiter(clientIP)

		// Check if the request is allowed
		if !limiter.Allow() {
			httpResponse := helper.GenerateBaseResponseWithError(nil, false, helper.OtpLimiterError, errors.New("not allowed"))
			return c.Status(http.StatusTooManyRequests).JSON(httpResponse)
		}

		// Continue to the next handler
		return c.Next()
	}
}