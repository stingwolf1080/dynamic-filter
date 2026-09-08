package hooks

import (
	"context"
	"testing"

	"github.com/stingwolf1080/dynamic-filter/pkg/helper"
	"github.com/stingwolf1080/dynamic-filter/pkg/resolver"
	"github.com/stingwolf1080/dynamic-filter/pkg/util/types"
)

type referenceHookModel struct {
	RouteID string             `ref:"route,required,into=Route"`
	Route   referenceHookRoute `json:"route,omitempty"`
}

type referenceHookRoute struct {
	Name string `json:"name"`
}

func TestRegisterDependencyResolver(t *testing.T) {
	registry := resolver.NewRegistry()
	if err := registry.Register("route", func(context.Context, any, string) types.Message {
		return types.Message{Status: "success", Code: 200, Data: map[string]any{"name": "Ha Long"}}
	}); err != nil {
		t.Fatal(err)
	}
	RegisterDependencyResolver[referenceHookModel](resolver.NewDependencyResolver(registry))

	model := &referenceHookModel{RouteID: "route-1"}
	message := Run(context.Background(), helper.GetNameModel[referenceHookModel](), "", BeforeInsert, model)
	if message.HasError() {
		t.Fatalf("valid reference hook failed: %v", message.MessageErr)
	}
	if model.Route.Name != "Ha Long" {
		t.Fatalf("resolved hook model = %#v", model)
	}
	message = Run(context.Background(), helper.GetNameModel[referenceHookModel](), "", BeforeInsert, &referenceHookModel{})
	if !message.HasError() {
		t.Fatal("missing required reference passed validation")
	}
}
