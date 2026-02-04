package helper

import (
    "context"
    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
)

// Create handles generic creation endpoints
func Create[Ti any, To any](c *fiber.Ctx, caller func(ctx context.Context, req *Ti) (*To, error)) error {
    req := new(Ti)
    if err := c.BodyParser(req); err != nil {
        return c.Status(fiber.StatusBadRequest).
            JSON(GenerateBaseResponseWithValidationError(nil, false, ValidationError, err))
    }

    res, err := caller(c.Context(), req)
    if err != nil {
        return c.Status(TranslateErrorToStatusCode(err)).
            JSON(GenerateBaseResponseWithError(nil, false, InternalError, err))
    }
    
    return c.Status(fiber.StatusCreated).
        JSON(GenerateBaseResponse(res, true, 0))
}

// Update handles generic update endpoints
func Update[Ti any, To any](c *fiber.Ctx, caller func(ctx context.Context, id uuid.UUID, req *Ti) (*To, error)) error {
    id, err := uuid.Parse(c.Params("id"))
    if err != nil {
        return c.Status(fiber.StatusBadRequest).
            JSON(GenerateBaseResponse(nil, false, ValidationError))
    }

    req := new(Ti)
    if err := c.BodyParser(req); err != nil {
        return c.Status(fiber.StatusBadRequest).
            JSON(GenerateBaseResponseWithValidationError(nil, false, ValidationError, err))
    }

    res, err := caller(c.Context(), id, req)
    if err != nil {
        return c.Status(TranslateErrorToStatusCode(err)).
            JSON(GenerateBaseResponseWithError(nil, false, InternalError, err))
    }

    return c.Status(fiber.StatusOK).
        JSON(GenerateBaseResponse(res, true, 0))
}

// Delete handles generic deletion endpoints
func Delete(c *fiber.Ctx, caller func(ctx context.Context, id uuid.UUID) error) error {
    id, err := uuid.Parse(c.Params("id"))
    if err != nil {
        return c.Status(fiber.StatusBadRequest).
            JSON(GenerateBaseResponse(nil, false, ValidationError))
    }

    if err := caller(c.Context(), id); err != nil {
        return c.Status(TranslateErrorToStatusCode(err)).
            JSON(GenerateBaseResponseWithError(nil, false, InternalError, err))
    }

    return c.Status(fiber.StatusOK).
        JSON(GenerateBaseResponse(nil, true, 0))
}

// GetByID handles generic get-by-id endpoints
func GetByID[To any](c *fiber.Ctx, caller func(ctx context.Context, id uuid.UUID) (*To, error)) error {
    id, err := uuid.Parse(c.Params("id"))
    if err != nil {
        return c.Status(fiber.StatusBadRequest).
            JSON(GenerateBaseResponse(nil, false, ValidationError))
    }

    res, err := caller(c.Context(), id)
    if err != nil {
        return c.Status(TranslateErrorToStatusCode(err)).
            JSON(GenerateBaseResponseWithError(nil, false, InternalError, err))
    }

    return c.Status(fiber.StatusOK).
        JSON(GenerateBaseResponse(res, true, 0))
}

// GetByFilter handles generic filter-based get endpoints
func GetByFilter[Ti any, To any](c *fiber.Ctx, caller func(ctx context.Context, req *Ti) (*To, error)) error {
    req := new(Ti)
    if err := c.BodyParser(req); err != nil {
        return c.Status(fiber.StatusBadRequest).
            JSON(GenerateBaseResponseWithValidationError(nil, false, ValidationError, err))
    }

    res, err := caller(c.Context(), req)
    if err != nil {
        return c.Status(TranslateErrorToStatusCode(err)).
            JSON(GenerateBaseResponseWithError(nil, false, InternalError, err))
    }

    return c.Status(fiber.StatusOK).
        JSON(GenerateBaseResponse(res, true, 0))
}