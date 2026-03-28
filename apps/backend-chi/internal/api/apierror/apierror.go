package apierror

import (
	"errors"
	"fmt"
)

type APIError struct {
	Status  int
	Message string
	Err     error
}

func (e *APIError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *APIError) Unwrap() error {
	return e.Err
}

func NewApiError(status int, message string, err error) *APIError {
	if err == nil {
		err = errors.New(message)
	}
	return &APIError{
		Status:  status,
		Message: message,
		Err:     err,
	}
}
