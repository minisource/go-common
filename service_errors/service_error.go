package service_errors

import (
	"fmt"
)

type ServiceError struct {
	EndUserMessage   string                 `json:"endUserMessage"`
	TechnicalMessage string                 `json:"technicalMessage,omitempty"`
	Err              error                  `json:"-"`
	Code             string                 `json:"code,omitempty"`
	Details          map[string]interface{} `json:"details,omitempty"`
	Stack            string                 `json:"stack,omitempty"` // Only in development
}

func (s *ServiceError) Error() string {
	return s.EndUserMessage
}

// NewServiceError creates a new service error
func NewServiceError(code, endUserMsg, technicalMsg string) *ServiceError {
	return &ServiceError{
		Code:             code,
		EndUserMessage:   endUserMsg,
		TechnicalMessage: technicalMsg,
	}
}

// WithError adds underlying error
func (s *ServiceError) WithError(err error) *ServiceError {
	s.Err = err
	if s.TechnicalMessage == "" && err != nil {
		s.TechnicalMessage = err.Error()
	}
	return s
}

// WithDetails adds additional details
func (s *ServiceError) WithDetails(details map[string]interface{}) *ServiceError {
	s.Details = details
	return s
}

// WithStack adds stack trace (for development)
func (s *ServiceError) WithStack(stack string) *ServiceError {
	s.Stack = stack
	return s
}

// GetDetails returns error details for API response
// isDevelopment controls whether to include technical details
func (s *ServiceError) GetDetails(isDevelopment bool) map[string]interface{} {
	result := map[string]interface{}{
		"message": s.EndUserMessage,
		"code":    s.Code,
	}

	if isDevelopment {
		if s.TechnicalMessage != "" {
			result["technical_message"] = s.TechnicalMessage
		}
		if s.Err != nil {
			result["error"] = s.Err.Error()
		}
		if s.Details != nil {
			result["details"] = s.Details
		}
		if s.Stack != "" {
			result["stack"] = s.Stack
		}
	}

	return result
}

// Common error constructors
func NewUnexpectedError() *ServiceError {
	return NewServiceError(UnExpectedError, "An unexpected error occurred", "")
}

func NewRecordNotFoundError() *ServiceError {
	return NewServiceError(RecordNotFound, "Record not found", "")
}

func NewPermissionDeniedError() *ServiceError {
	return NewServiceError(PermissionDenied, "Permission denied", "")
}

func NewValidationError(msg string) *ServiceError {
	return NewServiceError(ValidationError, msg, "")
}

func NewTokenError(code, msg string) *ServiceError {
	return NewServiceError(code, msg, "")
}

// Error with formatted message
func Errorf(code, format string, args ...interface{}) *ServiceError {
	return NewServiceError(code, fmt.Sprintf(format, args...), "")
}
