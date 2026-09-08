# Dynamic Filter API syntax

This package provides a powerful, URL-based query parser that translates complex query strings into dynamic database conditions. It is fully integrated with our generic repository and currently supports `mongox`, `pgx`, and `sqlx`.

For ID reference checking and dependency hydration before persistence, see the [Dependency Resolver guide](pkg/resolver/README.md).

## Table of Contents
- [Filtering (`filter`)](#filtering-filter)
  - [Supported Operators](#supported-operators)
  - [AND Conditions](#and-conditions)
  - [OR Conditions](#or-conditions)
  - [IN / NOT IN Operators](#in--not-in-operators)
- [Sorting (`sort`)](#sorting-sort)
- [Pagination (`page`)](#pagination-page)
- [Field Projection (`fields`)](#field-projection-fields)
- [Full-text Search (`search`)](#full-text-search-search)
- [Grouping (`group_by`)](#grouping-group_by)
- [Example Usage](#example-usage)
- [Dependency Resolver](#dependency-resolver)
  - [1. Resolve trực tiếp](#1-resolve-trực-tiếp)
  - [2. Resolve tự động trước khi save](#2-resolve-tự-động-trước-khi-save)
- [Benchmarks](#benchmarks)

---

## Filtering (`filter`)

Use the `filter[field_name]` parameter to filter results.

**Syntax**: `filter[field_name]=[operator]:value`

### Supported Operators

- `[eq]`: Equal
- `[ne]`: Not Equal
- `[gt]`: Greater Than
- `[gte]`: Greater Than or Equal
- `[lt]`: Less Than
- `[lte]`: Less Than or Equal
- `[in]`: In Array/List
- `[nin]`: Not In Array/List

**Example**:
- Get users older than 25:
  `filter[age]=[gt]:25`
- Get users with exactly the name "Alice":
  `filter[name]=[eq]:Alice`

### AND Conditions
To combine multiple conditions with **AND**, simply chain them together using the `&` symbol.

**Example**:
- Age is greater than or equal to 24 **AND** status is active:
  `filter[age]=[gte]:24&filter[status]=[eq]:active`

### OR Conditions
To use **OR** conditions, you provide multiple fields separated by a comma inside the bracket `[]`, and provide multiple operator-value pairs separated by commas.

**Syntax**: `filter[field1,field2]=[op1]:val1,[op2]:val2`

**Examples**:
- Name is "Alice Bob" **OR** Name is "Dave" (same field):
  `filter[name,name]=[eq]:Alice Bob,[eq]:Dave`
- Email is "test@example.com" **OR** username is "test":
  `filter[email,username]=[eq]:test@example.com,[eq]:test`

*(Note: The number of fields provided in the bracket must exactly match the number of operator-value pairs provided!)*

### IN / NOT IN Operators
For `[in]` and `[nin]` operators, provide the values separated by commas.

**Example**:
- Get users in specific departments:
  `filter[department]=[in]:HR,Engineering,Sales`

---

## Sorting (`sort`)

Use `sort[field_name]` to order your results. Supported values are `asc` and `desc`.

**Example**:
- Sort by age descending:
  `sort[age]=desc`
- Sort by created date ascending, then age descending:
  `sort[created_at]=asc&sort[age]=desc`

---

## Pagination (`page`)

Control the size and offset of the result set using `page[size]` and `page[number]`.

**Syntax**:
- `page[size]=N`: Limit the number of records returned (Max 100,000).
- `page[number]=N`: The page number (1-indexed).

**Example**:
- Get page 2 with 50 items per page:
  `page[size]=50&page[number]=2`

For cursor-based pagination (useful for MongoDB), you can use `page[next]=<id>`:
- `page[next]=60f3b...`

---

## Field Projection (`fields`)

Limit the fields returned by the database.

**Syntax**: `fields=field1,field2,field3`

**Example**:
- Only fetch `id`, `name`, and `email`:
  `fields=id,name,email`

---

## Full-text Search (`search`)

Use the `search` parameter to perform full-text search across indexed fields.

**Syntax**: `search=query string`

**Example**:
- Search for "Alice":
  `search=Alice`

---

## Grouping (`group_by`)

Group results by specific fields (limited to a maximum of 3 fields).

**Syntax**: `group_by=field1,field2`

**Example**:
- Group by department and role:
  `group_by=department,role`

---

## Example Usage

Here is a full example combining multiple features to query a dynamic Generic Repository:

```go
package main

import (
	"fmt"
	"log"
	
	"github.com/stingwolf1080/dynamic-filter/repository"
)

func main() {
	// Query string combining filters, sorting, and pagination:
	// 1. Age >= 24
	// 2. AND (Name == "Alice Bob" OR Name == "Dave")
	// 3. Sort by age descending
	// 4. Limit to 10 results
	
	filterStr := "filter[age]=[gte]:24&filter[name,name]=[eq]:Alice Bob,[eq]:Dave&sort[age]=desc&page[size]=10"
	
	readMsg := userRepo.GetByFilter(filterStr)
	if readMsg.Status == "error" {
		log.Fatalf("Error querying users: %v", readMsg.Message)
	}

	users, ok := readMsg.Data.(*[]User)
	if ok {
		for _, u := range *users {
			fmt.Printf("Result: %+v\n", u)
		}
	}
}
```

## Dependency Resolver

Dependency Resolver kiểm tra ID được khai báo bằng tag `ref` và có thể gắn data trả về vào chính struct hiện tại trước khi lưu DB. Hướng dẫn chi tiết hơn ở [pkg/resolver/README.md](pkg/resolver/README.md).

Khai báo model một lần:

```go
type Route struct {
    ID   string `json:"id" bson:"id"`
    Name string `json:"name" bson:"name"`
}

type Schedule struct {
    RouteID string `json:"route_id" bson:"route_id" ref:"route,required,into=Route"`
    Route   Route  `json:"route,omitempty" bson:"route,omitempty"`
}
```

Đăng ký resolver tại application startup. `prefix` được dùng trong phase check/query để lấy đúng tenant/table/collection:

```go
registry := resolver.NewRegistry()

must(registry.Register("route", func(
    ctx context.Context,
    data any,
    prefix string,
) types.Message {
    schedule := data.(*Schedule)

    route, err := routeRepo.FindByID(ctx, schedule.RouteID, prefix)
    if err != nil {
        return types.Message{
            Status:     "error",
            Code:       404,
            Message:    "route not found",
            MessageErr: err,
        }
    }
    return types.Message{Status: "success", Code: 200, Data: route}
}))

dependencyResolver := resolver.NewDependencyResolver(registry)
```

### 1. Resolve trực tiếp

Dùng khi service tự kiểm tra và hydrate model trước khi thực hiện logic/save:

```go
schedule := &Schedule{RouteID: "route-1"}

if err := dependencyResolver.Resolve(ctx, schedule, currentPrefix); err != nil {
    return err
}

// Phase 1: route resolver check/query RouteID bằng currentPrefix.
// Phase 2: Message.Data được gán trực tiếp vào schedule.Route.
// schedule đã sẵn sàng để save.
```

### 2. Resolve tự động trước khi save

Dùng khi muốn repository tự resolve trong `BeforeInsert`. Chỉ đăng ký hook một lần khi khởi tạo application:

```go
hooks.RegisterDependencyResolver[Schedule](dependencyResolver)
```

Sau đó gọi repository bình thường:

```go
message := scheduleRepository.Create(Schedule{RouteID: "route-1"})
if message.HasError() {
    return message.MessageErr
}
```

Flow tự động:

```text
repository.Create(data)
→ BeforeInsert
→ DependencyResolver.Resolve(ctx, &data, repository prefix)
→ route resolver check/query RouteID
→ gán Route data vào data.Route
→ save data vào DB
```

Nếu dùng `into=Route` nhưng struct không khai báo field exported `Route`, lỗi `ErrReferenceTargetNotFound` được trả về ngay lúc đọc tag, trước khi resolver query DB/remote service.

## Benchmarks

Run every benchmark in the project (without running normal tests):

```bash
go test ./... -run '^$' -bench . -benchmem -vet=off
```

`-vet=off` prevents unrelated vet warnings in example packages from blocking benchmark execution. For more stable comparison, repeat the run:

```bash
go test ./... -run '^$' -bench . -benchmem -vet=off -count=3
```
