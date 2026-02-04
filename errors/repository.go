package errors

import (
	"errors"
	"fmt"
)

// Standard repository errors
var (
	// ErrNotFound indicates the requested entity was not found
	ErrNotFound = errors.New("entity not found")

	// ErrDuplicate indicates a duplicate entity already exists
	ErrDuplicate = errors.New("entity already exists")

	// ErrConflict indicates a conflict with the current state
	ErrConflict = errors.New("conflict with current state")

	// ErrInvalidInput indicates invalid input data
	ErrInvalidInput = errors.New("invalid input")

	// ErrUnauthorized indicates the user is not authorized
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden indicates the user is forbidden from accessing the resource
	ErrForbidden = errors.New("forbidden")

	// ErrInternal indicates an internal server error
	ErrInternal = errors.New("internal server error")

	// ErrTimeout indicates an operation timeout
	ErrTimeout = errors.New("operation timeout")

	// ErrConnectionFailed indicates a connection failure
	ErrConnectionFailed = errors.New("connection failed")

	// ErrValidation indicates a validation error
	ErrValidation = errors.New("validation error")
)

// RepositoryError wraps repository errors with additional context
type RepositoryError struct {
	Op     string // Operation that failed
	Entity string // Entity type involved
	Err    error  // Underlying error
}

func (e *RepositoryError) Error() string {
	if e.Entity != "" {
		return fmt.Sprintf("%s %s: %v", e.Op, e.Entity, e.Err)
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

func (e *RepositoryError) Unwrap() error {
	return e.Err
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(entity, op string) error {
	return &RepositoryError{
		Op:     op,
		Entity: entity,
		Err:    ErrNotFound,
	}
}

// NewDuplicateError creates a new duplicate error
func NewDuplicateError(entity, op string) error {
	return &RepositoryError{
		Op:     op,
		Entity: entity,
		Err:    ErrDuplicate,
	}
}

// NewConflictError creates a new conflict error
func NewConflictError(entity, op string) error {
	return &RepositoryError{
		Op:     op,
		Entity: entity,
		Err:    ErrConflict,
	}
}

// NewValidationError creates a new validation error with details
func NewValidationError(op, message string) error {
	return &RepositoryError{
		Op:  op,
		Err: fmt.Errorf("%w: %s", ErrValidation, message),
	}
}

// NewInternalError creates a new internal error
func NewInternalError(op string, err error) error {
	return &RepositoryError{
		Op:  op,
		Err: fmt.Errorf("%w: %v", ErrInternal, err),
	}
}

// Error checking functions

// IsNotFound checks if the error is a not found error
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsDuplicate checks if the error is a duplicate error
func IsDuplicate(err error) bool {
	return errors.Is(err, ErrDuplicate)
}

// IsConflict checks if the error is a conflict error
func IsConflict(err error) bool {
	return errors.Is(err, ErrConflict)
}

// IsInvalidInput checks if the error is an invalid input error
func IsInvalidInput(err error) bool {
	return errors.Is(err, ErrInvalidInput)
}

// IsUnauthorized checks if the error is an unauthorized error
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

// IsForbidden checks if the error is a forbidden error
func IsForbidden(err error) bool {
	return errors.Is(err, ErrForbidden)
}

// IsInternal checks if the error is an internal error
func IsInternal(err error) bool {
	return errors.Is(err, ErrInternal)
}

// IsTimeout checks if the error is a timeout error
func IsTimeout(err error) bool {
	return errors.Is(err, ErrTimeout)
}

// IsValidation checks if the error is a validation error
func IsValidation(err error) bool {
	return errors.Is(err, ErrValidation)
}

// ServiceError represents a service-level error with HTTP status mapping
type ServiceError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
	Err        error  `json:"-"`
}

func (e *ServiceError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *ServiceError) Unwrap() error {
	return e.Err
}

// NewServiceError creates a new service error
func NewServiceError(code, message string, statusCode int, err error) *ServiceError {
	return &ServiceError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
	}
}

// Common service errors
func NotFoundServiceError(message string) *ServiceError {
	return &ServiceError{
		Code:       "NOT_FOUND",
		Message:    message,
		StatusCode: 404,
		Err:        ErrNotFound,
	}
}

func BadRequestServiceError(message string) *ServiceError {
	return &ServiceError{
		Code:       "BAD_REQUEST",
		Message:    message,
		StatusCode: 400,
		Err:        ErrInvalidInput,
	}
}

func UnauthorizedServiceError(message string) *ServiceError {
	return &ServiceError{
		Code:       "UNAUTHORIZED",
		Message:    message,
		StatusCode: 401,
		Err:        ErrUnauthorized,
	}
}

func ForbiddenServiceError(message string) *ServiceError {
	return &ServiceError{
		Code:       "FORBIDDEN",
		Message:    message,
		StatusCode: 403,
		Err:        ErrForbidden,
	}
}

func ConflictServiceError(message string) *ServiceError {
	return &ServiceError{
		Code:       "CONFLICT",
		Message:    message,
		StatusCode: 409,
		Err:        ErrConflict,
	}
}

func InternalServiceError(message string, err error) *ServiceError {
	return &ServiceError{
		Code:       "INTERNAL_ERROR",
		Message:    message,
		StatusCode: 500,
		Err:        err,
	}
}

func ValidationServiceError(message string) *ServiceError {
	return &ServiceError{
		Code:       "VALIDATION_ERROR",
		Message:    message,
		StatusCode: 422,
		Err:        ErrValidation,
	}
}
