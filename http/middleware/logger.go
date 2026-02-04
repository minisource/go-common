package middleware

import (
	"bytes"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/minisource/go-common/logging"
	"github.com/valyala/fasthttp"
)

// bodyLogWriter is a wrapper to capture the response body
type bodyLogWriter struct {
	*fasthttp.Response
	body *bytes.Buffer
}

// Read implements the io.Reader interface for bodyLogWriter
func (w *bodyLogWriter) Read(p []byte) (int, error) {
	return w.body.Read(p)
}

// Write captures the response body and writes it to the original response writer
func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.Response.BodyWriter().Write(b)
}

// DefaultStructuredLogger initializes the structured logger middleware with a given configuration
func DefaultStructuredLogger(cfg *logging.LoggerConfig) fiber.Handler {
	logger := logging.NewLogger(cfg)
	return structuredLogger(logger)
}

// structuredLogger is the main middleware function for logging requests and responses
func structuredLogger(logger logging.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip logging for Swagger endpoints
		if strings.Contains(c.Path(), "swagger") {
			return c.Next()
		}

		// Capture the request body
		bodyBytes := c.Request().Body()
		c.Request().SetBody(bodyBytes) // Restore the body for downstream handlers

		// Create a buffer to capture the response body
		blw := &bodyLogWriter{
			Response: c.Response(),
			body:     bytes.NewBufferString(""),
		}
		c.Response().SetBodyStream(blw, -1)

		// Start timer
		start := time.Now()

		// Process the request
		err := c.Next()

		// Logging parameters
		param := struct {
			Latency      time.Duration
			ClientIP     string
			Method       string
			StatusCode   int
			ErrorMessage string
			BodySize     int
			Path         string
			RequestBody  string
			ResponseBody string
		}{
			Latency:      time.Since(start),
			ClientIP:     c.IP(),
			Method:       c.Method(),
			StatusCode:   c.Response().StatusCode(),
			ErrorMessage: "", // Fiber does not have a built-in error collector like Gin
			BodySize:     len(blw.body.Bytes()),
			Path:         c.Path(),
			RequestBody:  string(bodyBytes),
			ResponseBody: blw.body.String(),
		}

		// Add query parameters to the path if present
		if rawQuery := string(c.Request().URI().QueryString()); rawQuery != "" {
			param.Path = param.Path + "?" + rawQuery
		}

		// Log the request and response details
		keys := map[logging.ExtraKey]interface{}{}
		keys[logging.Path] = param.Path
		keys[logging.ClientIp] = param.ClientIP
		keys[logging.Method] = param.Method
		keys[logging.Latency] = param.Latency
		keys[logging.StatusCode] = param.StatusCode
		keys[logging.ErrorMessage] = param.ErrorMessage
		keys[logging.BodySize] = param.BodySize
		keys[logging.RequestBody] = param.RequestBody
		keys[logging.ResponseBody] = param.ResponseBody

		logger.Info(logging.RequestResponse, logging.Api, "", keys)

		return err
	}
}
