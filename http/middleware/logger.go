package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/minisource/go-common/logging"
)

// DefaultStructuredLogger initializes the safe structured access log
// middleware. Body capture is OFF by default; only sanitized metadata is
// logged. Kept for backward compatibility — prefer AccessLog with an explicit
// config built from LoadAccessLogConfigFromEnv.
func DefaultStructuredLogger(cfg *logging.LoggerConfig) fiber.Handler {
	logger := logging.NewLogger(cfg)
	acfg := DefaultAccessLogConfig("")
	acfg.Logger = logger
	return AccessLog(acfg)
}
