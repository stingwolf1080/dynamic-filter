package resolver

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/stingwolf1080/dynamic-filter/pkg/util/types"
)

// Registry maps a transport-neutral reference name (for example "cruise") to a Resolver.
type Registry struct {
	mu        sync.RWMutex
	resolvers map[string]ResolveFunc
}

func NewRegistry() *Registry { return &Registry{resolvers: make(map[string]ResolveFunc)} }

// Register associates a reference name with the function used to resolve it.
func (r *Registry) Register(name string, resolver ResolveFunc) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("resolver name is empty")
	}
	if resolver == nil {
		return fmt.Errorf("resolver %q is nil", name)
	}
	r.mu.Lock()
	r.resolvers[name] = resolver
	r.mu.Unlock()
	return nil
}

// Resolve runs the registered function and returns its Message. Missing
// resolver names are represented as an error Message rather than an error
// return, keeping the public resolution flow on types.Message.
func (r *Registry) Resolve(ctx context.Context, name string, data any, prefix string) types.Message {
	r.mu.RLock()
	resolver, ok := r.resolvers[name]
	r.mu.RUnlock()
	if !ok {
		err := fmt.Errorf("%w: %s", ErrResolverNotFound, name)
		return types.Message{Status: "error", Code: 400, Message: err.Error(), MessageErr: err}
	}
	return resolver(ctx, data, prefix)
}
