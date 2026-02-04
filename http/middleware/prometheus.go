package middleware

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/minisource/go-common/metrics"
)


func Prometheus() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		path := c.Route().Path
		method := c.Method()
		err := c.Next()
		status := c.Response().StatusCode()
		metrics.HttpDuration.WithLabelValues(path, method, strconv.Itoa(status)).
			Observe(float64(time.Since(start) / time.Millisecond))
		return err
	}
}
