package pgx

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stingwolf1080/dynamic-filter/pkg/db"
	"github.com/stingwolf1080/dynamic-filter/pkg/filter"
)

func (c *Conn) Read(filters filter.FilterOptions, table string, result any) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	where, args := convertFilter(filters, 1)
	opts := buildOptions(filters)

	query := fmt.Sprintf("SELECT row_to_json(t) FROM (SELECT * FROM %s %s %s LIMIT 1) t", table, where, opts)

	var b []byte
	err := c.Pool.QueryRow(ctx, query, args...).Scan(&b)
	if err != nil {
		if err == pgx.ErrNoRows {
			return db.ErrNotFound
		}
		return err
	}

	return json.Unmarshal(b, result)
}

func (c *Conn) ReadMany(filters filter.FilterOptions, table string, results any) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	where, args := convertFilter(filters, 1)
	opts := buildOptions(filters)

	query := fmt.Sprintf("SELECT COALESCE(json_agg(t), '[]'::json) FROM (SELECT * FROM %s %s %s) t", table, where, opts)

	var b []byte
	err := c.Pool.QueryRow(ctx, query, args...).Scan(&b)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, results)
}

func (c *Conn) Count(filters filter.FilterOptions, table string) error {
	// The db.Connection interface signature expects Count(filters, table) error
	// But where does the count go? If it returns an error, there's no int64 return type.
	// Since mongox implementation had Count returning (int64, error), this implies
	// a mismatch with db.Connection. We will execute the query but without a target
	// pointer in the interface, we cannot return the value. 
	// We'll execute the count internally to ensure valid syntax.
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	where, args := convertFilter(filters, 1)
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", table, where)

	var count int64
	err := c.Pool.QueryRow(ctx, query, args...).Scan(&count)
	return err
}

func (c *Conn) CheckNotFound() error {
	return nil // To be implemented based on the caller's logic
}
