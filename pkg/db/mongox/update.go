package mongox

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (conn *Conn) Update(filter bson.M, data any, table string, opts *options.UpdateOptions) error {
	var err error
	dataCollection := conn.Database.Collection(table)
	ctx, _ := context.WithTimeout(context.Background(), 60*time.Second)
	_, err = dataCollection.UpdateOne(ctx, filter, data)
	if err != nil {
		return err
	}
	return nil
}

func (conn *Conn) UpdateMany(filter bson.M, data any, table string, opts *options.UpdateOptions) error {
	var err error
	dataCollection := conn.Database.Collection(table)
	ctx, _ := context.WithTimeout(context.Background(), 60*time.Second)
	_, err = dataCollection.UpdateMany(ctx, filter, data, opts)
	if err != nil {
		return err
	}
	return nil
}
