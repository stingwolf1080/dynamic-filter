package pgx

import (
	"fmt"
	"strings"

	"github.com/stingwolf1080/dynamic-filter/pkg/filter"
)

func convertFilter(filters filter.FilterOptions, startArgID int) (string, []any) {
	var conditions []string
	var args []any
	argID := startArgID

	if filters.Search != "" {
		searchCol := filters.CheckAlias("search")
		if searchCol == "search" {
			searchCol = "search_vector"
		}
		conditions = append(conditions, fmt.Sprintf("%s @@ websearch_to_tsquery('english', $%d)", searchCol, argID))
		args = append(args, filters.Search)
		argID++
	}

	for _, v := range filters.Filter {
		if len(v.Value) == 0 {
			continue
		}

		if v.OrOperator {
			var orConds []string
			for _, fv := range v.Value {
				field := filters.CheckAlias(fv.Field)
				op := mapOperator(fv.Operator)
				val := fv.ConvertValue(filters.Timezone())
				
				if op == "IN" || op == "NOT IN" {
					orConds = append(orConds, fmt.Sprintf("%s %s (ANY($%d))", field, op, argID))
				} else {
					orConds = append(orConds, fmt.Sprintf("%s %s $%d", field, op, argID))
				}
				args = append(args, val)
				argID++
			}
			conditions = append(conditions, fmt.Sprintf("(%s)", strings.Join(orConds, " OR ")))
			continue
		}

		if len(v.Value) > 1 {
			var andConds []string
			for _, fv := range v.Value {
				field := filters.CheckAlias(fv.Field)
				op := mapOperator(fv.Operator)
				val := fv.ConvertValue(filters.Timezone())
				
				if op == "IN" || op == "NOT IN" {
					andConds = append(andConds, fmt.Sprintf("%s %s (ANY($%d))", field, op, argID))
				} else {
					andConds = append(andConds, fmt.Sprintf("%s %s $%d", field, op, argID))
				}
				args = append(args, val)
				argID++
			}
			conditions = append(conditions, fmt.Sprintf("(%s)", strings.Join(andConds, " AND ")))
			continue
		}

		value := v.Value[0]
		field := filters.CheckAlias(value.Field)
		op := mapOperator(value.Operator)
		val := value.ConvertValue(filters.Timezone())

		if op == "IN" || op == "NOT IN" {
			conditions = append(conditions, fmt.Sprintf("%s %s (ANY($%d))", field, op, argID))
		} else {
			conditions = append(conditions, fmt.Sprintf("%s %s $%d", field, op, argID))
		}
		args = append(args, val)
		argID++
	}

	if len(conditions) == 0 {
		return "", args
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}

func mapOperator(op string) string {
	switch op {
	case "eq", "$eq":
		return "="
	case "ne", "$ne":
		return "<>"
	case "gt", "$gt":
		return ">"
	case "gte", "$gte":
		return ">="
	case "lt", "$lt":
		return "<"
	case "lte", "$lte":
		return "<="
	case "in", "$in":
		return "IN"
	case "nin", "$nin":
		return "NOT IN"
	case "like", "$regex":
		return "ILIKE"
	default:
		return "="
	}
}

func buildOptions(filters filter.FilterOptions) string {
	var opts []string
	if len(filters.Sort) > 0 {
		var sorts []string
		for _, s := range filters.Sort {
			dir := "ASC"
			if s.Index == -1 {
				dir = "DESC"
			}
			sorts = append(sorts, fmt.Sprintf("%s %s", filters.CheckAlias(s.Field), dir))
		}
		opts = append(opts, "ORDER BY "+strings.Join(sorts, ", "))
	}
	if filters.Pagination() {
		opts = append(opts, fmt.Sprintf("LIMIT %d", filters.Limit()))
		opts = append(opts, fmt.Sprintf("OFFSET %d", filters.Skip()))
	}
	return strings.Join(opts, " ")
}
