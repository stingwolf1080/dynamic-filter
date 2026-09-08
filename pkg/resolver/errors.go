package resolver

import "errors"

var (
	// ErrResolverNotFound means a ref tag has no registered resolver.
	ErrResolverNotFound = errors.New("reference resolver not found")
	// ErrReferenceRequired means a required ref field was empty.
	ErrReferenceRequired = errors.New("required reference is empty")
	// ErrReferenceNotFound means the referenced entity does not exist.
	ErrReferenceNotFound = errors.New("referenced entity not found")
	// ErrInvalidModel means references cannot be extracted from the supplied value.
	ErrInvalidModel = errors.New("reference model must be a struct or pointer to a struct")
	// ErrInvalidResolvedData means resolver data cannot be applied to its target field.
	ErrInvalidResolvedData = errors.New("resolved data cannot be applied to the target field")
	// ErrReferenceTargetNotFound means an into target is not a settable model field.
	ErrReferenceTargetNotFound = errors.New("reference target field not found")
)
