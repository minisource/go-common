package grpcclient

import "fmt"

// ServiceUnavailableError represents an error when a service is unavailable
type ServiceUnavailableError struct {
	ServiceName string
	Err         error
}

func (e *ServiceUnavailableError) Error() string {
	return fmt.Sprintf("service '%s' is unavailable: %v", e.ServiceName, e.Err)
}

func (e *ServiceUnavailableError) Unwrap() error {
	return e.Err
}

// NewServiceUnavailableError creates a new ServiceUnavailableError
func NewServiceUnavailableError(serviceName string, err error) error {
	return &ServiceUnavailableError{
		ServiceName: serviceName,
		Err:         err,
	}
}
