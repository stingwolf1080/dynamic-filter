package resolver

import (
	"context"
	"fmt"
)

// DependencyResolver resolves ref-tagged dependencies through its injected registry.
// It validates IDs and applies successful resolver data to their declared targets.
type DependencyResolver struct{ registry *Registry }

func NewDependencyResolver(registry *Registry) *DependencyResolver {
	return &DependencyResolver{registry: registry}
}

// Resolve checks every ref-tagged field on model and applies successful results
// to the corresponding into target on that same model pointer.
func (v *DependencyResolver) Resolve(ctx context.Context, model any, prefix string) error {
	if v == nil || v.registry == nil {
		return fmt.Errorf("reference validator registry is nil")
	}
	refs, err := ExtractReferences(model)
	if err != nil {
		return err
	}
	var dependency *Dependency
	for _, ref := range refs {
		if ref.Target != "" && dependency == nil {
			dependency, err = NewDependency(model)
			if err != nil {
				return fmt.Errorf("prepare resolved reference %s (%s): %w", ref.Field, ref.Name, err)
			}
		}
		if ref.Target != "" {
			if err := dependency.ValidateTarget(ref); err != nil {
				return fmt.Errorf("prepare resolved reference %s (%s): %w", ref.Field, ref.Name, err)
			}
		}
		if ref.ID == "" {
			if ref.Required {
				return fmt.Errorf("%w: %s", ErrReferenceRequired, ref.Field)
			}
			continue
		}
		message := v.registry.Resolve(ctx, ref.Name, model, prefix)
		if message.Code >= 400 || message.MessageErr != nil || message.Status == "error" {
			if message.MessageErr != nil {
				return fmt.Errorf("validate reference %s (%s): %w", ref.Field, ref.Name, message.MessageErr)
			}
			return fmt.Errorf("validate reference %s (%s): %s", ref.Field, ref.Name, message.Message)
		}
		if ref.Target != "" {
			if err := dependency.Apply(ref, message.Data); err != nil {
				return fmt.Errorf("apply resolved reference %s (%s): %w", ref.Field, ref.Name, err)
			}
		}
	}
	return nil
}

// Validate is kept as a compatibility alias for Resolve.
func (v *DependencyResolver) Validate(ctx context.Context, model any, prefix string) error {
	return v.Resolve(ctx, model, prefix)
}
