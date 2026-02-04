package response

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/minisource/go-common/i18n"
)

// I18nBuilder extends Builder with i18n support
type I18nBuilder struct {
	*Builder
	translator *i18n.Translator
	lang       string
}

// NewI18n creates a response builder with i18n support
func NewI18n(translator *i18n.Translator, lang string) *I18nBuilder {
	return &I18nBuilder{
		Builder:    New(),
		translator: translator,
		lang:       lang,
	}
}

// FromContext creates an i18n builder extracting language from Fiber context
func FromContext(c *fiber.Ctx, translator *i18n.Translator) *I18nBuilder {
	lang := extractLanguage(c)
	return NewI18n(translator, lang)
}

// extractLanguage gets the preferred language from request
func extractLanguage(c *fiber.Ctx) string {
	// Check query parameter
	if lang := c.Query("lang"); lang != "" {
		return lang
	}

	// Check header
	if lang := c.Get("Accept-Language"); lang != "" {
		// Simple parsing - take first language
		if len(lang) >= 2 {
			return lang[:2]
		}
	}

	// Check custom header
	if lang := c.Get("X-Language"); lang != "" {
		return lang
	}

	return "en" // Default
}

// Error sets a translated error
func (b *I18nBuilder) Error(code string, params ...map[string]interface{}) *I18nBuilder {
	var p map[string]interface{}
	if len(params) > 0 {
		p = params[0]
	}

	message := b.translator.TranslateWithLang(b.lang, code, p)
	b.Builder.Error(code, message)
	return b
}

// ErrorWithKey sets error with separate message key
func (b *I18nBuilder) ErrorWithKey(code, messageKey string, params ...map[string]interface{}) *I18nBuilder {
	var p map[string]interface{}
	if len(params) > 0 {
		p = params[0]
	}

	message := b.translator.TranslateWithLang(b.lang, messageKey, p)
	b.Builder.Error(code, message)
	return b
}

// Data sets the response data
func (b *I18nBuilder) Data(data interface{}) *I18nBuilder {
	b.Builder.Data(data)
	return b
}

// Status sets HTTP status
func (b *I18nBuilder) Status(code int) *I18nBuilder {
	b.Builder.Status(code)
	return b
}

// WithPagination adds pagination
func (b *I18nBuilder) WithPagination(p *Pagination) *I18nBuilder {
	b.Builder.WithPagination(p)
	return b
}

// Send sends the response
func (b *I18nBuilder) Send(c *fiber.Ctx) error {
	return b.Builder.Send(c)
}

// ============================================
// I18n convenience functions
// ============================================

// I18nResponse provides i18n response helpers bound to a translator
type I18nResponse struct {
	translator *i18n.Translator
}

// NewI18nResponse creates a new i18n response helper
func NewI18nResponse(translator *i18n.Translator) *I18nResponse {
	return &I18nResponse{translator: translator}
}

// OK sends a translated success message
func (r *I18nResponse) OK(c *fiber.Ctx, data interface{}) error {
	return FromContext(c, r.translator).Data(data).Send(c)
}

// Created sends a translated created response
func (r *I18nResponse) Created(c *fiber.Ctx, data interface{}) error {
	return FromContext(c, r.translator).Status(http.StatusCreated).Data(data).Send(c)
}

// BadRequest sends a translated bad request error
func (r *I18nResponse) BadRequest(c *fiber.Ctx, code string, params ...map[string]interface{}) error {
	return FromContext(c, r.translator).Status(http.StatusBadRequest).Error(code, params...).Send(c)
}

// Unauthorized sends a translated unauthorized error
func (r *I18nResponse) Unauthorized(c *fiber.Ctx, code string, params ...map[string]interface{}) error {
	return FromContext(c, r.translator).Status(http.StatusUnauthorized).Error(code, params...).Send(c)
}

// Forbidden sends a translated forbidden error
func (r *I18nResponse) Forbidden(c *fiber.Ctx, code string, params ...map[string]interface{}) error {
	return FromContext(c, r.translator).Status(http.StatusForbidden).Error(code, params...).Send(c)
}

// NotFound sends a translated not found error
func (r *I18nResponse) NotFound(c *fiber.Ctx, code string, params ...map[string]interface{}) error {
	return FromContext(c, r.translator).Status(http.StatusNotFound).Error(code, params...).Send(c)
}

// Conflict sends a translated conflict error
func (r *I18nResponse) Conflict(c *fiber.Ctx, code string, params ...map[string]interface{}) error {
	return FromContext(c, r.translator).Status(http.StatusConflict).Error(code, params...).Send(c)
}

// InternalError sends a translated internal error
func (r *I18nResponse) InternalError(c *fiber.Ctx, code string, params ...map[string]interface{}) error {
	return FromContext(c, r.translator).Status(http.StatusInternalServerError).Error(code, params...).Send(c)
}

// ValidationError sends translated validation errors
func (r *I18nResponse) ValidationError(c *fiber.Ctx, fieldErrors []ValidationError) error {
	lang := extractLanguage(c)

	// Translate each validation error message
	translated := make([]ValidationError, len(fieldErrors))
	for i, fe := range fieldErrors {
		translated[i] = ValidationError{
			Field:   fe.Field,
			Code:    fe.Code,
			Message: r.translator.TranslateWithLang(lang, fe.Code, map[string]interface{}{"field": fe.Field}),
		}
	}

	return New().Status(http.StatusUnprocessableEntity).ValidationErrors(translated).Send(c)
}
