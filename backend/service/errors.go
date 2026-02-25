package service

import "errors"

var (
	ErrNotFound           = errors.New("not found")
	ErrExists             = errors.New("exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type ParamsError struct {
	Message string
}

func (e ParamsError) Error() string {
	return e.Message
}

// Error type wrappers
type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

type NotFoundError struct {
	Message string
}

func (e NotFoundError) Error() string {
	return e.Message
}

type ConflictError struct {
	Message string
}

func (e ConflictError) Error() string {
	return e.Message
}

// Helper functions to check error types
func IsValidationError(err error) bool {
	var ve ValidationError
	var pe ParamsError
	return errors.As(err, &ve) || errors.As(err, &pe)
}

func IsNotFoundError(err error) bool {
	var nfe NotFoundError
	return errors.As(err, &nfe) || errors.Is(err, ErrNotFound)
}

func IsConflictError(err error) bool {
	var ce ConflictError
	return errors.As(err, &ce) || errors.Is(err, ErrExists)
}
