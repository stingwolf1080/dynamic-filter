package pgx

import (
	"context"
	"fmt"
	"time"

	"github.com/stingwolf1080/dynamic-filter/pkg/filter"
)

func (c *Conn) Delete(filters filter.FilterOptions, table string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	where, args := convertFilter(filters, 1)
	
	// Adding LIMIT 1 since it's Delete (one)
	// Postgres uses ctid for limiting deletes
	query := fmt.Sprintf("DELETE FROM %s WHERE ctid IN (SELECT ctid FROM %s %s LIMIT 1)", table, table, where)
	
	_, err := c.Pool.Exec(ctx, query, args...)
	return err
}

func (c *Conn) DeleteMany(filters filter.FilterOptions, table string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	where, args := convertFilter(filters, 1)

	query := fmt.Sprintf("DELETE FROM %s %s", table, where)
	
	_, err := c.Pool.Exec(ctx, query, args...)
	return err
}
