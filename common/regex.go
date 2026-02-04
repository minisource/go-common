package common

import "regexp"

func IsIranianNumber(phoneNumber string) bool {
	// Regex pattern for Iranian phone numbers: starts with +98 or 09, followed by 9 digits
	// Example: +989123456789 or 09123456789
	iranianNumberPattern := `^(?:\+98|0)9\d{9}$`
	matched, _ := regexp.MatchString(iranianNumberPattern, phoneNumber)
	return matched
}