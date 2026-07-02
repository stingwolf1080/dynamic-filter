package sqlx

import (
	"fmt"
	"strings"

	"github.com/stingwolf1080/dynamic-filter/pkg/filter"
)

func convertFilter(filters filter.FilterOptions) (string, []any) {
	var conditions []string
	var args []any

	if filters.Search != "" {
		searchCol := filters.CheckAlias("search")
		if searchCol == "search" {
			searchCol = "search_index"
		}
		// For SQLite FTS5 with external content tables, a subquery on rowid is the standard pattern
		conditions = append(conditions, fmt.Sprintf("rowid IN (SELECT rowid FROM %s WHERE %s MATCH ?)", searchCol, searchCol))
		args = append(args, filters.Search)
	}

	for _, v := range filters.Filter {
		var subConds []string
		for _, f := range v.Value {
			switch f.Operator {
			case "eq":
				subConds = append(subConds, fmt.Sprintf("%s = ?", f.Field))
				args = append(args, f.Value)
			case "ne":
				subConds = append(subConds, fmt.Sprintf("%s != ?", f.Field))
				args = append(args, f.Value)
			case "gt":
				subConds = append(subConds, fmt.Sprintf("%s > ?", f.Field))
				args = append(args, f.Value)
			case "gte":
				subConds = append(subConds, fmt.Sprintf("%s >= ?", f.Field))
				args = append(args, f.Value)
			case "lt":
				subConds = append(subConds, fmt.Sprintf("%s < ?", f.Field))
				args = append(args, f.Value)
			case "lte":
				subConds = append(subConds, fmt.Sprintf("%s <= ?", f.Field))
				args = append(args, f.Value)
			case "like":
				subConds = append(subConds, fmt.Sprintf("%s LIKE ?", f.Field))
				args = append(args, "%"+fmt.Sprint(f.Value)+"%")
			case "ilike":
				// SQLite LIKE is case-insensitive by default for ASCII.
				subConds = append(subConds, fmt.Sprintf("%s LIKE ?", f.Field))
				args = append(args, "%"+fmt.Sprint(f.Value)+"%")
			case "in":
				// For IN, we need multiple placeholders
				if vals, ok := f.Value.([]any); ok && len(vals) > 0 {
					placeholders := make([]string, len(vals))
					for i := range vals {
						placeholders[i] = "?"
						args = append(args, vals[i])
					}
					subConds = append(subConds, fmt.Sprintf("%s IN (%s)", f.Field, strings.Join(placeholders, ",")))
				} else {
					// Fallback for single value or comma-separated string if it wasn't parsed as slice
					subConds = append(subConds, fmt.Sprintf("%s IN (?)", f.Field))
					args = append(args, f.Value)
				}
			}
		}

		if len(subConds) > 0 {
			if v.OrOperator {
				conditions = append(conditions, "("+strings.Join(subConds, " OR ")+")")
			} else {
				conditions = append(conditions, "("+strings.Join(subConds, " AND ")+")")
			}
		}
	}

	if len(conditions) == 0 {
		return "", nil
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}

func buildOptions(filters filter.FilterOptions) string {
	var opts []string

	if len(filters.Sort) > 0 {
		var orders []string
		for _, s := range filters.Sort {
			dir := "ASC"
			if s.Index == -1 {
				dir = "DESC"
			}
			orders = append(orders, fmt.Sprintf("%s %s", s.Field, dir))
		}
		opts = append(opts, "ORDER BY "+strings.Join(orders, ", "))
	}

	if filters.Limit() > 0 {
		opts = append(opts, fmt.Sprintf("LIMIT %d", filters.Limit()))
	}

	if filters.Skip() > 0 {
		opts = append(opts, fmt.Sprintf("OFFSET %d", filters.Skip()))
	}

	return strings.Join(opts, " ")
}
