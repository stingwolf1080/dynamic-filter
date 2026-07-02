package sqlx

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/stingwolf1080/dynamic-filter/pkg/db"
	"github.com/stingwolf1080/dynamic-filter/pkg/filter"
)

func (c *Conn) Read(filters filter.FilterOptions, table string, result any) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	where, args := convertFilter(filters)
	opts := buildOptions(filters)

	query := fmt.Sprintf("SELECT * FROM %s %s %s LIMIT 1", table, where, opts)

	err := c.DB.GetContext(ctx, result, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return db.ErrNotFound
		}
		return err
	}

	return nil
}

func (c *Conn) ReadMany(filters filter.FilterOptions, table string, results any) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	where, args := convertFilter(filters)
	opts := buildOptions(filters)

	query := fmt.Sprintf("SELECT * FROM %s %s %s", table, where, opts)

	err := c.DB.SelectContext(ctx, results, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (c *Conn) Count(filters filter.FilterOptions, table string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	where, args := convertFilter(filters)
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", table, where)

	var count int64
	err := c.DB.GetContext(ctx, &count, query, args...)
	return err
}
