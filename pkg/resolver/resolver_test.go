package resolver

import (
	"context"
	"errors"
	"testing"

	"github.com/stingwolf1080/dynamic-filter/pkg/util/types"
)

type testModel struct {
	RouteID  string `ref:"route,required"`
	CruiseID string `ref:"cruise"`
}

type resolvedTestModel struct {
	RouteID string            `json:"route_id" ref:"route,required,into=Route"`
	Route   resolvedTestRoute `json:"route,omitempty"`
}

type resolvedTestRoute struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type benchmarkDependencyModel struct {
	RouteID  string            `ref:"route,required,into=Route"`
	Route    resolvedTestRoute `json:"route,omitempty"`
	CruiseID string            `ref:"cruise,required,into=Cruise"`
	Cruise   resolvedTestRoute `json:"cruise,omitempty"`
	SiteID   string            `ref:"site,required,into=Site"`
	Site     resolvedTestRoute `json:"site,omitempty"`
}

func TestExtractReferences(t *testing.T) {
	refs, err := ExtractReferences(&testModel{RouteID: " route-1 ", CruiseID: "cruise-1"})
	if err != nil {
		t.Fatalf("ExtractReferences() error = %v", err)
	}
	if len(refs) != 2 {
		t.Fatalf("got %d references, want 2", len(refs))
	}
	if refs[0] != (Reference{Field: "testModel.RouteID", Name: "route", ID: "route-1", Required: true}) {
		t.Fatalf("first reference = %#v", refs[0])
	}
}

func TestRegistryResolveRunsRegisteredFunction(t *testing.T) {
	registry := NewRegistry()
	model := testModel{RouteID: "route-1"}
	var gotData any
	var gotPrefix string
	if err := registry.Register("route", func(_ context.Context, data any, prefix string) types.Message {
		gotData = data
		gotPrefix = prefix
		return types.Message{Status: "success", Code: 200, Data: "route"}
	}); err != nil {
		t.Fatal(err)
	}

	message := registry.Resolve(context.Background(), "route", model, "tenant-a")
	if gotData != model || gotPrefix != "tenant-a" {
		t.Fatalf("function received data=%#v prefix=%q", gotData, gotPrefix)
	}
	if message.Code != 200 || message.Data != "route" {
		t.Fatalf("Resolve() = %#v", message)
	}
}

func TestDependencyResolver(t *testing.T) {
	registry := NewRegistry()
	for _, name := range []string{"route", "cruise"} {
		if err := registry.Register(name, func(context.Context, any, string) types.Message {
			return types.Message{Status: "success", Code: 200}
		}); err != nil {
			t.Fatal(err)
		}
	}
	if err := NewDependencyResolver(registry).Resolve(context.Background(), testModel{RouteID: "route-1", CruiseID: "cruise-1"}, "tenant-a"); err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
}

func TestDependencyResolverAppliesResolverDataToModel(t *testing.T) {
	registry := NewRegistry()
	if err := registry.Register("route", func(context.Context, any, string) types.Message {
		return types.Message{Status: "success", Code: 200, Data: map[string]any{
			"id":   "route-1",
			"name": "Ha Long route",
		}}
	}); err != nil {
		t.Fatal(err)
	}

	model := &resolvedTestModel{RouteID: "route-1"}
	if err := NewDependencyResolver(registry).Resolve(context.Background(), model, "tenant-a"); err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if model.Route.ID != "route-1" || model.Route.Name != "Ha Long route" {
		t.Fatalf("resolved model = %#v", model)
	}
}

func TestDependencyResolverRequiresPointerWhenResolverReturnsData(t *testing.T) {
	registry := NewRegistry()
	if err := registry.Register("route", func(context.Context, any, string) types.Message {
		return types.Message{Status: "success", Code: 200, Data: map[string]any{"name": "Ha Long route"}}
	}); err != nil {
		t.Fatal(err)
	}

	err := NewDependencyResolver(registry).Resolve(context.Background(), resolvedTestModel{RouteID: "route-1"}, "")
	if !errors.Is(err, ErrInvalidResolvedData) {
		t.Fatalf("error = %v, want ErrInvalidResolvedData", err)
	}
}

func TestExtractReferencesReadsTarget(t *testing.T) {
	refs, err := ExtractReferences(resolvedTestModel{RouteID: "route-1"})
	if err != nil {
		t.Fatal(err)
	}
	if len(refs) != 1 || refs[0].Target != "Route" {
		t.Fatalf("references = %#v", refs)
	}
}

