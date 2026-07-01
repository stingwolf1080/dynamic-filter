package pgx

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/stingwolf1080/dynamic-filter/pkg/filter"
)

func (c *Conn) Update(filters filter.FilterOptions, data any, table string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cols, vals := extractFieldsAndValues(data)
	if len(cols) == 0 {
		return fmt.Errorf("no fields to update")
	}

	var setConds []string
	argID := 1
	for _, col := range cols {
		setConds = append(setConds, fmt.Sprintf("%s = $%d", col, argID))
		argID++
	}

	where, whereArgs := convertFilter(filters, argID)
	
	// combine args
	vals = append(vals, whereArgs...)

	query := fmt.Sprintf("UPDATE %s SET %s %s", table, strings.Join(setConds, ", "), where)

	_, err := c.Pool.Exec(ctx, query, vals...)
	return err
}

func (c *Conn) UpdateMany(filters filter.FilterOptions, data any, table string) error {
	// For SQL, Update and UpdateMany are technically the same since WHERE matches multiple.
	return c.Update(filters, data, table)
}
