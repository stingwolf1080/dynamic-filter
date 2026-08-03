package sqlx

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"
)

func (c *Conn) Create(model any, table string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use sqlx.NamedExec which dynamically binds struct fields to the query.
	// Since we don't know the fields in advance without reflection,
	// we will rely on sqlx's ability to extract field names automatically.
	// Wait, to do an INSERT dynamically without reflection or knowing the fields:
	// We could use reflection to build the query, or use a helper function.
	// For now, assuming the caller has already set up the model.

	query, args, err := c.buildInsertQuery(model, table)
	if err != nil {
		return err
	}

	_, err = c.DB.ExecContext(ctx, query, args...)
	return err
}

func (c *Conn) CreateMany(models []any, table string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if len(models) == 0 {
		return nil
	}

	tx, err := c.DB.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, model := range models {
		query, args, err := c.buildInsertQuery(model, table)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// buildInsertQuery uses simple reflection to generate an INSERT statement.
func (c *Conn) buildInsertQuery(model any, table string) (string, []any, error) {
	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return "", nil, fmt.Errorf("model must be a struct, got %v", val.Kind())
	}

	typ := val.Type()
	var columns []string
	var placeholders []string
	var args []any

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		// Skip unexported fields
		if field.PkgPath != "" {
			continue
		}

		dbTag := field.Tag.Get("db")
		if dbTag == "-" {
			continue
		}
		if dbTag == "" {
			dbTag = strings.ToLower(field.Name)
		}

		columns = append(columns, dbTag)
		placeholders = append(placeholders, "?")
		args = append(args, val.Field(i).Interface())
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	return query, args, nil
}

func (c *Conn) CreateTextIndex(table string, fields []string) error {
	if len(fields) == 0 {
		return nil
	}
	indexName := fmt.Sprintf("idx_%s_text", table)
	driver := c.DB.DriverName()
	var query string

	if driver == "postgres" || driver == "pgx" || driver == "pq-timeouts" || driver == "pq" {
		var coalesceFields []string
		for _, field := range fields {
			coalesceFields = append(coalesceFields, fmt.Sprintf("coalesce(%s::text, '')", field))
		}
		tsVectorExpr := fmt.Sprintf("to_tsvector('simple', %s)", strings.Join(coalesceFields, " || ' ' || "))
		query = fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s USING GIN (%s)", indexName, table, tsVectorExpr)
	} else {
		query = fmt.Sprintf("CREATE FULLTEXT INDEX %s ON %s (%s)", indexName, table, strings.Join(fields, ", "))
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	_, err := c.DB.ExecContext(ctx, query)
	return err
}
