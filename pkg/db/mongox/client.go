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

	clientOptions := options.Client().ApplyURI(opts.URI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return err
	}

	c.Client = client
	if opts.Database != "" {
		c.Database = client.Database(opts.Database)
	}

	return nil
}

func (c *Conn) Close() error {
	if c.Client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return c.Client.Disconnect(ctx)
	}
	return nil
}

func (c *Conn) CheckNotFound() error {
	return mongo.ErrNoDocuments
}
