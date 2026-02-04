package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizePhoneNumber_E164(t *testing.T) {
	config := PhoneNumberConfig{
		DefaultCountryCode: "98",
		Format:             FormatE164,
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"With plus and country code", "+989011793041", "+989011793041"},
		{"Local format with zero", "09011793041", "+989011793041"},
		{"Without leading zero", "9011793041", "+989011793041"},
		{"With 00 prefix", "00989011793041", "+989011793041"},
		{"With spaces", "+98 901 179 3041", "+989011793041"},
		{"With dashes", "+98-901-179-3041", "+989011793041"},
		{"With parentheses", "+98(901)179-3041", "+989011793041"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizePhoneNumber(tt.input, config)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizePhoneNumber_Local(t *testing.T) {
	config := PhoneNumberConfig{
		DefaultCountryCode: "98",
		Format:             FormatLocal,
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"From E164", "+989011793041", "09011793041"},
		{"Already local", "09011793041", "09011793041"},
		{"Without leading zero", "9011793041", "09011793041"},
		{"With 00 prefix", "00989011793041", "09011793041"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizePhoneNumber(tt.input, config)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizePhoneNumber_International(t *testing.T) {
	config := PhoneNumberConfig{
		DefaultCountryCode: "98",
		Format:             FormatInternational,
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"From E164", "+989011793041", "989011793041"},
		{"From local", "09011793041", "989011793041"},
		{"Already international", "989011793041", "989011793041"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizePhoneNumber(tt.input, config)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizeIranPhone(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Plus format", "+989011793041", "+989011793041"},
		{"Local format", "09011793041", "+989011793041"},
		{"No leading zero", "9011793041", "+989011793041"},
		{"With 00 prefix", "00989011793041", "+989011793041"},
		{"With spaces and dashes", "0901-179-3041", "+989011793041"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeIranPhone(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizeIranPhoneLocal(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"From E164", "+989011793041", "09011793041"},
		{"Already local", "09011793041", "09011793041"},
		{"No leading zero", "9011793041", "09011793041"},
		{"With formatting", "+98-901-179-3041", "09011793041"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeIranPhoneLocal(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateIranMobileNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Valid local", "09011793041", true},
		{"Valid E164", "+989011793041", true},
		{"Valid no zero", "9011793041", true},
		{"Invalid too short", "0901179304", false},
		{"Invalid too long", "090117930411", false},
		{"Invalid not starting with 09", "08011793041", false},
		{"Invalid landline", "02188776655", false},
		{"Empty string", "", false},
		{"Non-numeric", "abc123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateIranMobileNumber(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizePhoneNumber_EmptyInput(t *testing.T) {
	config := PhoneNumberConfig{
		DefaultCountryCode: "98",
		Format:             FormatE164,
	}

	result := NormalizePhoneNumber("", config)
	assert.Equal(t, "", result)
}

func TestNormalizePhoneNumber_OtherCountries(t *testing.T) {
	tests := []struct {
		name        string
		countryCode string
		input       string
		expected    string
	}{
		{"US number", "1", "+15551234567", "+15551234567"},
		{"US local to E164", "1", "5551234567", "+15551234567"},
		{"UK number", "44", "+447911123456", "+447911123456"},
		{"UK local", "44", "07911123456", "+447911123456"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := PhoneNumberConfig{
				DefaultCountryCode: tt.countryCode,
				Format:             FormatE164,
			}
			result := NormalizePhoneNumber(tt.input, config)
			assert.Equal(t, tt.expected, result)
		})
	}
}
