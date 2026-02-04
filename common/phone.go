package common

import (
	"regexp"
	"strings"
)

// PhoneNumberConfig holds configuration for phone number normalization
type PhoneNumberConfig struct {
	DefaultCountryCode string // e.g., "98" for Iran
	Format             PhoneNumberFormat
}

// PhoneNumberFormat specifies the output format
type PhoneNumberFormat string

const (
	// FormatE164 returns phone in E.164 format: +989123456789
	FormatE164 PhoneNumberFormat = "e164"
	// FormatLocal returns phone in local format: 09123456789
	FormatLocal PhoneNumberFormat = "local"
	// FormatInternational returns phone without + : 989123456789
	FormatInternational PhoneNumberFormat = "international"
)

var (
	// Remove all non-digit characters
	digitRegex = regexp.MustCompile(`\D`)
)

// NormalizePhoneNumber cleans and normalizes a phone number according to config
//
// Examples with DefaultCountryCode="98" (Iran):
//
//	Input: "+989011793041"  → Output: "+989011793041" (E164)
//	Input: "09011793041"    → Output: "+989011793041" (E164)
//	Input: "9011793041"     → Output: "+989011793041" (E164)
//	Input: "+989011793041"  → Output: "09011793041"   (Local)
//	Input: "+989011793041"  → Output: "989011793041"  (International)
func NormalizePhoneNumber(phone string, config PhoneNumberConfig) string {
	if phone == "" {
		return ""
	}

	// Set default format if not specified
	if config.Format == "" {
		config.Format = FormatE164
	}

	// Extract only digits
	digits := digitRegex.ReplaceAllString(phone, "")

	// Remove leading zeros
	digits = strings.TrimLeft(digits, "0")

	// If no country code, add default
	if config.DefaultCountryCode != "" && !strings.HasPrefix(digits, config.DefaultCountryCode) {
		digits = config.DefaultCountryCode + digits
	}

	// Format according to requested output
	switch config.Format {
	case FormatE164:
		return "+" + digits
	case FormatLocal:
		// Remove country code and add leading zero for Iran
		if config.DefaultCountryCode != "" && strings.HasPrefix(digits, config.DefaultCountryCode) {
			digits = strings.TrimPrefix(digits, config.DefaultCountryCode)
		}
		return "0" + digits
	case FormatInternational:
		return digits
	default:
		return "+" + digits
	}
}

// NormalizeIranPhone is a convenience function for Iranian phone numbers
// Returns phone in E.164 format (+989123456789)
//
// Accepts:
//   - +989011793041 → +989011793041
//   - 09011793041   → +989011793041
//   - 9011793041    → +989011793041
//   - 00989011793041 → +989011793041
func NormalizeIranPhone(phone string) string {
	return NormalizePhoneNumber(phone, PhoneNumberConfig{
		DefaultCountryCode: "98",
		Format:             FormatE164,
	})
}

// NormalizeIranPhoneLocal returns Iranian phone in local format (09123456789)
//
// Accepts:
//   - +989011793041 → 09011793041
//   - 09011793041   → 09011793041
//   - 9011793041    → 09011793041
func NormalizeIranPhoneLocal(phone string) string {
	return NormalizePhoneNumber(phone, PhoneNumberConfig{
		DefaultCountryCode: "98",
		Format:             FormatLocal,
	})
}

// ValidateIranMobileNumber checks if a number is a valid Iranian mobile number
// Valid patterns: 09XXXXXXXXX (11 digits starting with 09)
func ValidateIranMobileNumber(phone string) bool {
	normalized := NormalizeIranPhoneLocal(phone)

	// Must be 11 digits starting with 09
	if len(normalized) != 11 {
		return false
	}

	if !strings.HasPrefix(normalized, "09") {
		return false
	}

	// Check if all characters are digits
	matched, _ := regexp.MatchString(`^09\d{9}$`, normalized)
	return matched
}
