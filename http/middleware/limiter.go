package middleware

import (
	"net/http"
	"net/url"

	"github.com/gofiber/fiber/v2"
	"github.com/didip/tollbooth/v7"
	helper "github.com/minisource/go-common/http/helper"
)

func LimitByRequest() fiber.Handler {
	lmt := tollbooth.NewLimiter(1, nil)
	return func(c *fiber.Ctx) error {
		// Convert fiber.Ctx to http.ResponseWriter and *http.Request
		writer := &fiberResponseWriter{c}
		req := convertFiberRequestToHTTP(c)

		// Call tollbooth.LimitByRequest with the converted request
		err := tollbooth.LimitByRequest(lmt, writer, req)
		if err != nil {
			httpResponse := helper.GenerateBaseResponseWithError(nil, false, helper.LimiterError, err)
			return c.Status(fiber.StatusTooManyRequests).JSON(httpResponse)
		}
		return c.Next()
	}
}

// fiberResponseWriter is a wrapper to adapt fiber.Ctx to http.ResponseWriter
type fiberResponseWriter struct {
	ctx *fiber.Ctx
}

func (w *fiberResponseWriter) Header() http.Header {
	return make(http.Header)
}

func (w *fiberResponseWriter) Write(data []byte) (int, error) {
	return w.ctx.Write(data)
}

func (w *fiberResponseWriter) WriteHeader(statusCode int) {
	w.ctx.Status(statusCode)
}

// convertFiberRequestToHTTP converts a fiber.Ctx request to an *http.Request
func convertFiberRequestToHTTP(c *fiber.Ctx) *http.Request {
	// Create a new URL object
	u := &url.URL{
		Scheme:   "http", // or "https" if applicable
		Host:     string(c.Request().Host()),
		Path:     string(c.Request().URI().Path()),
		RawQuery: string(c.Request().URI().QueryString()),
	}

	// Create a new http.Request
	req := &http.Request{
		Method: string(c.Request().Header.Method()),
		URL:    u,
		Header: make(http.Header),
	}

	// Copy headers from fasthttp to http.Request
	c.Request().Header.VisitAll(func(key, value []byte) {
		req.Header.Set(string(key), string(value))
	})

	return req
}