package middleware

import (
	"github.com/gofiber/fiber/v2"
	helper "github.com/minisource/go-common/http/helper"
)

func ErrorHandler(c *fiber.Ctx, err error) error {
	if err != nil {
		httpResponse := helper.GenerateBaseResponseWithError(nil, false, helper.CustomRecovery, err)
		return c.Status(fiber.StatusInternalServerError).JSON(httpResponse)
	}
	httpResponse := helper.GenerateBaseResponseWithAnyError(nil, false, helper.CustomRecovery, "Unknown error")
	return c.Status(fiber.StatusInternalServerError).JSON(httpResponse)
}