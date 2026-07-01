package db

// Error is a string wrapper that implements the error interface.
// This allows defining errors as constants.
type Error string

func (e Error) Error() string {
	return string(e)
}

const (
	ErrNotFound = Error("record not found")
)
