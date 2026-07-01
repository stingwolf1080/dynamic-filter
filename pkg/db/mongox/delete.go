package mongox

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (conn *Conn) DeleteOne(filter bson.M, table string, opts *options.DeleteOptions) error {
	var err error
	dataCollection := conn.Database.Collection(table)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err = dataCollection.DeleteOne(ctx, filter, opts)
	if err != nil {
		return err
	}
	return nil
}

func (conn *Conn) DeleteMany(filter bson.M, table string, opts *options.DeleteOptions) error {
	var err error
	dataCollection := conn.Database.Collection(table)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	_, err = dataCollection.DeleteMany(ctx, filter, opts)
	if err != nil {
		return err
	}
	return nil
}