func TestExtractReferencesRejectsEmptyTarget(t *testing.T) {
	type invalidModel struct {
		RouteID string `ref:"route,into="`
	}
	if _, err := ExtractReferences(invalidModel{RouteID: "route-1"}); err == nil {
		t.Fatal("ExtractReferences() unexpectedly succeeded")
	}
}

func TestExtractReferencesRejectsMissingTarget(t *testing.T) {
	type invalidTargetModel struct {
		RouteID string `ref:"route,into=Route"`
	}
	if _, err := ExtractReferences(invalidTargetModel{RouteID: "route-1"}); !errors.Is(err, ErrReferenceTargetNotFound) {
		t.Fatalf("error = %v, want ErrReferenceTargetNotFound", err)
	}
}

func TestDependencyResolverRejectsMissingTargetBeforeResolving(t *testing.T) {
	type invalidTargetModel struct {
		RouteID string `ref:"route,required,into=Route"`
	}

	registry := NewRegistry()
	called := false
	if err := registry.Register("route", func(context.Context, any, string) types.Message {
		called = true
		return types.Message{Status: "success", Code: 200}
	}); err != nil {
		t.Fatal(err)
	}

	err := NewDependencyResolver(registry).Resolve(context.Background(), &invalidTargetModel{RouteID: "route-1"}, "tenant-a")
	if !errors.Is(err, ErrReferenceTargetNotFound) {
		t.Fatalf("error = %v, want ErrReferenceTargetNotFound", err)
	}
	if called {
		t.Fatal("resolver was called before target validation")
	}
}

func TestDependencyResolverErrors(t *testing.T) {
	registry := NewRegistry()
	if err := registry.Register("route", func(context.Context, any, string) types.Message {
		return types.Message{Status: "error", Code: 404, Message: "not found", MessageErr: ErrReferenceNotFound}
	}); err != nil {
		t.Fatal(err)
	}
	validator := NewDependencyResolver(registry)

	err := validator.Resolve(context.Background(), testModel{}, "")
	if !errors.Is(err, ErrReferenceRequired) {
		t.Fatalf("error = %v, want required error", err)
	}
	err = validator.Resolve(context.Background(), testModel{RouteID: "missing"}, "")
	if !errors.Is(err, ErrReferenceNotFound) {
		t.Fatalf("error = %v, want not-found error", err)
	}

	if err := registry.Register("route", func(context.Context, any, string) types.Message {
		return types.Message{Status: "success", Code: 200}
	}); err != nil {
		t.Fatal(err)
	}
	err = validator.Resolve(context.Background(), testModel{RouteID: "present", CruiseID: "remote"}, "")
	if !errors.Is(err, ErrResolverNotFound) {
		t.Fatalf("error = %v, want resolver-not-found error", err)
	}
}

// BenchmarkDependencyResolverResolve measures the complete per-model flow:
// extract ref tags, invoke three prefix-aware resolvers, and apply each result
// to its into target on the same model pointer.
func BenchmarkDependencyResolverResolve(b *testing.B) {
	registry := NewRegistry()
	for _, name := range []string{"route", "cruise", "site"} {
		name := name
		if err := registry.Register(name, func(_ context.Context, data any, prefix string) types.Message {
			model := data.(*benchmarkDependencyModel)
			if prefix != "tenant-a" {
				return types.Message{Status: "error", Code: 404, Message: "unknown tenant"}
			}
			var id string
			switch name {
			case "route":
				id = model.RouteID
			case "cruise":
				id = model.CruiseID
			case "site":
				id = model.SiteID
			}
			return types.Message{Status: "success", Code: 200, Data: resolvedTestRoute{ID: id, Name: name}}
		}); err != nil {
			b.Fatal(err)
		}
	}

	dependencyResolver := NewDependencyResolver(registry)
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		model := &benchmarkDependencyModel{RouteID: "route-1", CruiseID: "cruise-1", SiteID: "site-1"}
		if err := dependencyResolver.Resolve(ctx, model, "tenant-a"); err != nil {
			b.Fatal(err)
		}
		if model.Route.ID == "" || model.Cruise.ID == "" || model.Site.ID == "" {
			b.Fatal("dependency data was not applied")
		}
	}
}
