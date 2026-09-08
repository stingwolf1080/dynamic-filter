// Package resolver validates references to entities owned by local or remote services.
package resolver

import "context"

import "github.com/stingwolf1080/dynamic-filter/pkg/util/types"

// ResolveFunc checks a reference in the current repository prefix. On success,
// Message.Data is applied to the model by DependencyResolver before persistence.
type ResolveFunc func(ctx context.Context, data any, prefix string) types.Message
