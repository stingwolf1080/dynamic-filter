package hooks

import (
	"context"

	"github.com/stingwolf1080/dynamic-filter/pkg/resolver"
	"github.com/stingwolf1080/dynamic-filter/pkg/util/types"
)

// RegisterDependencyResolver adds dependency resolution to BeforeInsert for T.
// The resolver and all of its repositories or remote clients are supplied by the caller.
func RegisterDependencyResolver[T any](dependencyResolver *resolver.DependencyResolver) {
	Register[T](BeforeInsert, func(ctx context.Context, data any, prefix string) types.Message {
		if err := dependencyResolver.Resolve(ctx, data, prefix); err != nil {
			return types.Message{Status: "error", Code: 400, Message: err.Error(), MessageErr: err}
		}
		return types.Message{}
	})
}
