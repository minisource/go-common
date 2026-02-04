package common

import (
	"regexp"
	"strings"
	"unicode"
)

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func HasUpper(s string) bool {
	for _, r := range s {
		if unicode.IsUpper(r) && unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

func HasLower(s string) bool {
	for _, r := range s {
		if unicode.IsLower(r) && unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

func HasLetter(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

func HasDigits(s string) bool {
	for _, r := range s {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

// To snake case : CountryId -> country_id
func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
