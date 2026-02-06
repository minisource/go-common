package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewNotFoundError(t *testing.T) {
	err := NewNotFoundError("user", "find")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user")
	assert.Contains(t, err.Error(), "find")
	assert.True(t, IsNotFound(err))
}

func TestNewDuplicateError(t *testing.T) {
	err := NewDuplicateError("email", "create")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email")
	assert.True(t, IsDuplicate(err))
}

func TestNewConflictError(t *testing.T) {
	err := NewConflictError("order", "update")

	assert.Error(t, err)
	assert.True(t, IsConflict(err))
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("create", "email is required")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email is required")
	assert.True(t, IsValidation(err))
}

func TestNewInternalError(t *testing.T) {
	innerErr := errors.New("database connection failed")
	err := NewInternalError("query", innerErr)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection failed")
	assert.True(t, IsInternal(err))
}

func TestIsNotFound(t *testing.T) {
	notFoundErr := NewNotFoundError("user", "find")
	otherErr := NewDuplicateError("user", "create")

	assert.True(t, IsNotFound(notFoundErr))
	assert.False(t, IsNotFound(otherErr))
}

func TestIsUnauthorized(t *testing.T) {
	unauthorizedErr := &RepositoryError{Op: "test", Err: ErrUnauthorized}
	otherErr := NewNotFoundError("user", "find")

	assert.True(t, IsUnauthorized(unauthorizedErr))
	assert.False(t, IsUnauthorized(otherErr))
}

func TestIsForbidden(t *testing.T) {
	forbiddenErr := &RepositoryError{Op: "test", Err: ErrForbidden}
	otherErr := NewNotFoundError("user", "find")

	assert.True(t, IsForbidden(forbiddenErr))
	assert.False(t, IsForbidden(otherErr))
}

func TestRepositoryErrorUnwrap(t *testing.T) {
	err := NewNotFoundError("user", "find")

	var repoErr *RepositoryError
	assert.True(t, errors.As(err, &repoErr))
	assert.Equal(t, ErrNotFound, errors.Unwrap(err))
}

func TestStandardErrors(t *testing.T) {
	// Verify standard errors are defined
	assert.NotNil(t, ErrNotFound)
	assert.NotNil(t, ErrDuplicate)
	assert.NotNil(t, ErrConflict)
	assert.NotNil(t, ErrInvalidInput)
	assert.NotNil(t, ErrUnauthorized)
	assert.NotNil(t, ErrForbidden)
	assert.NotNil(t, ErrInternal)
	assert.NotNil(t, ErrTimeout)
	assert.NotNil(t, ErrConnectionFailed)
	assert.NotNil(t, ErrValidation)
}

func TestRepositoryErrorMessage(t *testing.T) {
	err := &RepositoryError{
		Op:     "create",
		Entity: "user",
		Err:    ErrDuplicate,
	}

	msg := err.Error()
	assert.Contains(t, msg, "create")
	assert.Contains(t, msg, "user")
	assert.Contains(t, msg, "already exists")
}

func TestRepositoryErrorMessageWithoutEntity(t *testing.T) {
	err := &RepositoryError{
		Op:  "connect",
		Err: ErrConnectionFailed,
	}

	msg := err.Error()
	assert.Contains(t, msg, "connect")
	assert.NotContains(t, msg, "  ") // No double spaces from empty entity
}
