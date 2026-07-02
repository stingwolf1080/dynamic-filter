package sqlx

import (
	"context"
	"fmt"
	"time"

	"github.com/stingwolf1080/dynamic-filter/pkg/filter"
)

func (c *Conn) Delete(filters filter.FilterOptions, table string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	where, args := convertFilter(filters)

	query := fmt.Sprintf("DELETE FROM %s %s", table, where)
	// Optionally we could add LIMIT 1 for safety if the interface expects deleting a single row,
	// but SQLite requires compiling with SQLITE_ENABLE_UPDATE_DELETE_LIMIT to support LIMIT on DELETE.
	// We will trust the filter logic.

	_, err := c.DB.ExecContext(ctx, query, args...)
	return err
}

func (c *Conn) DeleteMany(filters filter.FilterOptions, table string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	where, args := convertFilter(filters)

	query := fmt.Sprintf("DELETE FROM %s %s", table, where)

	_, err := c.DB.ExecContext(ctx, query, args...)
	return err
}
