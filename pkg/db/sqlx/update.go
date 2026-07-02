package sqlx

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/stingwolf1080/dynamic-filter/pkg/filter"
)

func (c *Conn) Update(filters filter.FilterOptions, data any, table string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// In SQLite, an UPDATE statement uses SET col1 = ?, col2 = ?
	setClause, setArgs, err := c.buildUpdateSetClause(data)
	if err != nil {
		return err
	}

	where, whereArgs := convertFilter(filters)

	query := fmt.Sprintf("UPDATE %s SET %s %s", table, setClause, where)

	args := append(setArgs, whereArgs...)

	_, err = c.DB.ExecContext(ctx, query, args...)
	return err
}

func (c *Conn) UpdateMany(filters filter.FilterOptions, data any, table string) error {
	// UpdateMany behaves identically to Update in SQL context, as the WHERE clause dictates how many rows are matched.
	return c.Update(filters, data, table)
}

func (c *Conn) buildUpdateSetClause(data any) (string, []any, error) {
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() == reflect.Map {
		// Handle map[string]any
		mapVal, ok := data.(map[string]any)
		if !ok {
			return "", nil, fmt.Errorf("expected map[string]any")
		}

		var clauses []string
		var args []any
		for k, v := range mapVal {
			clauses = append(clauses, fmt.Sprintf("%s = ?", k))
			args = append(args, v)
		}
		return strings.Join(clauses, ", "), args, nil
	}

	if val.Kind() != reflect.Struct {
		return "", nil, fmt.Errorf("data must be a struct or map[string]any, got %v", val.Kind())
	}

	typ := val.Type()
	var clauses []string
	var args []any

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" {
			continue // skip unexported
		}

		dbTag := field.Tag.Get("db")
		if dbTag == "-" {
			continue
		}
		if dbTag == "" {
			dbTag = strings.ToLower(field.Name)
		}

		// Optionally skip empty values if needed, but standard SQL updates all specified fields.
		clauses = append(clauses, fmt.Sprintf("%s = ?", dbTag))
		args = append(args, val.Field(i).Interface())
	}

	return strings.Join(clauses, ", "), args, nil
}
