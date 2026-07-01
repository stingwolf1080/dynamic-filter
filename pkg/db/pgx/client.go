package pgx

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stingwolf1080/dynamic-filter/pkg/db"
)

type Conn struct {
	Pool *pgxpool.Pool
}

func (c *Conn) Connect(opts db.Options) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, opts.URI)
	if err != nil {
		return err
	}

	if err := pool.Ping(ctx); err != nil {
		return err
	}

	c.Pool = pool
	return nil
}

func (c *Conn) Close() error {
	if c.Pool != nil {
		c.Pool.Close()
	}
	return nil
}
