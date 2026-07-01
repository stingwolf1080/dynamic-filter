package mongox

import (
	"context"
	"time"

	"github.com/stingwolf1080/dynamic-filter/pkg/db"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Conn struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func (c *Conn) Connect(opts db.Options) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(opts.URI))
	if err != nil {
		return err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return err
	}

	c.Client = client
	c.Database = client.Database(opts.Database)

	return nil
}

func (c *Conn) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return c.Client.Disconnect(ctx)
}
