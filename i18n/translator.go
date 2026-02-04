package i18n

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
)

//go:embed locales/*.json
var localesFS embed.FS

// Translator manages translations
type Translator struct {
	translations map[string]map[string]interface{}
	mu           sync.RWMutex
	defaultLang  string
}

var (
	instance *Translator
	once     sync.Once
)

// GetTranslator returns the singleton translator instance
func GetTranslator() *Translator {
	once.Do(func() {
		instance = &Translator{
			translations: make(map[string]map[string]interface{}),
			defaultLang:  "en",
		}
		instance.LoadTranslations()
	})
	return instance
}

// LoadTranslations loads all translation files
func (t *Translator) LoadTranslations() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	languages := []string{"en", "fa"}

	for _, lang := range languages {
		data, err := localesFS.ReadFile(fmt.Sprintf("locales/%s.json", lang))
		if err != nil {
			return fmt.Errorf("failed to read locale file %s: %w", lang, err)
		}

		var translations map[string]interface{}
		if err := json.Unmarshal(data, &translations); err != nil {
			return fmt.Errorf("failed to parse locale file %s: %w", lang, err)
		}

		t.translations[lang] = translations
	}

	return nil
}

// GetLangFromContext extracts language from context or fiber context
func (t *Translator) GetLangFromContext(ctx interface{}) string {
	// Try Fiber context first
	if fctx, ok := ctx.(*fiber.Ctx); ok {
		// Check query param
		if lang := fctx.Query("lang"); lang != "" {
			return t.normalizeLang(lang)
		}
		// Check Accept-Language header
		if lang := fctx.Get("Accept-Language"); lang != "" {
			return t.parseAcceptLanguage(lang)
		}
	}

	// Try standard context
	if stdCtx, ok := ctx.(context.Context); ok {
		if lang, ok := stdCtx.Value("lang").(string); ok {
			return t.normalizeLang(lang)
		}
	}

	return t.defaultLang
}

// parseAcceptLanguage parses Accept-Language header
func (t *Translator) parseAcceptLanguage(header string) string {
	if header == "" {
		return t.defaultLang
	}

	// Simple parser: take first language
	parts := strings.Split(header, ",")
	if len(parts) > 0 {
		lang := strings.TrimSpace(strings.Split(parts[0], ";")[0])
		return t.normalizeLang(lang)
	}

	return t.defaultLang
}

// normalizeLang normalizes language code
func (t *Translator) normalizeLang(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))

	// Handle variants
	switch {
	case strings.HasPrefix(lang, "fa"), strings.HasPrefix(lang, "per"):
		return "fa"
	case strings.HasPrefix(lang, "en"):
		return "en"
	default:
		return t.defaultLang
	}
}

// Translate translates a key with optional parameters
func (t *Translator) Translate(ctx interface{}, key string, params ...map[string]interface{}) string {
	lang := t.GetLangFromContext(ctx)
	return t.TranslateWithLang(lang, key, params...)
}

// TranslateWithLang translates a key with specific language
func (t *Translator) TranslateWithLang(lang, key string, params ...map[string]interface{}) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Get translations for language
	translations, ok := t.translations[lang]
	if !ok {
		translations = t.translations[t.defaultLang]
	}

	// Navigate nested keys (e.g., "errors.not_found")
	keys := strings.Split(key, ".")
	var current interface{} = translations

	for _, k := range keys {
		if m, ok := current.(map[string]interface{}); ok {
			current = m[k]
		} else {
			// Key not found, return the key itself
			return key
		}
	}

	// Convert to string
	result, ok := current.(string)
	if !ok {
		return key
	}

	// Replace parameters
	if len(params) > 0 {
		for k, v := range params[0] {
			placeholder := fmt.Sprintf("{{.%s}}", k)
			result = strings.ReplaceAll(result, placeholder, fmt.Sprint(v))
		}
	}

	return result
}

// T is a shorthand for Translate
func T(ctx interface{}, key string, params ...map[string]interface{}) string {
	return GetTranslator().Translate(ctx, key, params...)
}

// TLang is a shorthand for TranslateWithLang
func TLang(lang, key string, params ...map[string]interface{}) string {
	return GetTranslator().TranslateWithLang(lang, key, params...)
}

// SetDefaultLanguage sets the default language
func SetDefaultLanguage(lang string) {
	translator := GetTranslator()
	translator.mu.Lock()
	defer translator.mu.Unlock()
	translator.defaultLang = lang
}
