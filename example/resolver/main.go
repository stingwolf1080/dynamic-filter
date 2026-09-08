package main

import (
	"context"
	"fmt"

	"github.com/stingwolf1080/dynamic-filter/pkg/resolver"
	"github.com/stingwolf1080/dynamic-filter/pkg/util/types"
)

// Schedule contains transport-neutral reference names. The resolver functions
// decide how each reference is checked (repository, gRPC, HTTP, and so on).
type Schedule struct {
	RouteID  string `ref:"route,required,into=Route"`
	Route    Route  `json:"route,omitempty"`
	CruiseID string `ref:"cruise,required"`
}

type Route struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func main() {
	registry := resolver.NewRegistry()

	// In a real application this function can call a repository. data carries
	// the model and prefix identifies the current tenant/collection for phase 1.
	must(registry.Register("route", func(_ context.Context, data any, prefix string) types.Message {
		schedule := data.(*Schedule)
		if prefix == "tenant-a" && schedule.RouteID == "route-1" {
			return types.Message{Status: "success", Code: 200, Data: Route{ID: "route-1", Name: "Ha Long"}}
		}
		return types.Message{Status: "error", Code: 404, Message: "route not found"}
	}))

	// This function could instead call a gRPC or HTTP client.
	must(registry.Register("cruise", func(_ context.Context, data any, _ string) types.Message {
		schedule := data.(*Schedule)
		if schedule.CruiseID == "cruise-1" {
			return types.Message{Status: "success", Code: 200}
		}
		return types.Message{Status: "error", Code: 404, Message: "cruise not found"}
	}))

	dependencyResolver := resolver.NewDependencyResolver(registry)
	schedule := &Schedule{RouteID: "route-1", CruiseID: "cruise-1"}
	fmt.Printf("before dependency resolution: %#v\n", schedule)

	// Resolve is the public dependency-resolution entry point. Internally it:
	//   1. extracts RouteID's ref:"route,...,into=Route" tag;
	//   2. calls the "route" resolver with this same *Schedule;
	//   3. creates a Dependency for schedule and applies Message.Data to Route.
	if err := dependencyResolver.Resolve(context.Background(), schedule, "tenant-a"); err != nil {
		panic(err)
	}
	fmt.Printf("after dependency resolution; model ready to save: %#v\n", schedule)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
