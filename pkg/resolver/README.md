# Dependency Resolver

`resolver` checks IDs declared with a `ref` tag and, optionally, writes the resolved data into the same model pointer. It is useful when a model stores a reference ID such as `RouteID`, while the create/update flow also needs the corresponding `Route` data.

The resolution flow has two phases:

```text
1. Check/query
   RouteID + prefix → registered "route" resolver → Message.Data

2. Apply
   Message.Data + into=Route → Schedule.Route
```

`prefix` is used only by phase 1, for example to select the current tenant or collection. Phase 2 changes the supplied model pointer directly and does not use `prefix`.

## Model declaration

Use `ref:"resolver-name,options"` on the string ID field.

```go
type Route struct {
    ID   string `json:"id" bson:"id"`
    Name string `json:"name" bson:"name"`
}

type Schedule struct {
    RouteID  string `json:"route_id" bson:"route_id" ref:"route,required,into=Route"`
    Route    Route  `json:"route,omitempty" bson:"route,omitempty"`
    CruiseID string `json:"cruise_id" bson:"cruise_id" ref:"cruise,required"`
}
```

Supported tag forms:

| Tag | Behaviour |
| --- | --- |
| `ref:"route"` | Check a non-empty ID using the `route` resolver. |
| `ref:"route,required"` | The ID must be non-empty, then check it. |
| `ref:"route,into=Route"` | Check the ID, then assign resolver data to `Route`. |
| `ref:"route,required,into=Route"` | Require, check, then assign data. |

`into` is optional. Use it only when resolved data should be kept on the model. The target must be an exported field on the same struct; it can be either a value (`Route`) or pointer (`*Route`) field.

## Register resolvers

Create one registry at application startup and register one resolver for each tag name. A resolver receives the current model and the dynamic `prefix`.

```go
registry := resolver.NewRegistry()

err := registry.Register("route", func(
    ctx context.Context,
    data any,
    prefix string,
) types.Message {
    schedule := data.(*Schedule)

    // Use prefix to query the correct tenant/table/collection.
    route, err := routeRepo.FindByID(ctx, schedule.RouteID, prefix)
    if err != nil {
        return types.Message{
            Status:     "error",
            Code:       404,
            Message:    "route not found",
            MessageErr: err,
        }
    }

    return types.Message{
        Status: "success",
        Code:   200,
        Data:   route,
    }
})
if err != nil {
    panic(err)
}
```

`Message.Data` may be the exact target type (for example `Route` or `*Route`) or a JSON-compatible map that can be decoded into the target field.

## Resolve one model directly

Create a resolver from the registry, then pass the current model pointer and the runtime prefix:

```go
dependencyResolver := resolver.NewDependencyResolver(registry)

schedule := &Schedule{
    RouteID:  "route-1",
    CruiseID: "cruise-1",
}

if err := dependencyResolver.Resolve(ctx, schedule, currentPrefix); err != nil {
    return err
}

// schedule.Route now contains the resolver result.
// Save schedule with the repository/DB as usual.
```

For models that use `into=...`, pass a non-nil pointer to a struct. The resolver needs that pointer to apply `Message.Data` to the target field.

## Use with repository hooks

To resolve dependencies automatically before an insert, register the hook once during application setup:

```go
dependencyResolver := resolver.NewDependencyResolver(registry)
hooks.RegisterDependencyResolver[Schedule](dependencyResolver)
```

The repository hook supplies its current prefix automatically:

```text
repository.Create(data)
    → BeforeInsert hook
    → DependencyResolver.Resolve(ctx, &data, repositoryPrefix)
    → check/query through resolver
    → apply Message.Data to fields declared by into
    → database save
```

## Errors and configuration checks

Configuration is checked while tags are extracted, before any resolver query:

```go
type InvalidSchedule struct {
    RouteID string `ref:"route,into=Route"`
    // Route field is missing.
}
```

Resolving `InvalidSchedule` returns `ErrReferenceTargetNotFound`; the `route` resolver is not called.

Other important errors are:

| Error | Meaning |
| --- | --- |
| `ErrReferenceRequired` | A `required` ID is empty. |
| `ErrResolverNotFound` | No resolver was registered for the tag name. |
| `ErrReferenceTargetNotFound` | `into` points to a missing or unexported field. |
| `ErrInvalidResolvedData` | The model is not an applicable pointer, or data cannot be assigned. |

## Performance

Tag metadata is cached per `reflect.Type`. The cache stores only tag configuration and field indexes; IDs and resolved data are always read from and applied to the current model pointer.

Run the resolver benchmark:

```bash
go test ./pkg/resolver -run '^$' -bench . -benchmem
```
