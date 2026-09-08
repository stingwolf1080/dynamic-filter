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

## Benchmarks

Run every benchmark in the project (without running normal tests):

```bash
go test ./... -run '^$' -bench . -benchmem -vet=off
```

`-vet=off` prevents unrelated vet warnings in example packages from blocking benchmark execution. For more stable comparison, repeat the run:

```bash
go test ./... -run '^$' -bench . -benchmem -vet=off -count=3
```
