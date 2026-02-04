package common

import (
	"os"
	"strings"
)

// Environment represents the application environment
type Environment string

const (
	EnvDevelopment Environment = "development"
	EnvStaging     Environment = "staging"
	EnvProduction  Environment = "production"
	EnvTest        Environment = "test"
)

// GetEnvironment returns current environment
func GetEnvironment() Environment {
	env := strings.ToLower(os.Getenv("APP_ENV"))
	if env == "" {
		env = strings.ToLower(os.Getenv("ENV"))
	}
	if env == "" {
		env = strings.ToLower(os.Getenv("ENVIRONMENT"))
	}

	switch env {
	case "development", "dev":
		return EnvDevelopment
	case "staging", "stage":
		return EnvStaging
	case "production", "prod":
		return EnvProduction
	case "test", "testing":
		return EnvTest
	default:
		return EnvDevelopment // Default to development for safety
	}
}

// IsDevelopment checks if environment is development
func IsDevelopment() bool {
	return GetEnvironment() == EnvDevelopment
}

// IsProduction checks if environment is production
func IsProduction() bool {
	return GetEnvironment() == EnvProduction
}

// IsStaging checks if environment is staging
func IsStaging() bool {
	return GetEnvironment() == EnvStaging
}

// IsTest checks if environment is test
func IsTest() bool {
	return GetEnvironment() == EnvTest
}

// ShouldShowDetailedErrors returns true if detailed errors should be shown
func ShouldShowDetailedErrors() bool {
	return IsDevelopment() || IsTest()
}
