package sqlx

import (
	"github.com/jmoiron/sqlx"
	"github.com/stingwolf1080/dynamic-filter/pkg/db"
	_ "modernc.org/sqlite"
)

type Conn struct {
	DB *sqlx.DB
}

func (c *Conn) Connect(opts db.Options) error {
	dbConn, err := sqlx.Connect("sqlite", opts.URI)
	if err != nil {
		return err
	}

	c.DB = dbConn
	return nil
}

func (c *Conn) Close() error {
	if c.DB != nil {
		return c.DB.Close()
	}
	return nil
}

func (c *Conn) CheckNotFound() error {
	return db.ErrNotFound
}
