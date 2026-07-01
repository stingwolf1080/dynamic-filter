package pgx

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

	cols, vals := extractFieldsAndValues(model)
	if len(cols) == 0 {
		return fmt.Errorf("no fields to insert")
	}

	placeholders := make([]string, len(cols))
	for i := range cols {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(cols, ", "), strings.Join(placeholders, ", "))
	
	_, err := c.Pool.Exec(ctx, query, vals...)
	return err
}

func (c *Conn) CreateMany(models []any, table string) error {
	if len(models) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cols, _ := extractFieldsAndValues(models[0])
	if len(cols) == 0 {
		return fmt.Errorf("no fields to insert")
	}

	var allVals []any
	var allPlaceholders []string
	argID := 1

	for _, model := range models {
		_, vals := extractFieldsAndValues(model)
		placeholders := make([]string, len(cols))
		for i := range cols {
			placeholders[i] = fmt.Sprintf("$%d", argID)
			argID++
		}
		allPlaceholders = append(allPlaceholders, fmt.Sprintf("(%s)", strings.Join(placeholders, ", ")))
		allVals = append(allVals, vals...)
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", table, strings.Join(cols, ", "), strings.Join(allPlaceholders, ", "))

	_, err := c.Pool.Exec(ctx, query, allVals...)
	return err
}

func extractFieldsAndValues(model any) ([]string, []any) {
	v := reflect.ValueOf(model)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	var cols []string
	var vals []any

	if v.Kind() == reflect.Struct {
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			tag := field.Tag.Get("db")
			if tag == "" {
				tag = field.Tag.Get("json")
			}
			if tag == "" || tag == "-" {
				continue
			}
			// handle json omit
			tag = strings.Split(tag, ",")[0]
			cols = append(cols, tag)
			vals = append(vals, v.Field(i).Interface())
		}
	} else if v.Kind() == reflect.Map {
		for _, key := range v.MapKeys() {
			cols = append(cols, key.String())
			vals = append(vals, v.MapIndex(key).Interface())
		}
	}

	return cols, vals
}
