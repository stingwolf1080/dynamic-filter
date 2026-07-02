package db

import "fmt"

// Error is a custom error type that includes an error code and supports wrapping.
type Error struct {
	Code    int
	Message string
	Err     error
}

func (e Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e Error) Unwrap() error {
	return e.Err
}

var errorRegistry = make(map[int]Error)

// Register manually adds an error to the global cache
func Register(err Error) {
	errorRegistry[err.Code] = err
}

// GetErrorByCode retrieves an error from the cache by its code
func GetErrorByCode(code int) (Error, bool) {
	err, exists := errorRegistry[code]
	return err, exists
}

// GetAllErrors returns all registered errors
func GetAllErrors() map[int]Error {
	return errorRegistry
}

func (e Error) GetCode() int {
	return e.Code
}

func (e Error) GetMessage() string {
	return e.Message
}

// NewError creates a new reusable error and registers it in the cache.
func NewError(code int, message string) Error {
	err := Error{Code: code, Message: message}
	Register(err)
	return err
}

// Wrap wraps an existing error with a specific code and custom message.
func Wrap(err error, code int, message string) Error {
	return Error{Code: code, Message: message, Err: err}
}

var (
	ErrUnauthorized = NewError(1001, "unauthorized")
	ErrBadRequest   = NewError(1002, "bad request")
	ErrForbidden    = NewError(1003, "forbidden")
	ErrNotFound     = NewError(1004, "record not found")
	ErrInternal     = NewError(1005, "internal error")
	ErrDatabase     = NewError(1006, "system database error")
)
